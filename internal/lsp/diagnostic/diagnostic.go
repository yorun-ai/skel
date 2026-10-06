// Package diagnostic converts skelc diagnostics to Language Server Protocol diagnostics.
package diagnostic

import (
	"encoding/json"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	skeldiagnostic "go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/lsp/source"
)

// RelatedSource resolves a related diagnostic path to its document and content.
type RelatedSource func(path string) (uri.URI, string, bool)

// ToProtocol converts one skelc diagnostic and all resolvable related locations.
func ToProtocol(item skeldiagnostic.Diagnostic, content string, resolve RelatedSource) protocol.Diagnostic {
	var resolver RelatedBuffer
	if resolve != nil {
		resolver = func(path string) (uri.URI, source.Buffer, bool) {
			u, text, ok := resolve(path)
			return u, source.New(text), ok
		}
	}
	return ToProtocolBuffer(item, source.New(content), resolver)
}

// RelatedBuffer resolves related locations without rebuilding their line indexes.
type RelatedBuffer func(path string) (uri.URI, source.Buffer, bool)

// ToProtocolBuffer reuses indexes for the primary and related source ranges.
func ToProtocolBuffer(item skeldiagnostic.Diagnostic, content source.Buffer, resolve RelatedBuffer) protocol.Diagnostic {
	related := make([]protocol.DiagnosticRelatedInformation, 0, len(item.Related))
	if resolve != nil {
		for _, information := range item.Related {
			relatedURI, relatedContent, ok := resolve(information.Range.Start.File)
			if !ok {
				continue
			}
			related = append(related, protocol.DiagnosticRelatedInformation{
				Location: protocol.Location{URI: relatedURI, Range: RangeBuffer(relatedContent, information.Range)},
				Message:  information.Message,
			})
		}
	}
	return protocol.Diagnostic{
		Range: RangeBuffer(content, item.Range), Severity: Severity(item.Severity), RelatedInformation: related,
		Code: protocol.String(item.Code), Source: protocol.NewOptional("skelc"),
		Message: protocol.String(item.Message), Data: SuggestionData(item.Suggestion),
	}
}

// Range converts a one-based Skel source range to an LSP UTF-16 range.
func Range(content string, sourceRange skeldiagnostic.SourceRange) protocol.Range {
	return RangeBuffer(source.New(content), sourceRange)
}

// RangeBuffer converts a source range using an existing line index.
func RangeBuffer(buffer source.Buffer, sourceRange skeldiagnostic.SourceRange) protocol.Range {
	start := buffer.IdentifierRange(sourceRange.Start.Line, sourceRange.Start.Column, "").Start
	end := buffer.IdentifierRange(sourceRange.End.Line, sourceRange.End.Column, "").Start
	if source.ComparePosition(end, start) <= 0 {
		end = start
		end.Character++
	}
	return protocol.Range{Start: start, End: end}
}

// Severity converts skelc severity to LSP severity.
func Severity(severity skeldiagnostic.Severity) protocol.DiagnosticSeverity {
	if severity == skeldiagnostic.SeverityWarning {
		return protocol.DiagnosticSeverityWarning
	}
	return protocol.DiagnosticSeverityError
}

// SuggestionData encodes a skelc suggestion as LSP diagnostic data.
func SuggestionData(suggestion *skeldiagnostic.Suggestion) protocol.LSPAny {
	if suggestion == nil {
		return nil
	}
	content, err := json.Marshal(suggestion)
	if err != nil {
		return nil
	}
	return protocol.LSPAny(content)
}
