package workspace

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/binding"
	"go.yorun.ai/skelc/internal/lsp/source"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

func entryDefinition(entry *grammar.SkelEntry) (string, lexer.Position, protocol.SymbolKind, string) {
	identifier, kind := binding.Declaration(entry)
	if identifier == nil {
		return "", lexer.Position{}, protocol.SymbolKindNull, ""
	}
	protocolKind, detail := protocol.SymbolKindNull, ""
	switch kind {
	case binding.Enum:
		protocolKind, detail = protocol.SymbolKindEnum, "enum"
	case binding.Data:
		protocolKind, detail = protocol.SymbolKindStruct, "data"
	case binding.Config:
		protocolKind, detail = protocol.SymbolKindStruct, "config"
	case binding.Actor:
		protocolKind, detail = protocol.SymbolKindInterface, "actor"
	case binding.Resource:
		protocolKind, detail = protocol.SymbolKindObject, "resource"
	case binding.Service:
		protocolKind, detail = protocol.SymbolKindInterface, "service"
		if entry.Service.Open {
			detail = "open service"
		}
		if entry.Service.Api {
			detail = "api service"
		}
	case binding.Web:
		protocolKind, detail = protocol.SymbolKindInterface, "web"
	case binding.Event:
		protocolKind, detail = protocol.SymbolKindEvent, "event"
	case binding.Task:
		protocolKind, detail = protocol.SymbolKindFunction, "task"
	}
	return identifier.Value, identifier.Pos, protocolKind, detail
}

func entrySymbol(source source.Buffer, entry *grammar.SkelEntry, name, detail, description string, deprecated bool, kind protocol.SymbolKind, range_ protocol.Range) Symbol {
	children := []Symbol{}
	switch {
	case entry.Enum != nil:
		for _, item := range entry.Enum.Items {
			children = append(children, newDecoratedSymbol(source, item.Name, "enum item", item.Decorators, protocol.SymbolKindEnumMember, nil))
		}
	case entry.Data != nil:
		children = dataMemberSymbols(source, entry.Data.Members)
	case entry.Config != nil:
		children = dataMemberSymbols(source, entry.Config.Members)
	case entry.Actor != nil:
		for _, via := range entry.Actor.Vias {
			children = append(children, newSymbol(source, via.Name, "actor transport", "", false, protocol.SymbolKindInterface, nil))
		}
		for _, section := range entry.Actor.Sections {
			if section.Auth == nil {
				continue
			}
			if section.Auth.Credential != nil {
				children = append(children, sectionSymbol(source, "credential", section.Auth.Credential.Pos, dataMemberSymbols(source, section.Auth.Credential.Members)))
			}
			if section.Auth.Info != nil {
				children = append(children, sectionSymbol(source, "info", section.Auth.Info.Pos, dataMemberSymbols(source, section.Auth.Info.Members)))
			}
		}
	case entry.Resource != nil:
		for _, section := range entry.Resource.Sections {
			if section.Check != nil {
				children = append(children, resourceCheckSymbol(source, section.Check))
			}
			if section.Action != nil {
				actionChildren := make([]Symbol, 0, len(section.Action.Checks))
				for _, check := range section.Action.Checks {
					actionChildren = append(actionChildren, resourceCheckSymbol(source, check))
				}
				children = append(children, newDecoratedSymbol(source, section.Action.Name, "action", section.Action.Decorators, protocol.SymbolKindMethod, actionChildren))
			}
		}
	case entry.Service != nil:
		for _, section := range entry.Service.Sections {
			if section.Method == nil {
				continue
			}
			method := section.Method
			methodChildren := []Symbol{}
			if method.Input != nil {
				methodChildren = argumentSymbols(source, method.Input.Arguments)
			}
			methodDescription, methodDeprecated := documentationFromDecoratorGroups(section.Decorators, method.Decorators)
			children = append(children, newSymbol(source, method.Name, "method", methodDescription, methodDeprecated, protocol.SymbolKindMethod, methodChildren))
		}
	case entry.Event != nil:
		if entry.Event.Payload != nil {
			children = dataMemberSymbols(source, entry.Event.Payload.Members)
		}
	case entry.Task != nil:
		for _, trigger := range entry.Task.Triggers {
			triggerChildren := []Symbol{}
			if trigger.Input != nil {
				triggerChildren = argumentSymbols(source, trigger.Input.Arguments)
			}
			children = append(children, newDecoratedSymbol(source, trigger.Name, "trigger", trigger.Decorators, protocol.SymbolKindEvent, triggerChildren))
		}
	}
	return finishSymbol(Symbol{Name: name, Detail: detail, Description: description, Deprecated: deprecated, Kind: kind, Range: range_, Children: children})
}

func dataMemberSymbols(source source.Buffer, members []*grammar.DataMember) []Symbol {
	symbols := make([]Symbol, 0, len(members))
	for _, member := range members {
		symbols = append(symbols, newDecoratedSymbol(source, member.Name, "field", member.Decorators, protocol.SymbolKindField, nil))
	}
	return symbols
}

func argumentSymbols(source source.Buffer, arguments []*grammar.Argument) []Symbol {
	symbols := make([]Symbol, 0, len(arguments))
	for _, argument := range arguments {
		symbols = append(symbols, newDecoratedSymbol(source, argument.Name, "parameter", argument.Decorators, protocol.SymbolKindVariable, nil))
	}
	return symbols
}

func resourceCheckSymbol(source source.Buffer, check *grammar.ResourceCheck) Symbol {
	var arguments []*grammar.Argument
	if check.Input != nil {
		arguments = check.Input.Arguments
	}
	children := argumentSymbols(source, arguments)
	return newDecoratedSymbol(source, check.Name, "check", check.Decorators, protocol.SymbolKindFunction, children)
}

func newDecoratedSymbol(source source.Buffer, name *grammar.Identifier, detail string, decorators []*grammar.Decorator, kind protocol.SymbolKind, children []Symbol) Symbol {
	description, deprecated := documentationFromDecoratorGroups(decorators)
	return newSymbol(source, name, detail, description, deprecated, kind, children)
}

func newSymbol(source source.Buffer, name *grammar.Identifier, detail, description string, deprecated bool, kind protocol.SymbolKind, children []Symbol) Symbol {
	range_ := identifierRange(source, name.Pos, name.Value)
	return finishSymbol(Symbol{Name: name.Value, Detail: detail, Description: description, Deprecated: deprecated, Kind: kind, Range: range_, Children: children})
}

func sectionSymbol(source source.Buffer, name string, pos lexer.Position, children []Symbol) Symbol {
	range_ := identifierRange(source, pos, name)
	return finishSymbol(Symbol{Name: name, Detail: "actor auth section", Kind: protocol.SymbolKindObject, Range: range_, Children: children})
}

func finishSymbol(symbol Symbol) Symbol {
	for _, child := range symbol.Children {
		if source.ComparePosition(child.Range.End, symbol.Range.End) > 0 {
			symbol.Range.End = child.Range.End
		}
	}
	return symbol
}

func documentationFromDecoratorGroups(groups ...[]*grammar.Decorator) (string, bool) {
	description := ""
	deprecatedReason := ""
	for _, decorators := range groups {
		for _, decorator := range decorators {
			if decorator.Value == nil {
				continue
			}
			switch decorator.Name.Value {
			case "desc":
				if description == "" {
					description, _ = grammar.UnquoteDescriptionString(decorator.Value.Raw)
				}
			case "deprecated":
				if deprecatedReason == "" {
					deprecatedReason, _ = grammar.UnquoteDescriptionString(decorator.Value.Raw)
				}
			}
		}
	}
	parts := make([]string, 0, 2)
	if description != "" {
		parts = append(parts, description)
	}
	if deprecatedReason != "" {
		parts = append(parts, "Deprecated: "+deprecatedReason)
	}
	return strings.Join(parts, "\n\n"), deprecatedReason != ""
}
