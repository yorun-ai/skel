package analyzer

import (
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser/grammar"
	"go.yorun.ai/skelc/internal/util/webpath"
)

func parseWeb(reporter *_DiagnosticReporter, gw *grammar.Web, pub bool) (*model.Web, bool) {
	valid := checkCaseAdvanced(reporter, "Web", "", "Web", caseTypeCamel, gw.Name)
	meta, metaValid := parseDecoratorMeta(reporter, gw.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s web does not support decorator @example", gw.Name.Pos) && valid
	valid = reporter.checkNot(pub, "%s web %s does not support pub", gw.Name.Pos, gw.Name.Value) && valid
	audiences, audiencesValid := parseWebAudiences(reporter, gw.Audiences)
	valid = audiencesValid && valid
	valid = reporter.check(len(audiences) > 0, "%s web %s must declare at least one actor", gw.Name.Pos, gw.Name.Value) && valid
	authMode := model.AuthModeUnset
	var authMarker *grammar.AuthMarker
	for _, section := range gw.Sections {
		if section.Auth == nil {
			continue
		}
		if authMarker != nil {
			reporter.reportDuplicatef("%s duplicated web auth marker", section.Auth.Pos)
			valid = false
			continue
		}
		authMarker = section.Auth
		if authMarker.Value == "off" {
			authMode = model.AuthModeOff
		} else {
			parsed, ok := parseAuthMode(reporter, authMarker, model.AuthModeUnset)
			authMode = parsed
			valid = ok && valid
		}
	}
	mountPath := ""
	for index, mount := range gw.Mounts {
		if reporter.cancelled() {
			break
		}
		valid = reporter.check(index == 0, "%s web %s must declare mount at most once", mount.Pos, gw.Name.Value) && valid
		if err := webpath.Validate(mount.Path.Value); err != nil {
			valid = reporter.check(false, "%s invalid web mount path: %s", mount.Path.Pos, err) && valid
		}
		mountPath = mount.Path.Value
	}
	return &model.Web{
		Pos:              position(gw.Name.Pos),
		Name:             gw.Name.Value,
		SkelName:         "",
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Audiences:        audiences,
		Auth:             authMode,
		AuthPos:          authMarkerPosition(authMarker),
		MountPath:        mountPath,
	}, valid
}

func parseWebAudiences(reporter *_DiagnosticReporter, audiences []*grammar.WebAudience) ([]*model.ActorAudience, bool) {
	serviceAudiences := make([]*grammar.ServiceAudience, 0, len(audiences))
	for _, audience := range audiences {
		if reporter.cancelled() {
			break
		}
		serviceAudiences = append(serviceAudiences, &grammar.ServiceAudience{
			Pos:     audience.Pos,
			Keyword: audience.Keyword,
			Actor:   audience.Actor,
			Via:     audience.Via,
		})
	}
	return parseServiceAudiences(reporter, serviceAudiences)
}
