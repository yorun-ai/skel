package analyzer

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser/grammar"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

func buildResourceCheckService(domainName string, resource *model.Resource) *model.Service {
	methods := []*model.Method{}
	serviceName := resource.Name + "CheckService"
	for _, check := range resource.Checks {
		methods = append(methods, prepareResourceCheckMethod(domainName, serviceName, check))
	}
	for _, action := range resource.Actions {
		for _, check := range action.Checks {
			methods = append(methods, prepareResourceCheckMethod(domainName, serviceName, check))
		}
	}
	if len(methods) == 0 {
		return nil
	}
	return &model.Service{
		Pos:      resource.Pos,
		Name:     serviceName,
		SkelName: domainName + "." + serviceName,
		Auth:     model.AuthModeAuth,
		Methods:  methods,
	}
}

func prepareResourceCheckMethod(domainName string, serviceName string, check *model.ResourceCheck) *model.Method {
	method := check.Method
	if method.ArgumentsData != nil {
		method.ArgumentsData.Name = serviceName + method.ArgumentsData.Name
		method.ArgumentsData.Domain = domainName
		method.ArgumentsData.SkelName = domainName + "." + method.ArgumentsData.Name
	}
	return method
}

func parseResource(reporter *_DiagnosticReporter, ge *grammar.Resource, pub bool) (*model.Resource, bool) {
	valid := checkCase(reporter, "Resource", caseTypeCamel, ge.Name)
	meta, metaValid := parseDecoratorMeta(reporter, ge.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s resource does not support decorator @example", ge.Name.Pos) && valid

	checks := make([]*model.ResourceCheck, 0, len(ge.Checks))
	checkPos := map[string]lexer.Position{}
	for _, grammarCheck := range ge.Checks {
		if reporter.cancelled() {
			break
		}
		check, checkValid := parseResourceCheck(reporter, "", grammarCheck)
		valid = checkValid && valid
		duplicatedPosition, duplicated := checkPos[check.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated resource check %s found, also present at %s", grammarCheck.Name.Pos, check.Name, duplicatedPosition)
			valid = false
			continue
		}
		checkPos[check.Name] = grammarCheck.Name.Pos
		checks = append(checks, check)
	}

	actions := make([]*model.ResourceAction, 0, len(ge.Actions))
	actionPos := map[string]lexer.Position{}
	for _, grammarAction := range ge.Actions {
		if reporter.cancelled() {
			break
		}
		action, actionValid := parseResourceAction(reporter, grammarAction, checkPos)
		valid = actionValid && valid
		duplicatedPosition, duplicated := actionPos[action.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated resource action %s found, also present at %s", action.Pos, action.Name, duplicatedPosition)
			valid = false
			continue
		}
		actionPos[action.Name] = lexer.Position{Filename: action.Pos.File, Line: action.Pos.Line, Column: action.Pos.Column}
		actions = append(actions, action)
	}
	valid = reporter.check(len(actions) > 0, "%s resource %s must have at least one action", ge.Name.Pos, ge.Name.Value) && valid

	return &model.Resource{
		Pos:              position(ge.Name.Pos),
		Name:             ge.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Pub:              pub,
		Checks:           checks,
		Actions:          actions,
	}, valid
}

func parseResourceAction(reporter *_DiagnosticReporter, ga *grammar.ResourceAction, resourceCheckPos map[string]lexer.Position) (*model.ResourceAction, bool) {
	valid := checkCase(reporter, "ResourceAction", caseTypeLowerCamel, ga.Name)
	meta, metaValid := parseDecoratorMeta(reporter, ga.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	valid = reporter.checkNot(meta.HasExample, "%s resource action does not support decorator @example", ga.Name.Pos) && valid

	checks := make([]*model.ResourceCheck, 0, len(ga.Checks))
	checkPos := map[string]lexer.Position{}
	for _, grammarCheck := range ga.Checks {
		if reporter.cancelled() {
			break
		}
		check, checkValid := parseResourceCheck(reporter, ga.Name.Value, grammarCheck)
		valid = checkValid && valid
		if duplicatedPosition, duplicated := resourceCheckPos[check.Name]; duplicated {
			reporter.reportDuplicatef("%s duplicated resource action check %s found, also present at %s", grammarCheck.Name.Pos, check.Name, duplicatedPosition)
			valid = false
			continue
		}
		duplicatedPosition, duplicated := checkPos[check.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated resource action check %s found, also present at %s", grammarCheck.Name.Pos, check.Name, duplicatedPosition)
			valid = false
			continue
		}
		checkPos[check.Name] = grammarCheck.Name.Pos
		checks = append(checks, check)
	}

	return &model.ResourceAction{
		Pos:              position(ga.Name.Pos),
		Name:             ga.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Checks:           checks,
	}, valid
}

func parseResourceCheck(reporter *_DiagnosticReporter, actionName string, gc *grammar.ResourceCheck) (*model.ResourceCheck, bool) {
	valid := checkCase(reporter, "ResourceCheck", caseTypeLowerCamel, gc.Name)
	meta, metaValid := parseDecoratorMeta(reporter, gc.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	args := []*model.Argument{}
	argPos := map[string]lexer.Position{}
	inputDescription := ""
	inputSensitive := false
	if gc.Input != nil {
		inputMeta, inputValid := parseDecoratorMeta(reporter, gc.Input.Decorators, _DecoratorContext{
			allowDesc:      true,
			allowSensitive: true,
		})
		valid = inputValid && valid
		inputDescription = inputMeta.Description
		inputSensitive = inputMeta.Sensitive

		for _, grammarArgument := range gc.Input.Arguments {
			if reporter.cancelled() {
				break
			}
			arg, argumentValid := parseArgument(reporter, grammarArgument)
			valid = argumentValid && valid
			if duplicatedPosition, duplicated := argPos[arg.Name]; duplicated {
				reporter.reportDuplicatef("%s duplicated Argument %s found, also present at %s", arg.Pos, arg.Name, duplicatedPosition)
				valid = false
				continue
			}
			argPos[arg.Name] = grammarArgument.Name.Pos
			args = append(args, arg)
		}
	}
	codeArgument := newPermissionCodeArgument()
	for suffix := 1; ; suffix++ {
		if reporter.cancelled() {
			break
		}
		if _, exists := argPos[codeArgument.Name]; !exists {
			break
		}
		codeArgument.Name = fmt.Sprintf("code%d", suffix)
	}
	args = append([]*model.Argument{codeArgument}, args...)
	methodName := "check" + nameutil.ToCamel(actionName) + nameutil.ToCamel(gc.Name.Value)
	if actionName == "" {
		methodName = "check" + nameutil.ToCamel(gc.Name.Value)
	}
	method := &model.Method{
		Pos:                position(gc.Name.Pos),
		Name:               methodName,
		SkelName:           methodName,
		Description:        meta.Description,
		Deprecated:         meta.Deprecated,
		DeprecatedReason:   meta.DeprecatedReason,
		Auth:               model.AuthModeAuth,
		Arguments:          args,
		InputDescription:   inputDescription,
		ArgumentsSensitive: inputSensitive,
	}
	if len(args) > 0 {
		method.ArgumentsData = &model.Data{
			Name:    fmt.Sprintf("%sArguments", nameutil.ToCamel(method.Name)),
			Members: buildArgumentMembers(method.Arguments),
		}
	}
	return &model.ResourceCheck{
		Name:             gc.Name.Value,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Method:           method,
	}, valid
}

func newPermissionCodeArgument() *model.Argument {
	return &model.Argument{
		Name:   "code",
		Source: model.ArgumentSourcePermissionCode,
		Type:   &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString},
	}
}
