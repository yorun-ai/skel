package analyzer

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/sliceutil"
)

var actorViaKinds = []model.ActorViaKind{
	model.ActorViaClient,
	model.ActorViaAgent,
	model.ActorViaOpenAPI,
}

type _ActorAuth struct {
	Credential      *model.Data
	Info            *model.Data
	IdentifierField string
}

func parseActor(reporter *_DiagnosticReporter, ga *grammar.Actor) (*model.Actor, bool) {
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
	var authCredential *model.Data
	var authInfo *model.Data
	var identifierField string
	if auth != nil {
		authCredential = auth.Credential
		authInfo = auth.Info
		identifierField = auth.IdentifierField
	}
	permEnabled, permissionValid := actorPermissionDeclared(reporter, ga)
	valid = permissionValid && valid
	return &model.Actor{
		Pos:              position(ga.Name.Pos),
		Name:             ga.Name.Value,
		SkelName:         "",
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Pub:              ga.Pub,
		Vias:             vias,
		AuthEnabled:      auth != nil,
		AuthCredential:   authCredential,
		AuthInfo:         authInfo,
		IdentifierField:  identifierField,
		PermEnabled:      permEnabled,
	}, valid
}

func parseActorAuth(reporter *_DiagnosticReporter, ga *grammar.Actor) (*_ActorAuth, bool) {
	authSection, valid := actorAuthSection(reporter, ga)
	if authSection == nil {
		return nil, valid
	}
	credential, credentialValid := parseActorCredential(reporter, ga, authSection)
	info, identifierField, infoValid := parseActorInfo(reporter, ga, authSection)
	return &_ActorAuth{Credential: credential, Info: info, IdentifierField: identifierField}, credentialValid && infoValid && valid
}

func parseActorCredential(reporter *_DiagnosticReporter, ga *grammar.Actor, authSection *grammar.ActorAuth) (*model.Data, bool) {
	credentialSection := authSection.Credential
	meta, metaValid := parseDecoratorMeta(reporter, credentialSection.Decorators, _DecoratorContext{
		allowSensitive: true,
	})
	name := &grammar.Identifier{
		Pos:   ga.Name.Pos,
		Value: ga.Name.Value + "Credential",
	}
	credential, valid := parseDataLike(reporter, &grammar.Data{
		Pos:     credentialSection.Pos,
		Pub:     ga.Pub,
		Name:    name,
		Members: credentialSection.Members,
	}, model.DataKindData)
	valid = metaValid && valid
	credential.Sensitive = meta.Sensitive
	valid = reporter.check(len(credential.Members) > 0, "%s actor credential must have at least one member", credentialSection.Pos) && valid
	hasRequiredField := false
	for _, member := range credential.Members {
		if reporter.cancelled() {
			break
		}
		valid = reporter.check(member.Type.Kind == model.TypeKindScalar && member.Type.Scalar == model.ScalarString,
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

func parseActorInfo(reporter *_DiagnosticReporter, ga *grammar.Actor, authSection *grammar.ActorAuth) (*model.Data, string, bool) {
	infoSection := authSection.Info
	meta, metaValid := parseDecoratorMeta(reporter, infoSection.Decorators, _DecoratorContext{
		allowSensitive: true,
	})
	name := &grammar.Identifier{
		Pos:   ga.Name.Pos,
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
	}, model.DataKindData)
	valid = metaValid && valid
	info.Sensitive = meta.Sensitive
	info.Pub = ga.Pub
	for _, member := range info.Members {
		if reporter.cancelled() {
			break
		}
		if member.Name == identifierField {
			kind := member.Type
			valid = reporter.check(kind.Kind == model.TypeKindScalar && !kind.Nullable &&
				(kind.Scalar == model.ScalarString || kind.Scalar == model.ScalarUUID || kind.Scalar == model.ScalarInt),
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

func actorPermissionDeclared(reporter *_DiagnosticReporter, ga *grammar.Actor) (bool, bool) {
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
		return false, valid
	}
	return true, valid
}

func parseActorVias(reporter *_DiagnosticReporter, owner *grammar.Identifier, grammarVias []*grammar.ActorVia) ([]*model.ActorVia, bool) {
	valid := reporter.check(len(grammarVias) > 0, "%s actor %s must have at least one via", owner.Pos, owner.Value)

	parsedVias := make([]*model.ActorVia, 0, len(grammarVias))
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

func parseActorVia(reporter *_DiagnosticReporter, gv *grammar.ActorVia) (*model.ActorVia, bool) {
	valid := checkCase(reporter, "ActorVia", caseTypeLowerCamel, gv.Name)
	_, ok := sliceutil.Find(actorViaKinds, func(candidate model.ActorViaKind) bool {
		return string(candidate) == gv.Name.Value
	})
	valid = reporter.check(ok, "%s unexpected actor via %s, supported=client/agent/openapi", gv.Name.Pos, gv.Name.Value) && valid
	return &model.ActorVia{
		Name: gv.Name.Value,
		Pos:  position(gv.Name.Pos),
	}, valid
}

func buildActorAuthService(actor *model.Actor) *model.Service {
	if !actor.AuthEnabled {
		return nil
	}
	serviceName := actor.Name + "AuthService"
	credentialType := dataRefType(actor.AuthCredential)
	infoType := dataRefType(actor.AuthInfo)
	credentialArgument := &model.Argument{
		Name: "credential",
		Pos:  actor.AuthCredential.Pos,
		Type: credentialType,
	}
	credentialMethod := &model.Method{
		Name:       "auth",
		SkelName:   "auth",
		Pos:        actor.Pos,
		Auth:       model.AuthModeRequired,
		Arguments:  []*model.Argument{credentialArgument},
		ResultType: infoType,
	}
	credentialMethod.ArgumentsData = &model.Data{
		Name:     serviceName + "AuthArguments",
		Domain:   actor.AuthCredential.Domain,
		SkelName: actor.AuthCredential.Domain + "." + serviceName + "AuthArguments",
		Members:  buildArgumentMembers(credentialMethod.Arguments),
	}
	actor.AuthMethod = credentialMethod
	return &model.Service{
		Name:     serviceName,
		SkelName: actor.AuthCredential.Domain + "." + serviceName,
		Pos:      actor.Pos,
		Methods:  []*model.Method{credentialMethod},
	}
}

func buildActorPermissionService(actor *model.Actor) *model.Service {
	if !actor.PermEnabled {
		return nil
	}
	serviceName := actor.Name + "PermissionService"
	domain := strings.TrimSuffix(actor.SkelName, "."+actor.Name)
	skelPrefix := domain + "."
	codesArgument := &model.Argument{
		Name: "codes",
		Pos:  actor.Pos,
		Type: &model.Type{
			Kind: model.TypeKindList,
			List: &model.ListType{Value: &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString}},
		},
	}
	method := &model.Method{
		Name:      "checkCodes",
		SkelName:  "checkCodes",
		Pos:       actor.Pos,
		Auth:      model.AuthModeRequired,
		Arguments: []*model.Argument{codesArgument},
		ResultType: &model.Type{
			Kind: model.TypeKindMap,
			Map: &model.MapType{
				Key:   &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString},
				Value: &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarBoolean},
			},
		},
	}
	method.ArgumentsData = &model.Data{
		Name:     serviceName + "CheckCodesArguments",
		Domain:   domain,
		SkelName: skelPrefix + serviceName + "CheckCodesArguments",
		Members:  buildArgumentMembers(method.Arguments),
	}
	actor.PermMethod = method
	return &model.Service{
		Name:     serviceName,
		SkelName: skelPrefix + serviceName,
		Pos:      actor.Pos,
		Methods:  []*model.Method{method},
	}
}

func dataRefType(data *model.Data) *model.Type {
	return &model.Type{
		Kind:     model.TypeKindData,
		Data:     data,
		SkelName: data.SkelName,
	}
}
