package features

import (
	"strings"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/lsp/index"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

type _DecoratorAllowance uint8

const (
	allowDesc _DecoratorAllowance = 1 << iota
	allowExample
	allowSensitive
	allowDeprecated
	allowIdentifier
)

type _DecoratorTarget struct {
	offset    int
	allowance _DecoratorAllowance
	existing  _DecoratorAllowance
}

func allowedDecoratorsAt(document *index.Document, position protocol.Position) []string {
	offset := document.Buffer.Offset(position)
	prefixStart := offset
	for prefixStart > 0 {
		r := document.Source[prefixStart-1]
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			break
		}
		prefixStart--
	}
	if prefixStart == 0 || document.Source[prefixStart-1] != '@' {
		return nil
	}

	decoratorStart := prefixStart - 1
	sanitized := document.Source[:decoratorStart] +
		strings.Repeat(" ", offset-decoratorStart) +
		document.Source[offset:]
	content, _ := compiler.ParseSourceRecovering(document.Path, []byte(sanitized))
	if content == nil {
		return nil
	}

	targets := collectDecoratorTargets(content, sanitized)
	var selected *_DecoratorTarget
	for index := range targets {
		target := &targets[index]
		if target.offset < offset || !decoratorGapOnly(sanitized[offset:target.offset]) {
			continue
		}
		if selected == nil || target.offset < selected.offset {
			selected = target
		}
	}
	if selected == nil {
		return nil
	}

	decorators := make([]string, 0, len(completionDecorators))
	for _, decorator := range completionDecorators {
		bit := decoratorBit(decorator)
		if bit != 0 &&
			selected.allowance&bit != 0 &&
			selected.existing&bit == 0 {
			decorators = append(decorators, decorator)
		}
	}
	return decorators
}

func collectDecoratorTargets(content *grammar.SkelContent, source string) []_DecoratorTarget {
	targets := make([]_DecoratorTarget, 0)
	add := func(offset int, allowance _DecoratorAllowance, decorators []*grammar.Decorator) {
		if offset >= 0 {
			targets = append(targets, _DecoratorTarget{
				offset: offset, allowance: allowance, existing: existingDecorators(decorators),
			})
		}
	}
	addKeyword := func(offset int, keyword string, allowance _DecoratorAllowance, decorators []*grammar.Decorator) {
		add(keywordOffsetBefore(source, offset, keyword), allowance, decorators)
	}
	addBlockKeyword := func(offset int, keyword string, allowance _DecoratorAllowance, decorators []*grammar.Decorator) {
		add(keywordOffsetAfterDecorators(source, offset, keyword), allowance, decorators)
	}
	addMembers := func(members []*grammar.DataMember) {
		for _, member := range members {
			add(identifierOffset(member.Name), allowDesc|allowExample|allowSensitive|allowDeprecated, member.Decorators)
		}
	}
	addArguments := func(arguments []*grammar.Argument) {
		for _, argument := range arguments {
			add(identifierOffset(argument.Name), allowDesc|allowExample|allowSensitive|allowDeprecated, argument.Decorators)
		}
	}

	if content.Domain != nil && content.Domain.Name != nil && len(content.Domain.Name.Parts) > 0 {
		addKeyword(identifierOffset(content.Domain.Name.Parts[0]), "domain", allowDesc, content.Domain.Decorators)
	}
	for _, entry := range content.Entries {
		switch {
		case entry.Enum != nil:
			addKeyword(identifierOffset(entry.Enum.Name), "enum", allowDesc|allowDeprecated, entry.Enum.Decorators)
			for _, item := range entry.Enum.Items {
				add(identifierOffset(item.Name), allowDesc|allowDeprecated, item.Decorators)
			}
		case entry.Data != nil:
			addKeyword(identifierOffset(entry.Data.Name), "data", allowDesc|allowSensitive|allowDeprecated, entry.Data.Decorators)
			addMembers(entry.Data.Members)
		case entry.Config != nil:
			addKeyword(identifierOffset(entry.Config.Name), "config", allowDesc|allowSensitive|allowDeprecated, entry.Config.Decorators)
			for _, member := range entry.Config.Members {
				add(identifierOffset(member.Name), allowDesc|allowExample|allowSensitive|allowDeprecated, member.Decorators)
			}
		case entry.Event != nil:
			addKeyword(identifierOffset(entry.Event.Name), "event", allowDesc|allowDeprecated, entry.Event.Decorators)
			if entry.Event.Payload != nil {
				addBlockKeyword(entry.Event.Payload.Pos.Offset, "payload", allowSensitive, entry.Event.Payload.Decorators)
				addMembers(entry.Event.Payload.Members)
			}
		case entry.Actor != nil:
			addKeyword(identifierOffset(entry.Actor.Name), "actor", allowDesc|allowDeprecated, entry.Actor.Decorators)
			for _, section := range entry.Actor.Sections {
				if section.Auth == nil {
					continue
				}
				if section.Auth.Credential != nil {
					addBlockKeyword(section.Auth.Credential.Pos.Offset, "credential", allowSensitive, section.Auth.Credential.Decorators)
					addMembers(section.Auth.Credential.Members)
				}
				if section.Auth.Info != nil {
					addBlockKeyword(section.Auth.Info.Pos.Offset, "info", allowSensitive, section.Auth.Info.Decorators)
					for _, member := range section.Auth.Info.Members {
						add(identifierOffset(member.Name), allowDesc|allowExample|allowSensitive|allowDeprecated|allowIdentifier, member.Decorators)
					}
				}
			}
		case entry.Resource != nil:
			addKeyword(identifierOffset(entry.Resource.Name), "resource", allowDesc|allowDeprecated, entry.Resource.Decorators)
			for _, action := range entry.Resource.Actions {
				addKeyword(identifierOffset(action.Name), "action", allowDesc|allowDeprecated, action.Decorators)
				for _, check := range action.Checks {
					addResourceCheckTarget(addKeyword, addBlockKeyword, addArguments, check)
				}
			}
			for _, check := range entry.Resource.Checks {
				addResourceCheckTarget(addKeyword, addBlockKeyword, addArguments, check)
			}
		case entry.Service != nil:
			addKeyword(identifierOffset(entry.Service.Name), "service", allowDesc|allowDeprecated, entry.Service.Decorators)
			methods := make([]_DecoratedMethod, 0)
			if len(entry.Service.Sections) > 0 {
				for _, section := range entry.Service.Sections {
					if section.Method != nil {
						decorators := append(append([]*grammar.Decorator{}, section.Decorators...), section.Method.Decorators...)
						methods = append(methods, _DecoratedMethod{method: section.Method, decorators: decorators})
					}
				}
			} else {
				for _, method := range entry.Service.Methods {
					methods = append(methods, _DecoratedMethod{method: method, decorators: method.Decorators})
				}
			}
			for _, decorated := range methods {
				addKeyword(identifierOffset(decorated.method.Name), "method", allowDesc|allowDeprecated, decorated.decorators)
				addMethodTargets(addBlockKeyword, addArguments, decorated.method.Input, decorated.method.Output)
			}
		case entry.Task != nil:
			addKeyword(identifierOffset(entry.Task.Name), "task", allowDesc|allowDeprecated, entry.Task.Decorators)
			for _, trigger := range entry.Task.Triggers {
				addKeyword(identifierOffset(trigger.Name), "trigger", allowDesc|allowDeprecated, trigger.Decorators)
				addMethodTargets(addBlockKeyword, addArguments, trigger.Input, nil)
			}
		case entry.Web != nil:
			addKeyword(identifierOffset(entry.Web.Name), "web", allowDesc|allowDeprecated, entry.Web.Decorators)
		}
	}
	return targets
}

type _DecoratedMethod struct {
	method     *grammar.Method
	decorators []*grammar.Decorator
}

func addResourceCheckTarget(
	addKeyword func(int, string, _DecoratorAllowance, []*grammar.Decorator),
	addBlockKeyword func(int, string, _DecoratorAllowance, []*grammar.Decorator),
	addArguments func([]*grammar.Argument),
	check *grammar.ResourceCheck,
) {
	addKeyword(identifierOffset(check.Name), "check", allowDesc|allowDeprecated, check.Decorators)
	addMethodTargets(addBlockKeyword, addArguments, check.Input, nil)
}

func addMethodTargets(
	addBlockKeyword func(int, string, _DecoratorAllowance, []*grammar.Decorator),
	addArguments func([]*grammar.Argument),
	input *grammar.MethodInput,
	output *grammar.MethodOutput,
) {
	if input != nil {
		addBlockKeyword(input.Pos.Offset, "input", allowDesc|allowSensitive, input.Decorators)
		addArguments(input.Arguments)
	}
	if output != nil {
		addBlockKeyword(output.Pos.Offset, "output", allowDesc|allowExample|allowSensitive, output.Decorators)
	}
}

func identifierOffset(identifier *grammar.Identifier) int {
	if identifier == nil {
		return -1
	}
	return identifier.Pos.Offset
}

func decoratorBit(decorator string) _DecoratorAllowance {
	switch decorator {
	case "desc":
		return allowDesc
	case "example":
		return allowExample
	case "identifier":
		return allowIdentifier
	case "sensitive":
		return allowSensitive
	case "deprecated":
		return allowDeprecated
	default:
		return 0
	}
}

func existingDecorators(decorators []*grammar.Decorator) _DecoratorAllowance {
	var existing _DecoratorAllowance
	for _, decorator := range decorators {
		if decorator != nil && decorator.Name != nil {
			existing |= decoratorBit(decorator.Name.Value)
		}
	}
	return existing
}
