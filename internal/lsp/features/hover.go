package features

import (
	"context"
	"slices"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skel/internal/lsp/workspace"
	"go.yorun.ai/skel/internal/symbol"
)

func (s *Service) Hover(_ context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	snapshot := s.Snapshot
	document := snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return nil, nil
	}
	if occurrence, ok := snapshot.OccurrenceAt(document, params.Position); ok {
		definitions := snapshot.Definitions(snapshot.ResolveKey(document, occurrence.Key))
		if len(definitions) > 0 {
			definition := definitions[0].Definition
			name := definition.Key
			if symbol.ParseKey(name).Scope != "" {
				name = definition.Name
			}
			return hoverResult(occurrence.Range, definition.Detail, name, definition.Description), nil
		}
	}
	if symbol, ok := symbolAt(document.Symbols, params.Position); ok {
		return hoverResult(selectionRange(symbol), symbol.Detail, symbol.Name, symbol.Description), nil
	}
	for _, token := range document.Buffer.IdentifierTokens() {
		range_ := document.Buffer.Range(token.Start, token.End)
		if containsPosition(range_, params.Position) && slices.Contains(completionTypes, token.Value) {
			return hoverResult(range_, "built-in type", token.Value, ""), nil
		}
	}
	return nil, nil
}

func hoverResult(range_ protocol.Range, detail, name, description string) *protocol.Hover {
	if detail == "" {
		detail = "declaration"
	}
	value := "```skel\n" + detail + " " + name + "\n```"
	if description != "" {
		value += "\n\n" + description
	}
	return &protocol.Hover{Contents: &protocol.MarkupContent{Kind: protocol.MarkupKindMarkdown, Value: value}, Range: &range_}
}

func symbolAt(symbols []workspace.Symbol, position protocol.Position) (workspace.Symbol, bool) {
	for _, symbol := range symbols {
		selection := selectionRange(symbol)
		if containsPosition(selection, position) {
			return symbol, true
		}
		if child, ok := symbolAt(symbol.Children, position); ok {
			return child, true
		}
	}
	return workspace.Symbol{}, false
}
