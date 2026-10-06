package analyzer

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
)

func buildArgumentMembers(args []*schema.Argument) []*schema.DataMember {
	return sliceutil.Map(args, func(arg *schema.Argument) *schema.DataMember {
		return &schema.DataMember{
			Name:             arg.Name,
			Description:      arg.Description,
			Deprecated:       arg.Deprecated,
			DeprecatedReason: arg.DeprecatedReason,
			Example:          arg.Example,
			Sensitive:        arg.Sensitive,
			Type:             arg.Type,
		}
	})
}

func parseMethods(reporter *_DiagnosticReporter, owner *grammar.Identifier, methods []*grammar.Method) ([]*schema.Method, bool) {
	parsedMethods := make([]*schema.Method, 0, len(methods))
	methodPos := map[string]lexer.Position{}
	valid := true

	for _, grammarMethod := range methods {
		if reporter.cancelled() {
			break
		}
		method, methodValid := parseMethod(reporter, grammarMethod)
		valid = methodValid && valid
		duplicatedPosition, duplicated := methodPos[method.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated method %s found, also present at %s", method.Pos, method.Name, duplicatedPosition)
			valid = false
			continue
		}
		if method.ArgumentsData != nil {
			method.ArgumentsData.Name = fmt.Sprintf("%s%s", owner.Value, method.ArgumentsData.Name)
		}
		methodPos[method.Name] = lexer.Position{Filename: method.Pos.File, Line: method.Pos.Line, Column: method.Pos.Column}
		parsedMethods = append(parsedMethods, method)
	}

	valid = reporter.check(len(parsedMethods) > 0, "%s missing method for %s", owner.Pos, owner.Value) && valid
	return parsedMethods, valid
}

func parseMethod(reporter *_DiagnosticReporter, gm *grammar.Method) (*schema.Method, bool) {
	valid := checkCase(reporter, "Method", caseTypeLowerCamel, gm.Name)
	meta, metaValid := parseDecoratorMeta(reporter, gm.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	require, requireValid := parseRequire(reporter, gm.Require)
	valid = requireValid && valid
	authMode, authModeValid := parseAuthMode(reporter, methodAuthMarker(gm), schema.AuthModeUnset)
	valid = authModeValid && valid
	method := &schema.Method{
		Pos:              position(gm.Name.Pos),
		Name:             gm.Name.Value,
		SkelName:         gm.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Auth:             authMode,
		AuthPos:          authMarkerPosition(methodAuthMarker(gm)),
		Require:          require,
		Arguments:        []*schema.Argument{},
	}
	input := methodInput(gm)
	output := methodOutput(gm)

	if input != nil {
		inputMeta, inputValid := parseDecoratorMeta(reporter, input.Decorators, _DecoratorContext{
			allowDesc:      true,
			allowSensitive: true,
		})
		valid = inputValid && valid
		method.InputDescription = inputMeta.Description
		method.ArgumentsSensitive = inputMeta.Sensitive
		argPos := map[string]lexer.Position{}
		for _, grammarArgument := range input.Arguments {
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
			method.Arguments = append(method.Arguments, arg)
		}
	}
	if output != nil {
		outputMeta, outputValid := parseDecoratorMeta(reporter, output.Decorators, _DecoratorContext{
			allowDesc:      true,
			allowExample:   true,
			allowSensitive: true,
			requireDesc:    true,
		})
		valid = outputValid && valid
		method.OutputDescription = outputMeta.Description
		method.ResultSensitive = outputMeta.Sensitive
		if outputMeta.HasExample {
			method.OutputExample = outputMeta.Example
		}
		method.ResultType, outputValid = parseType(reporter, output.Type)
		valid = outputValid && valid
	}

	if len(method.Arguments) > 0 {
		method.ArgumentsData = &schema.Data{
			Name:    fmt.Sprintf("%sArguments", nameutil.ToCamel(method.Name)),
			Members: buildArgumentMembers(method.Arguments),
		}
	}

	return method, valid
}

func methodAuthMarker(gm *grammar.Method) *grammar.AuthMarker {
	return gm.Auth
}

func methodInput(gm *grammar.Method) *grammar.MethodInput {
	return gm.Input
}

func methodOutput(gm *grammar.Method) *grammar.MethodOutput {
	return gm.Output
}

func parseArgument(reporter *_DiagnosticReporter, ga *grammar.Argument) (*schema.Argument, bool) {
	valid := checkCase(reporter, "Argument", caseTypeLowerCamel, ga.Name)
	meta, metaValid := parseDecoratorMeta(reporter, ga.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowExample:    true,
		allowSensitive:  true,
		allowDeprecated: true,
		requireDesc:     true,
	})
	valid = metaValid && valid
	argType, typeValid := parseType(reporter, ga.Type)
	valid = typeValid && valid
	arg := &schema.Argument{
		Pos:              position(ga.Name.Pos),
		Name:             ga.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
		Sensitive:        meta.Sensitive,
		Type:             argType,
	}
	if meta.HasExample {
		arg.Example = meta.Example
	}
	return arg, valid
}
