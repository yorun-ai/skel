package analyzer

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
)

func parseService(reporter *_DiagnosticReporter, gs *grammar.Service) (*model.Service, bool) {
	valid := reporter.checkNot(gs.Api && gs.Pub, "%s api and pub are mutually exclusive", gs.Name.Pos)
	valid = reporter.checkNot(gs.Ext && (gs.Api || gs.Pub), "%s ext, api and pub are mutually exclusive", gs.Name.Pos) && valid
	suffix := "Service"
	if gs.Api {
		suffix = "ApiService"
	}
	valid = checkCaseAdvanced(reporter, "Service", "", suffix, caseTypeCamel, gs.Name) && valid
	meta, metaValid := parseDecoratorMeta(reporter, gs.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s service does not support decorator @example", gs.Name.Pos) && valid
	audiences, audiencesValid := parseServiceAudiences(reporter, serviceAudiences(gs))
	valid = audiencesValid && valid
	valid = reporter.checkNot(gs.Api && audiencesValid && len(audiences) == 0, "%s API service must declare at least one for Actor", gs.Name.Pos) && valid
	authMarker, authValid := serviceAuthMarker(reporter, gs)
	valid = authValid && valid
	authMode, authModeValid := parseAuthMode(reporter, authMarker, model.AuthModeUnset)
	valid = authModeValid && valid
	requireGrammar, requireSectionValid := serviceRequire(reporter, gs)
	valid = requireSectionValid && valid
	require, requireValid := parseRequire(reporter, requireGrammar)
	valid = requireValid && valid
	methods, methodsValid := parseMethods(reporter, gs.Name, serviceMethods(gs))
	valid = methodsValid && valid
	return &model.Service{
		Pos:              position(gs.Name.Pos),
		Name:             gs.Name.Value,
		Pub:              gs.Pub,
		Api:              gs.Api,
		Ext:              gs.Ext,
		Audiences:        audiences,
		Auth:             authMode,
		AuthPos:          authMarkerPosition(authMarker),
		Require:          require,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Methods:          methods,
		SkelName:         "",
	}, valid
}

func serviceAudiences(gs *grammar.Service) []*grammar.ServiceAudience {
	if len(gs.Sections) == 0 {
		return gs.Audiences
	}
	audiences := make([]*grammar.ServiceAudience, 0)
	for _, section := range gs.Sections {
		if section.Audience != nil {
			audiences = append(audiences, section.Audience)
		}
	}
	return audiences
}

func serviceMethods(gs *grammar.Service) []*grammar.Method {
	if len(gs.Sections) == 0 {
		return gs.Methods
	}
	methods := make([]*grammar.Method, 0)
	for _, section := range gs.Sections {
		if section.Method != nil {
			method := *section.Method
			method.Decorators = append(append([]*grammar.Decorator{}, section.Decorators...), section.Method.Decorators...)
			methods = append(methods, &method)
		}
	}
	return methods
}

func serviceAuthMarker(reporter *_DiagnosticReporter, gs *grammar.Service) (*grammar.AuthMarker, bool) {
	if len(gs.Sections) == 0 {
		return gs.Auth, true
	}
	var marker *grammar.AuthMarker
	var markerPos lexer.Position
	valid := true
	for _, section := range gs.Sections {
		if reporter.cancelled() {
			break
		}
		if section.Auth == nil {
			continue
		}
		if marker != nil {
			reporter.reportDuplicatef("%s duplicated service auth marker found, also present at %s", section.Auth.Pos, markerPos)
			valid = false
			continue
		}
		marker = section.Auth
		markerPos = section.Auth.Pos
	}
	return marker, valid
}

func serviceRequire(reporter *_DiagnosticReporter, gs *grammar.Service) (*grammar.Require, bool) {
	if len(gs.Sections) == 0 {
		return nil, true
	}
	var require *grammar.Require
	var requirePos lexer.Position
	valid := true
	for _, section := range gs.Sections {
		if reporter.cancelled() {
			break
		}
		if section.Require == nil {
			continue
		}
		if require != nil {
			reporter.reportDuplicatef("%s duplicated service require found, also present at %s", section.Require.Pos, requirePos)
			valid = false
			continue
		}
		require = section.Require
		requirePos = section.Require.Pos
	}
	return require, valid
}

func parseAuthMode(reporter *_DiagnosticReporter, marker *grammar.AuthMarker, defaultMode model.AuthMode) (model.AuthMode, bool) {
	if marker == nil {
		return defaultMode, true
	}
	switch marker.Value {
	case "required", "optional", "anonymous":
		return model.AuthMode(marker.Value), true
	case "off":
		reporter.reportf("%s auth off is only supported on web declarations", marker.Pos)
		return defaultMode, false
	case string(model.AuthModeAuth):
		return model.AuthModeAuth, true
	case string(model.AuthModeNoAuth):
		return model.AuthModeNoAuth, true
	}
	reporter.reportf("%s unexpected auth marker %s", marker.Pos, marker.Value)
	return defaultMode, false
}

func parseServiceAudiences(reporter *_DiagnosticReporter, audiences []*grammar.ServiceAudience) ([]*model.ActorAudience, bool) {
	if len(audiences) == 0 {
		return []*model.ActorAudience{}, true
	}
	parsed := make([]*model.ActorAudience, 0, len(audiences))
	audiencePos := map[string]lexer.Position{}
	valid := true
	for _, audience := range audiences {
		if reporter.cancelled() {
			break
		}
		actorIdent := audience.Actor.Parts[len(audience.Actor.Parts)-1]
		valid = checkCaseAdvanced(reporter, "Actor", "", "Actor", caseTypeCamel, actorIdent) && valid
		via := ""
		if audience.Via != nil {
			via = audience.Via.Value
		}
		name := audience.Actor.String()
		key := fmt.Sprintf("%s:%s", name, via)
		duplicatedPosition, duplicated := audiencePos[key]
		if duplicated {
			if via != "" {
				reporter.reportDuplicatef("%s duplicated service audience %s via %s found, also present at %s", actorIdent.Pos, name, via, duplicatedPosition)
			} else {
				reporter.reportDuplicatef("%s duplicated service audience %s found, also present at %s", actorIdent.Pos, name, duplicatedPosition)
			}
			valid = false
			continue
		}
		audiencePos[key] = actorIdent.Pos
		parsed = append(parsed, &model.ActorAudience{
			Actor: name,
			Via:   via,
			Pos:   position(audience.Pos),
		})
	}
	return parsed, valid
}

func authMarkerPosition(marker *grammar.AuthMarker) model.Position {
	if marker == nil {
		return model.Position{}
	}
	return position(marker.Pos)
}
