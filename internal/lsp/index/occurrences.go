package index

import (
	"slices"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/binding"
	"go.yorun.ai/skelc/internal/lsp/source"
)

func indexOccurrences(document *Document) []Occurrence {
	return BindOccurrences(document, func(id binding.SymbolID) []binding.Symbol {
		result := []binding.Symbol{}
		for _, symbol := range document.Bindings.Symbols {
			if symbol.ID == id {
				result = append(result, symbol)
			}
		}
		return result
	})
}

// BindOccurrences projects shared bindings into LSP ranges. The workspace
// supplies input-scoped lookup so references can resolve across source files.
func BindOccurrences(document *Document, lookup binding.Lookup) []Occurrence {
	occurrences := []Occurrence{}
	seen := map[protocol.Range]bool{}
	add := func(id binding.SymbolID, range_ protocol.Range, resolution binding.Resolution) {
		if !seen[range_] {
			seen[range_] = true
			occurrences = append(occurrences, Occurrence{Key: id.Key(), Range: range_, Binding: resolution})
		}
	}
	for _, symbol := range document.Bindings.Symbols {
		add(symbol.ID, document.Buffer.Range(symbol.Span.Start, symbol.Span.End), binding.Resolution{Target: symbol.ID, Kind: symbol.Kind, Status: binding.Resolved})
	}
	for _, reference := range document.Bindings.References {
		resolved := binding.Resolve(reference, document.Bindings.Imports, lookup)
		if resolved.Status != binding.UnknownImport {
			add(resolved.Target, document.Buffer.Range(reference.Span.Start, reference.Span.End), resolved)
		}
	}
	slices.SortFunc(occurrences, func(a, b Occurrence) int { return source.ComparePosition(a.Range.Start, b.Range.Start) })
	return occurrences
}
