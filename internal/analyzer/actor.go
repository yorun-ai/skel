package analyzer

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

func parseActor(reporter *_DiagnosticReporter, ga *grammar.Actor) (*schema.Actor, bool) {
	valid := checkCaseAdvanced(reporter, "Actor", "", "Actor", caseTypeCamel, ga.Name)
	meta, metaValid := parseDecoratorMeta(reporter, ga.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s actor does not support decorator @example", ga.Name.Pos) && valid
	for _, section := range ga.Sections {
		if section.Via != nil {
			reporter.warnIgnoredVia(section.Via.Pos)
		}
	}
	auth, authValid := parseActorAuth(reporter, ga)
	valid = authValid && valid
	permission, permissionValid := parseActorPermission(reporter, ga)
	valid = permissionValid && valid
	return &schema.Actor{
		Pos:              position(ga.Name.Pos),
		Name:             ga.Name.Value,
		SkelName:         "",
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Pub:              ga.Pub,
		Auth:             auth,
		Permission:       permission,
	}, valid
}

func parseActorAuth(reporter *_DiagnosticReporter, ga *grammar.Actor) (*schema.ActorAuth, bool) {
	authSection, valid := actorAuthSection(reporter, ga)
	if authSection == nil {
		return nil, valid
	}
	credential, credentialValid := parseActorCredential(reporter, ga, authSection)
	info, identifierField, infoValid := parseActorInfo(reporter, ga, authSection)
	var impersonators []*schema.ActorImpersonator
	seen := map[string]bool{}
	if authSection.Impersonator != nil {
		for _, ref := range authSection.Impersonator.Actors {
			if reporter.cancelled() {
				break
			}
			name := ref.String()
			valid = reporter.checkNotDuplicate(seen[name], "%s duplicated impersonator %s", ref.Pos, name) && valid
			seen[name] = true
			impersonators = append(impersonators, new(schema.ActorImpersonator{
				Pos:   position(ref.Pos),
				Actor: name,
			}))
		}
	}
	return new(schema.ActorAuth{
		Pos:             position(authSection.Pos),
		Credential:      credential,
		Info:            info,
		IdentifierField: identifierField,
		Impersonators:   impersonators,
	}), credentialValid && infoValid && valid
}

func parseActorCredential(reporter *_DiagnosticReporter, ga *grammar.Actor, authSection *grammar.ActorAuth) (*schema.Data, bool) {
	credentialSection := authSection.Credential
	meta, metaValid := parseDecoratorMeta(reporter, credentialSection.Decorators, _DecoratorContext{
		allowSensitive: true,
	})
	name := &grammar.Identifier{
		Pos:   credentialSection.Pos,
		Value: ga.Name.Value + "Credential",
	}
	credential, valid := parseDataLike(reporter, &grammar.Data{
		Pos:     credentialSection.Pos,
		Pub:     ga.Pub,
		Name:    name,
		Members: credentialSection.Members,
	}, schema.DataKindData)
	valid = metaValid && valid
	credential.Sensitive = meta.Sensitive
	valid = reporter.check(len(credential.Members) > 0, "%s actor credential must have at least one member", credentialSection.Pos) && valid
	hasRequiredField := false
	for _, member := range credential.Members {
		if reporter.cancelled() {
			break
		}
		valid = reporter.check(member.Type.Kind == schema.TypeKindScalar && member.Type.Scalar == schema.ScalarString,
			"%s actor credential member %s must be string or string?", member.Pos, member.Name) && valid
		if !member.Type.Nullable {
			hasRequiredField = true
		}
	}
	if valid {
		valid = reporter.check(hasRequiredField, "%s actor credential must have at least one required string member", credentialSection.Pos)
	}
	credential.Pub = ga.Pub
	return credential, valid
}

func parseActorInfo(reporter *_DiagnosticReporter, ga *grammar.Actor, authSection *grammar.ActorAuth) (*schema.Data, string, bool) {
	infoSection := authSection.Info
	meta, metaValid := parseDecoratorMeta(reporter, infoSection.Decorators, _DecoratorContext{
		allowSensitive: true,
	})
	name := &grammar.Identifier{
		Pos:   infoSection.Pos,
		Value: ga.Name.Value + "Info",
	}
	members := make([]*grammar.DataMember, 0, len(infoSection.Members))
	identifierField := ""
	for _, source := range infoSection.Members {
		if reporter.cancelled() {
			break
		}
		member := *source
		member.Decorators = nil
		for _, decorator := range source.Decorators {
			if reporter.cancelled() {
				break
			}
			if decorator.Name.Value != "identifier" {
				member.Decorators = append(member.Decorators, decorator)
				continue
			}
			metaValid = reporter.check(identifierField == "", "%s actor info supports only one @identifier", decorator.Name.Pos) && metaValid
			metaValid = reporter.check(decorator.Value == nil, "%s decorator @identifier does not accept an argument", decorator.Name.Pos) && metaValid
			identifierField = member.Name.Value
		}
		members = append(members, &member)
	}
	info, valid := parseDataLike(reporter, &grammar.Data{
		Pos:     infoSection.Pos,
		Pub:     ga.Pub,
		Name:    name,
		Members: members,
	}, schema.DataKindData)
	valid = metaValid && valid
	valid = reporter.check(len(info.Members) > 0, "%s actor info must have at least one member", infoSection.Pos) && valid
	info.Sensitive = meta.Sensitive
	info.Pub = ga.Pub
	for _, member := range info.Members {
		if reporter.cancelled() {
			break
		}
		if member.Name == identifierField {
			kind := member.Type
			valid = reporter.check(kind.Kind == schema.TypeKindScalar && !kind.Nullable &&
				(kind.Scalar == schema.ScalarString || kind.Scalar == schema.ScalarUUID || kind.Scalar == schema.ScalarInt),
				"%s @identifier requires a non-nullable string, uuid, or int field", member.Pos) && valid
		}
	}
	return info, identifierField, valid
}

func actorAuthSection(reporter *_DiagnosticReporter, ga *grammar.Actor) (*grammar.ActorAuth, bool) {
	var auth *grammar.ActorAuth
	var authPos lexer.Position
	valid := true
	for _, section := range ga.Sections {
		if reporter.cancelled() {
			break
		}
		if section.Auth == nil {
			continue
		}
		if auth != nil {
			reporter.reportDuplicatef("%s duplicated actor auth found, also present at %s", section.Auth.Pos, authPos)
			valid = false
			continue
		}
		auth = section.Auth
		authPos = section.Auth.Pos
	}
	if auth == nil {
		return nil, valid
	}
	valid = reporter.check(auth.Credential != nil && auth.Info != nil,
		"%s actor %s auth must define credential and info together", ga.Name.Pos, ga.Name.Value) && valid
	if auth.Credential == nil || auth.Info == nil {
		return nil, false
	}
	return auth, valid
}

func parseActorPermission(reporter *_DiagnosticReporter, ga *grammar.Actor) (*schema.ActorPermission, bool) {
	var permission *grammar.ActorPermission
	var permissionPos lexer.Position
	valid := true
	for _, section := range ga.Sections {
		if reporter.cancelled() {
			break
		}
		if section.Permission == nil {
			continue
		}
		if permission != nil {
			reporter.reportDuplicatef("%s duplicated actor permission found, also present at %s", section.Permission.Pos, permissionPos)
			valid = false
			continue
		}
		permission = section.Permission
		permissionPos = section.Permission.Pos
	}
	if permission == nil {
		return nil, valid
	}
	return new(schema.ActorPermission{Pos: position(permission.Pos)}), valid
}

func buildActorAuthService(actor *schema.Actor) *schema.Service {
	if actor.Auth == nil {
		return nil
	}
	serviceName := actor.Name + "AuthService"
	credentialType := dataRefType(actor.Auth.Credential)
	auth := actor.Auth
	result := new(schema.Data{
		Name:     actor.Name + "AuthResult",
		Domain:   auth.Info.Domain,
		SkelName: auth.Info.Domain + "." + actor.Name + "AuthResult",
		Pos:      auth.Pos,
		Kind:     schema.DataKindData,
		Pub:      actor.Pub,
		Members: []*schema.DataMember{
			new(schema.DataMember{
				Pos:  auth.Pos,
				Name: "info",
				Type: dataRefType(auth.Info),
			}),
		},
	})
	auth.Result = result
	if len(auth.Impersonators) > 0 {
		auth.ImpersonatorInfo = new(schema.Data{
			Name:     actor.Name + "ImpersonatorInfo",
			Domain:   auth.Info.Domain,
			SkelName: auth.Info.Domain + "." + actor.Name + "ImpersonatorInfo",
			Pos:      auth.Pos,
			Kind:     schema.DataKindData,
			Pub:      actor.Pub,
		})
		kind := dataRefType(auth.ImpersonatorInfo)
		kind.Nullable = true
		result.Members = append(result.Members, new(schema.DataMember{
			Pos:  auth.Pos,
			Name: "impersonatorInfo",
			Type: kind,
		}))
	}
	infoType := dataRefType(result)
	credentialArgument := &schema.Argument{
		Name: "credential",
		Pos:  actor.Auth.Credential.Pos,
		Type: credentialType,
	}
	credentialMethod := &schema.Method{
		Name:       "auth",
		SkelName:   "auth",
		Pos:        actor.Pos,
		AuthMode:   schema.AuthModeRequired,
		Arguments:  []*schema.Argument{credentialArgument},
		ResultType: infoType,
	}
	credentialMethod.ArgumentsData = &schema.Data{
		Name:     serviceName + "AuthArguments",
		Domain:   actor.Auth.Credential.Domain,
		SkelName: actor.Auth.Credential.Domain + "." + serviceName + "AuthArguments",
		Members:  buildArgumentMembers(credentialMethod.Arguments),
	}
	actor.Auth.Method = credentialMethod
	return &schema.Service{
		Name:     serviceName,
		SkelName: actor.Auth.Credential.Domain + "." + serviceName,
		Pos:      actor.Pos,
		Methods:  []*schema.Method{credentialMethod},
	}
}

func buildActorPermissionService(actor *schema.Actor) *schema.Service {
	if actor.Permission == nil {
		return nil
	}
	serviceName := actor.Name + "PermissionService"
	domain := strings.TrimSuffix(actor.SkelName, "."+actor.Name)
	skelPrefix := domain + "."
	codesArgument := &schema.Argument{
		Name: "codes",
		Pos:  actor.Pos,
		Type: &schema.Type{
			Kind: schema.TypeKindList,
			List: &schema.ListType{Element: &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}},
		},
	}
	method := &schema.Method{
		Name:      "checkCodes",
		SkelName:  "checkCodes",
		Pos:       actor.Pos,
		AuthMode:  schema.AuthModeRequired,
		Arguments: []*schema.Argument{codesArgument},
		ResultType: &schema.Type{
			Kind: schema.TypeKindMap,
			Map: &schema.MapType{
				Key:   &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString},
				Value: &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarBoolean},
			},
		},
	}
	method.ArgumentsData = &schema.Data{
		Name:     serviceName + "CheckCodesArguments",
		Domain:   domain,
		SkelName: skelPrefix + serviceName + "CheckCodesArguments",
		Members:  buildArgumentMembers(method.Arguments),
	}
	actor.Permission.Method = method
	return &schema.Service{
		Name:     serviceName,
		SkelName: skelPrefix + serviceName,
		Pos:      actor.Pos,
		Methods:  []*schema.Method{method},
	}
}

func dataRefType(data *schema.Data) *schema.Type {
	return &schema.Type{
		Kind:     schema.TypeKindData,
		Data:     data,
		SkelName: data.SkelName,
	}
}

func (p *Analysis) normalizeImpersonators(actor *schema.Actor, allowMissingImports bool) {
	auth := actor.Auth
	if auth.ImpersonatorInfo == nil {
		return
	}
	auth.ImpersonatorInfo.Members = nil
	fields := map[string]bool{}
	resolved := map[string]bool{}
	for _, ref := range auth.Impersonators {
		if p.reporter.cancelled() {
			break
		}
		qualifier, _, external := nameutil.SplitQualified(ref.Actor)
		if external && p.importsMap[qualifier] == nil && allowMissingImports {
			continue
		}
		target := p.actorByRef(ref.Actor)
		if !p.reporter.checkReference(target != nil, "%s undefined impersonator actor %s", ref.Pos, ref.Actor) {
			continue
		}
		if !p.reporter.check(target.Auth != nil, "%s impersonator actor %s must declare auth info", ref.Pos, ref.Actor) {
			continue
		}
		if !p.reporter.check(!external || target.Pub, "%s impersonator actor %s is not public", ref.Pos, ref.Actor) {
			continue
		}
		if !p.reporter.check(!actor.Pub || target.Pub, "%s public actor %s cannot expose private impersonator %s", ref.Pos, actor.Name, ref.Actor) {
			continue
		}
		ref.ActorSkelName = target.SkelName
		ref.InfoFieldName = nameutil.ToLowerCamel(target.Name) + "Info"
		if !p.reporter.checkNotDuplicate(resolved[target.SkelName], "%s duplicated impersonator actor %s", ref.Pos, ref.Actor) {
			continue
		}
		resolved[target.SkelName] = true
		if !p.reporter.checkNotDuplicate(fields[ref.InfoFieldName], "%s impersonator info field %s conflicts with another actor", ref.Pos, ref.InfoFieldName) {
			continue
		}
		fields[ref.InfoFieldName] = true
		kind := dataRefType(target.Auth.Info)
		kind.Nullable = true
		if external {
			imported := p.importsMap[qualifier]
			kind.ExternalDomain = imported.Domain.name
			kind.ExternalAlias = imported.Schema.Alias
			kind.ExternalAliasExplicit = imported.Schema.ExplicitAlias
		}
		auth.ImpersonatorInfo.Members = append(auth.ImpersonatorInfo.Members, new(schema.DataMember{
			Pos:  ref.Pos,
			Name: ref.InfoFieldName,
			Type: kind,
		}))
	}
}
