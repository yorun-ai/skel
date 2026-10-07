package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/lsp/source"
	"go.yorun.ai/skel/internal/sourcediff"
	"go.yorun.ai/skel/schema"
	"go.yorun.ai/skel/schema/diff"
)

type _CompatibilityDiagnosticData struct {
	Impact diff.ImpactLevel `json:"impact"`
	Change diff.ChangeType  `json:"change"`
	Symbol string           `json:"symbol"`
}

func appendCompatibilityDiagnostics(
	ctx context.Context,
	differ *sourcediff.Differ,
	diagnostics map[uri.URI][]protocol.Diagnostic,
	domains []compiler.WorkspaceDomain,
	sources []compiler.Source,
	paths map[string]uri.URI,
	option CompatibilityOptions,
) {
	contentByPath := newSourceBuffers(sources)
	for _, domain := range domains {
		if ctx.Err() != nil {
			return
		}
		fallback := domainFallback(domain, paths, contentByPath)
		report, err := differ.DiffWorkspaceDomain(ctx, FilesystemDomain(domain), sourcediff.Option{BaselineSkelIn: option.BaselineSkelIn})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if errors.Is(err, sourcediff.ErrGitHistoryUnavailable) {
				continue
			}
			if fallback.URI != "" {
				diagnostics[fallback.URI] = append(diagnostics[fallback.URI], protocol.Diagnostic{
					Range: fallback.Range, Severity: protocol.DiagnosticSeverityWarning, Code: protocol.String("schema.baseline"),
					Source: protocol.NewOptional("skelc schema"), Message: protocol.String(err.Error()),
				})
			}
			continue
		}
		for _, change := range report.Changes {
			if change.Impact == diff.ImpactCompatible && !option.IncludeCompatible {
				continue
			}
			documentURI, range_ := compatibilityLocation(change.Candidate, paths, contentByPath, fallback)
			if documentURI == "" {
				continue
			}
			data, _ := json.Marshal(_CompatibilityDiagnosticData{Impact: change.Impact, Change: change.Change, Symbol: change.Symbol})
			diagnostics[documentURI] = append(diagnostics[documentURI], protocol.Diagnostic{
				Range: range_, Severity: compatibilitySeverity(change.Impact), Code: protocol.String("schema." + change.Code),
				Source: protocol.NewOptional("skelc schema"), Message: protocol.String("[" + string(change.Impact) + "] " + change.Message),
				Data: protocol.LSPAny(data),
			})
		}
	}
}

type _CompatibilityLocation struct {
	URI   uri.URI
	Range protocol.Range
}

func domainFallback(domain compiler.WorkspaceDomain, paths map[string]uri.URI, contents *_SourceBuffers) _CompatibilityLocation {
	for _, candidate := range domain.Sources {
		path := filepath.Clean(candidate.Path)
		if documentURI, ok := paths[path]; ok {
			position := schema.Position{Line: 1, Column: 1}
			if candidate.Parsed != nil && candidate.Parsed.Domain != nil && candidate.Parsed.Domain.Name != nil {
				position.Line = candidate.Parsed.Domain.Name.Pos.Line
				position.Column = candidate.Parsed.Domain.Name.Pos.Column
			}
			return _CompatibilityLocation{URI: documentURI, Range: positionRange(contents.get(path), position)}
		}
	}
	return _CompatibilityLocation{}
}

func compatibilityLocation(
	position *schema.Position,
	paths map[string]uri.URI,
	contents *_SourceBuffers,
	fallback _CompatibilityLocation,
) (uri.URI, protocol.Range) {
	if position == nil || position.File == "" {
		return fallback.URI, fallback.Range
	}
	path := filepath.Clean(position.File)
	documentURI, ok := paths[path]
	if !ok {
		return fallback.URI, fallback.Range
	}
	return documentURI, positionRange(contents.get(path), *position)
}

func positionRange(content source.Buffer, position schema.Position) protocol.Range {
	range_ := content.IdentifierRange(position.Line, position.Column, "")
	range_.End.Character++
	return range_
}

func compatibilitySeverity(impact diff.ImpactLevel) protocol.DiagnosticSeverity {
	switch impact {
	case diff.ImpactBreaking:
		return protocol.DiagnosticSeverityWarning
	case diff.ImpactDangerous:
		return protocol.DiagnosticSeverityInformation
	default:
		return protocol.DiagnosticSeverityHint
	}
}
