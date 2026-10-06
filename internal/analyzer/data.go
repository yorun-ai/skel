package analyzer

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/skelmeta"
)

func parseData(reporter *_DiagnosticReporter, gs *grammar.Data) (*model.Data, bool) {
	return parseDataLike(reporter, gs, model.DataKindData)
}

func parseConfig(reporter *_DiagnosticReporter, gs *grammar.Data) (*model.Data, bool) {
	return parseDataLike(reporter, gs, model.DataKindConfig)
}

func parseEvent(reporter *_DiagnosticReporter, ge *grammar.Event) (*model.Data, bool) {
	members := []*grammar.DataMember{}
	if ge.Payload != nil {
		members = ge.Payload.Members
	}
	event, valid := parseDataLike(reporter, &grammar.Data{
		Pos:            ge.Pos,
		Decorators:     ge.Decorators,
		Pub:            ge.Pub,
		Name:           ge.Name,
		Qualifier:      ge.Qualifier,
		Members:        members,
		TypeParameters: ge.TypeParameters,
	}, model.DataKindEvent)
	event.Ext = ge.Ext
	valid = reporter.checkNot(ge.Ext && ge.Pub, "%s ext and pub are mutually exclusive", ge.Name.Pos) && valid
	if ge.Payload == nil {
		return event, valid
	}
	payloadMeta, payloadValid := parseDecoratorMeta(reporter, ge.Payload.Decorators, _DecoratorContext{
		allowSensitive: true,
	})
	event.Sensitive = payloadMeta.Sensitive
	return event, payloadValid && valid
}

func parseDataLike(reporter *_DiagnosticReporter, gs *grammar.Data, kind model.DataKind) (*model.Data, bool) {
	valid := checkCase(reporter, "Data", caseTypeCamel, gs.Name)
	if kind == model.DataKindConfig {
		valid = reporter.check(strings.HasSuffix(gs.Name.Value, "Config"), "%s Config name must end with Config", gs.Name.Pos) && valid
		qualifierValid := reporter.check(gs.Qualifier != nil,
			"%s Config %s requires lifecycle qualifier eternal/instant",
			gs.Name.Pos, gs.Name.Value)
		valid = qualifierValid && valid
		if qualifierValid {
			valid = reporter.check(
				gs.Qualifier.Value == string(model.ConfigLifecycleEternal) || gs.Qualifier.Value == string(model.ConfigLifecycleInstant),
				"%s Config %s has invalid lifecycle qualifier %s, expected eternal/instant",
				gs.Qualifier.Pos, gs.Name.Value, gs.Qualifier.Value) && valid
		}
		valid = reporter.check(len(gs.TypeParameters) == 0,
			"%s Config %s does not support type parameters",
			gs.Name.Pos, gs.Name.Value) && valid
	} else if kind == model.DataKindEvent {
		valid = reporter.check(strings.HasSuffix(gs.Name.Value, "Event"), "%s Event name must end with Event", gs.Name.Pos) && valid
		valid = reporter.check(len(gs.TypeParameters) == 0,
			"%s Event %s does not support type parameters",
			gs.Name.Pos, gs.Name.Value) && valid
		if gs.Qualifier != nil {
			reporter.reportf("%s Event %s does not support qualifier %s", gs.Qualifier.Pos, gs.Name.Value, gs.Qualifier.Value)
			valid = false
		}
	} else {
		valid = checkNotReservedKindSuffix(reporter, "Data", gs.Name) && valid
		if gs.Qualifier != nil {
			reporter.reportf("%s Data %s does not support qualifier %s", gs.Qualifier.Pos, gs.Name.Value, gs.Qualifier.Value)
			valid = false
		}
	}
	parsedData := &model.Data{
		Pos:     position(gs.Name.Pos),
		Name:    gs.Name.Value,
		Kind:    kind,
		Pub:     gs.Pub,
		Members: []*model.DataMember{},
	}
	if gs.Qualifier != nil {
		parsedData.Lifecycle = model.ConfigLifecycle(gs.Qualifier.Value)
	}
	meta, metaValid := parseDecoratorMeta(reporter, gs.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowSensitive:  kind != model.DataKindEvent,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	parsedData.Description = meta.Description
	parsedData.Sensitive = meta.Sensitive
	parsedData.Deprecated = meta.Deprecated
	parsedData.DeprecatedReason = meta.DeprecatedReason
	typeParamPos := map[string]lexer.Position{}
	memberPos := map[string]lexer.Position{}

	for _, grammarTypeParameter := range gs.TypeParameters {
		if reporter.cancelled() {
			break
		}
		typeParameter, parameterValid := parseTypeParameter(reporter, grammarTypeParameter)
		valid = parameterValid && valid
		duplicatedPosition, duplicated := typeParamPos[typeParameter.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated TypeParameter %s found, also present at %s", typeParameter.Pos, typeParameter.Name, duplicatedPosition)
			valid = false
			continue
		}
		typeParamPos[typeParameter.Name] = lexer.Position{Filename: typeParameter.Pos.File, Line: typeParameter.Pos.Line, Column: typeParameter.Pos.Column}
		parsedData.TypeParameters = append(parsedData.TypeParameters, typeParameter)
	}

	for _, grammarMember := range gs.Members {
		if reporter.cancelled() {
			break
		}
		member, memberValid := parseDataMember(reporter, grammarMember)
		valid = memberValid && valid
		valid = reporter.check(member.Name != skelmeta.SensitiveMarkerFieldName(),
			"%s DataMember %s is reserved for the generated sensitive marker method",
			member.Pos, skelmeta.SensitiveMarkerFieldName()) && valid
		duplicatedPosition, duplicated := memberPos[member.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated DataMember %s found, also present at %s", member.Pos, member.Name, duplicatedPosition)
			valid = false
			continue
		}
		memberPos[member.Name] = lexer.Position{Filename: member.Pos.File, Line: member.Pos.Line, Column: member.Pos.Column}
		parsedData.Members = append(parsedData.Members, member)
	}

	return parsedData, valid
}

func parseTypeParameter(reporter *_DiagnosticReporter, gtp *grammar.TypeParameter) (*model.TypeParameter, bool) {
	valid := checkCaseAdvanced(reporter, "TypeParameter", "T", "", caseTypeCamel, gtp.Name)
	valid = reporter.checkNot(gtp.Nullable, "%s TypeParameter %s cannot be nullable", gtp.Pos, gtp.Name.Value) && valid
	return &model.TypeParameter{
		Name: gtp.Name.Value,
		Pos:  position(gtp.Name.Pos),
	}, valid
}

func parseDataMember(reporter *_DiagnosticReporter, gsm *grammar.DataMember) (*model.DataMember, bool) {
	valid := checkCase(reporter, "DataMember", caseTypeLowerCamel, gsm.Name)
	meta, metaValid := parseDecoratorMeta(reporter, gsm.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowExample:    true,
		allowSensitive:  true,
		allowDeprecated: true,
		requireDesc:     true,
	})
	valid = metaValid && valid
	memberType, typeValid := parseType(reporter, gsm.Type)
	valid = typeValid && valid
	return &model.DataMember{
		Pos:              position(gsm.Name.Pos),
		Name:             gsm.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Example:          meta.Example,
		Sensitive:        meta.Sensitive,
		Type:             memberType,
	}, valid
}
