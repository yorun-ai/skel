package analyzer

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser/grammar"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

func parseTask(reporter *_DiagnosticReporter, gt *grammar.Task) (*model.Task, bool) {
	valid := checkCaseAdvanced(reporter, "Task", "", "Task", caseTypeCamel, gt.Name)
	meta, metaValid := parseDecoratorMeta(reporter, gt.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s task does not support decorator @example", gt.Name.Pos) && valid
	triggers, triggersValid := parseTaskTriggers(reporter, gt.Name, gt.Triggers)
	valid = triggersValid && valid
	return &model.Task{
		Pos:              position(gt.Name.Pos),
		Name:             gt.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Triggers:         triggers,
	}, valid
}

func parseTaskTriggers(reporter *_DiagnosticReporter, owner *grammar.Identifier, triggers []*grammar.TaskTrigger) ([]*model.TaskTrigger, bool) {
	parsedTriggers := make([]*model.TaskTrigger, 0, len(triggers))
	triggerPos := map[string]lexer.Position{}
	valid := true

	for _, grammarTrigger := range triggers {
		if reporter.cancelled() {
			break
		}
		trigger, triggerValid := parseTaskTrigger(reporter, grammarTrigger)
		valid = triggerValid && valid
		duplicatedPosition, duplicated := triggerPos[trigger.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated task trigger %s found, also present at %s", trigger.Pos, trigger.Name, duplicatedPosition)
			valid = false
			continue
		}
		if trigger.ArgumentsData != nil {
			trigger.ArgumentsData.Name = fmt.Sprintf("%s%s", owner.Value, trigger.ArgumentsData.Name)
		}
		triggerPos[trigger.Name] = lexer.Position{Filename: trigger.Pos.File, Line: trigger.Pos.Line, Column: trigger.Pos.Column}
		parsedTriggers = append(parsedTriggers, trigger)
	}

	valid = reporter.check(len(parsedTriggers) > 0, "%s missing task trigger for %s", owner.Pos, owner.Value) && valid

	return parsedTriggers, valid
}

func parseTaskTrigger(reporter *_DiagnosticReporter, gt *grammar.TaskTrigger) (*model.TaskTrigger, bool) {
	valid := checkCase(reporter, "TaskTrigger", caseTypeLowerCamel, gt.Name)
	meta, metaValid := parseDecoratorMeta(reporter, gt.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid

	trigger := &model.TaskTrigger{
		Pos:              position(gt.Name.Pos),
		Name:             gt.Name.Value,
		SkelName:         gt.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Arguments:        []*model.Argument{},
	}
	if gt.Input == nil {
		return trigger, valid
	}

	inputMeta, inputValid := parseDecoratorMeta(reporter, gt.Input.Decorators, _DecoratorContext{
		allowDesc:      true,
		allowSensitive: true,
	})
	valid = inputValid && valid
	trigger.InputDescription = inputMeta.Description
	trigger.ArgumentsSensitive = inputMeta.Sensitive
	argPos := map[string]lexer.Position{}
	for _, grammarArgument := range gt.Input.Arguments {
		if reporter.cancelled() {
			break
		}
		arg, argumentValid := parseArgument(reporter, grammarArgument)
		valid = argumentValid && valid
		duplicatedPosition, duplicated := argPos[arg.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated Argument %s found, also present at %s", arg.Pos, arg.Name, duplicatedPosition)
			valid = false
			continue
		}
		argPos[arg.Name] = lexer.Position{Filename: arg.Pos.File, Line: arg.Pos.Line, Column: arg.Pos.Column}
		trigger.Arguments = append(trigger.Arguments, arg)
	}

	if len(trigger.Arguments) > 0 {
		trigger.ArgumentsData = &model.Data{
			Name:    fmt.Sprintf("%sArguments", nameutil.ToCamel(trigger.Name)),
			Members: buildArgumentMembers(trigger.Arguments),
		}
	}

	return trigger, valid
}
