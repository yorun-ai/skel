package analyzer

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
)

var actorViaKinds = []schema.ActorViaKind{
	schema.ActorViaClient,
	schema.ActorViaAgent,
	schema.ActorViaOpenAPI,
}

func parseActor(reporter *_DiagnosticReporter, ga *grammar.Actor) (*schema.Actor, bool) {
	valid := checkCaseAdvanced(reporter, "Actor", "", "Actor", caseTypeCamel, ga.Name)
	meta, metaValid := parseDecoratorMeta(reporter, ga.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s actor does not support decorator @example", ga.Name.Pos) && valid
	vias, viasValid := parseActorVias(reporter, ga.Name, ga.Vias)
	valid = viasValid && valid
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
		Vias:             vias,
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
	return new(schema.ActorAuth{
		Pos: position(authSection.Pos), Credential: credential, Info: info, IdentifierField: identifierField,
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

func parseActorVias(reporter *_DiagnosticReporter, owner *grammar.Identifier, grammarVias []*grammar.ActorVia) ([]*schema.ActorVia, bool) {
	valid := reporter.check(len(grammarVias) > 0, "%s actor %s must have at least one via", owner.Pos, owner.Value)

	parsedVias := make([]*schema.ActorVia, 0, len(grammarVias))
	viaPos := map[string]lexer.Position{}
	for _, grammarVia := range grammarVias {
		if reporter.cancelled() {
			break
		}
		via, viaValid := parseActorVia(reporter, grammarVia)
		valid = viaValid && valid
		duplicatedPosition, duplicated := viaPos[via.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated actor via %s found, also present at %s", via.Pos, via.Name, duplicatedPosition)
			valid = false
			continue
		}
		viaPos[via.Name] = lexer.Position{Filename: via.Pos.File, Line: via.Pos.Line, Column: via.Pos.Column}
		parsedVias = append(parsedVias, via)
	}
	return parsedVias, valid
}

func parseActorVia(reporter *_DiagnosticReporter, gv *grammar.ActorVia) (*schema.ActorVia, bool) {
	valid := checkCase(reporter, "ActorVia", caseTypeLowerCamel, gv.Name)
	_, ok := sliceutil.Find(actorViaKinds, func(candidate schema.ActorViaKind) bool {
		return string(candidate) == gv.Name.Value
	})
	valid = reporter.check(ok, "%s unexpected actor via %s, supported=client/agent/openapi", gv.Name.Pos, gv.Name.Value) && valid
	return &schema.ActorVia{
		Name: gv.Name.Value,
		Pos:  position(gv.Name.Pos),
	}, valid
}

func buildActorAuthService(actor *schema.Actor) *schema.Service {
	if actor.Auth == nil {
		return nil
	}
	serviceName := actor.Name + "AuthService"
	credentialType := dataRefType(actor.Auth.Credential)
	infoType := dataRefType(actor.Auth.Info)
	credentialArgument := &schema.Argument{
		Name: "credential",
		Pos:  actor.Auth.Credential.Pos,
		Type: credentialType,
	}
	credentialMethod := &schema.Method{
		Name:       "auth",
		SkelName:   "auth",
		Pos:        actor.Pos,
		Auth:       schema.AuthModeRequired,
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
			List: &schema.ListType{Value: &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}},
		},
	}
	method := &schema.Method{
		Name:      "checkCodes",
		SkelName:  "checkCodes",
		Pos:       actor.Pos,
		Auth:      schema.AuthModeRequired,
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
