package features

import (
	"context"
	"slices"
	"strings"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/lsp/index"
)

var completionKeywords = []string{
	"actor", "action", "all", "any", "as", "auth", "check", "config", "credential", "data",
	"domain", "enum", "event", "for", "import", "info", "input", "method", "mount", "noauth", "output",
	"api", "open", "payload", "permission", "pub", "require", "resource", "service", "task", "trigger", "via", "web",
}

var completionDecorators = []string{"deprecated", "desc", "example", "identifier", "sensitive"}

var completionTypes = []string{
	"binary", "bool", "decimal", "duration", "float", "int", "json", "list", "localdate",
	"localdatetime", "localtime", "map", "string", "timestamp", "uuid",
}

var actorViaCompletionValues = []string{"agent", "client", "openapi"}
var configLifecycleCompletionValues = []string{"eternal", "instant"}

func (s *Service) Completion(_ context.Context, params *protocol.CompletionParams) (protocol.CompletionResult, error) {
	snapshot := s.Snapshot
	document := snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return protocol.CompletionItemSlice{}, nil
	}
	if document.Buffer.InNonCode(params.Position) {
		return protocol.CompletionItemSlice{}, nil
	}
	if prefix, range_, ok := decoratorPrefixBeforePositionBuffer(document.Buffer, params.Position); ok {
		decorators := allowedDecoratorsAt(document, params.Position)
		items := make(protocol.CompletionItemSlice, 0, len(decorators))
		for _, decorator := range decorators {
			if !strings.HasPrefix(decorator, prefix) {
				continue
			}
			item := protocol.CompletionItem{
				Label: "@" + decorator, Kind: protocol.CompletionItemKindKeyword,
				Detail:     protocol.NewOptional("Skel decorator"),
				FilterText: protocol.NewOptional(decorator),
				TextEdit:   &protocol.TextEdit{Range: range_, NewText: decorator},
			}
			if decorator == "deprecated" {
				item.TextEdit = &protocol.TextEdit{Range: range_, NewText: `deprecated("")`}
				if s.SnippetSupport {
					item.InsertTextFormat = protocol.InsertTextFormatSnippet
					item.TextEdit = &protocol.TextEdit{Range: range_, NewText: `deprecated("$0")`}
				}
			}
			items = append(items, item)
		}
		return items, nil
	}
	if values := completionValuesBeforePositionBuffer(document.Buffer, params.Position); len(values) > 0 {
		items := make(protocol.CompletionItemSlice, 0, len(values))
		for _, value := range values {
			items = append(items, protocol.CompletionItem{
				Label: value, Kind: protocol.CompletionItemKindValue,
				Detail: protocol.NewOptional("Skel value"),
			})
		}
		return items, nil
	}

	items := map[string]protocol.CompletionItem{}
	qualifier := qualifierBeforePositionBuffer(document.Buffer, params.Position)
	if qualifier != "" {
		domain := document.Imports[qualifier]
		for _, candidate := range snapshot.DocumentsFor(document, domain) {
			for _, definition := range candidate.Definitions {
				items[definition.Name] = symbolCompletion(definition, domain)
			}
		}
	} else {
		for _, keyword := range completionKeywords {
			items[keyword] = protocol.CompletionItem{
				Label: keyword, Kind: protocol.CompletionItemKindKeyword,
				Detail: protocol.NewOptional("Skel keyword"),
			}
		}
		for _, kind := range completionTypes {
			items[kind] = protocol.CompletionItem{
				Label: kind, Kind: protocol.CompletionItemKindClass,
				Detail: protocol.NewOptional("Skel built-in type"),
			}
		}
		for alias, domain := range document.Imports {
			items[alias] = protocol.CompletionItem{
				Label: alias, Kind: protocol.CompletionItemKindModule,
				Detail: protocol.NewOptional(domain),
			}
		}
		for _, candidate := range snapshot.DocumentsFor(document, document.Domain) {
			for _, definition := range candidate.Definitions {
				items[definition.Name] = symbolCompletion(definition, document.Domain)
			}
		}
	}

	labels := make([]string, 0, len(items))
	for label := range items {
		labels = append(labels, label)
	}
	slices.Sort(labels)
	result := make(protocol.CompletionItemSlice, 0, len(labels))
	for _, label := range labels {
		result = append(result, items[label])
	}
	return result, nil
}

func symbolCompletion(definition index.Definition, domain string) protocol.CompletionItem {
	item := protocol.CompletionItem{
		Label: definition.Name, Kind: completionKind(definition.Kind), Detail: protocol.NewOptional(domain + "." + definition.Name),
	}
	if definition.Description != "" {
		item.Documentation = protocol.String(definition.Description)
	}
	if definition.Deprecated {
		item.Tags = []protocol.CompletionItemTag{protocol.CompletionItemTagDeprecated}
	}
	return item
}

func completionKind(kind protocol.SymbolKind) protocol.CompletionItemKind {
	switch kind {
	case protocol.SymbolKindEnum:
		return protocol.CompletionItemKindEnum
	case protocol.SymbolKindInterface:
		return protocol.CompletionItemKindInterface
	case protocol.SymbolKindEvent:
		return protocol.CompletionItemKindEvent
	case protocol.SymbolKindFunction:
		return protocol.CompletionItemKindFunction
	case protocol.SymbolKindStruct, protocol.SymbolKindObject:
		return protocol.CompletionItemKindStruct
	default:
		return protocol.CompletionItemKindReference
	}
}
