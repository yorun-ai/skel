package workspace

import (
	"slices"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skel/internal/lsp/source"
	"go.yorun.ai/skel/internal/symbol"
)

func indexOccurrences(document *Document) []Occurrence {
	return bindOccurrences(document, func(id symbol.SymbolID) []symbol.Symbol {
		result := []symbol.Symbol{}
		for _, declaration := range document.Bindings.Symbols {
			if declaration.ID == id {
				result = append(result, declaration)
			}
		}
		return result
	})
}

// bindOccurrences projects shared bindings into LSP ranges. The workspace
// supplies input-scoped lookup so references can resolve across source files.
func bindOccurrences(document *Document, lookup symbol.Lookup) []Occurrence {
	occurrences := []Occurrence{}
	seen := map[protocol.Range]bool{}
	add := func(id symbol.SymbolID, range_ protocol.Range, resolution symbol.Resolution) {
		if !seen[range_] {
			seen[range_] = true
			occurrences = append(occurrences, Occurrence{Key: id.Key(), Range: range_, Binding: resolution})
		}
	}
	for _, declaration := range document.Bindings.Symbols {
		add(declaration.ID, document.Buffer.Range(declaration.Span.Start, declaration.Span.End), symbol.Resolution{Target: declaration.ID, Kind: declaration.Kind, Status: symbol.Resolved})
	}
	for _, reference := range document.Bindings.References {
		resolved := symbol.Resolve(reference, document.Bindings.Imports, lookup)
		if resolved.Status != symbol.UnknownImport {
			add(resolved.Target, document.Buffer.Range(reference.Span.Start, reference.Span.End), resolved)
		}
	}
	slices.SortFunc(occurrences, func(a, b Occurrence) int { return source.ComparePosition(a.Range.Start, b.Range.Start) })
	return occurrences
}
