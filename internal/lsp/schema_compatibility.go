package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/lsp/analysis"
	"go.yorun.ai/skel/internal/lsp/features"
	"go.yorun.ai/skel/internal/lsp/workspace"
	"go.yorun.ai/skel/internal/schema"
	"go.yorun.ai/skel/internal/sourcediff"
)

const (
	commandSchemaDiff = "skel.schema.diff"
)

func (s *_Server) CodeLens(_ context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	s.mu.RLock()
	enabled := s.schemaCompatibility.CodeLens
	s.mu.RUnlock()
	if !enabled {
		return []protocol.CodeLens{}, nil
	}
	if params.TextDocument.URI.Scheme() == "untitled" {
		return []protocol.CodeLens{}, nil
	}
	document := s.workspace.Snapshot().Document(params.TextDocument.URI)
	if document == nil || document.Domain == "" {
		return []protocol.CodeLens{}, nil
	}
	range_ := protocol.Range{}
	if document.Parsed != nil && document.Parsed.Domain != nil && document.Parsed.Domain.Name != nil {
		position := document.Parsed.Domain.Name.Pos
		range_ = document.Buffer.IdentifierRange(position.Line, position.Column, document.Parsed.Domain.Name.String())
	}
	argument, _ := json.Marshal(string(params.TextDocument.URI))
	return []protocol.CodeLens{{
		Range: range_,
		Command: protocol.Command{
			Title: "Check schema compatibility", Command: features.CommandShowSchemaCompatibility,
			Arguments: []protocol.LSPAny{argument},
		},
	}}, nil
}

func (s *_Server) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (protocol.LSPAny, error) {
	if params.Command != commandSchemaDiff {
		return nil, fmt.Errorf("unsupported command: %s", params.Command)
	}
	if len(params.Arguments) != 1 {
		return nil, fmt.Errorf("%s requires one document URI", commandSchemaDiff)
	}
	var rawURI string
	if err := json.Unmarshal(params.Arguments[0], &rawURI); err != nil || rawURI == "" {
		return nil, fmt.Errorf("%s requires a valid document URI", commandSchemaDiff)
	}
	report, err := diffDocument(ctx, s.workspace.Snapshot(), uri.URI(rawURI), s.analysisOptions())
	if err != nil {
		return nil, err
	}
	content, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("encode schema compatibility report: %w", err)
	}
	return protocol.LSPAny(content), nil
}

func diffDocument(
	ctx context.Context,
	snapshot workspace.Snapshot,
	documentURI uri.URI,
	option analysis.Options,
) (*schema.Report, error) {
	document := snapshot.Document(documentURI)
	if document == nil {
		return nil, fmt.Errorf("Skel document is not part of the workspace: %s", documentURI)
	}
	if document.Domain == "" {
		return nil, fmt.Errorf("cannot determine the Skel domain for %s", documentURI)
	}
	sources, _ := analysis.SemanticSources(snapshot.DocumentsMap())
	analyzer := compiler.NewWorkspaceAnalyzer()
	diagnostics, domains, err := analyzer.AnalyzeDomainsContext(ctx, sources)
	if err != nil {
		return nil, err
	}
	for _, domain := range domains {
		for _, candidate := range domain.Sources {
			if filepath.Clean(candidate.Path) == filepath.Clean(document.AnalysisPath()) {
				return sourcediff.DiffWorkspaceDomain(ctx, analysis.FilesystemDomain(domain), sourcediff.Option{BaselineSkelIn: option.Compatibility.BaselineSkelIn, Strict: option.Strict})
			}
		}
	}
	for _, diagnostic := range diagnostics {
		if filepath.Clean(diagnostic.Position.File) == filepath.Clean(document.AnalysisPath()) {
			return nil, fmt.Errorf("cannot compare an invalid Skel domain: %s", diagnostic.Message)
		}
	}
	return nil, fmt.Errorf("cannot build schema for domain %s", document.Domain)
}
