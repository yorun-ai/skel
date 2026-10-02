// Package analysis runs cancellable semantic analysis for immutable LSP
// workspace snapshots.
package analysis

import (
	"context"
	"path/filepath"
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/loader"
	lspdiagnostic "go.yorun.ai/skelc/internal/lsp/diagnostic"
	"go.yorun.ai/skelc/internal/lsp/source"
	"go.yorun.ai/skelc/internal/lsp/workspace"
)

// SemanticSources converts indexed LSP documents into compiler sources.
func SemanticSources(documents map[uri.URI]*workspace.Document) ([]compiler.Source, map[string]uri.URI) {
	directoryInputs := map[uri.URI]string{}
	for _, document := range documents {
		if filepath.Base(document.Path) == loader.DomainFileName {
			directory, err := uri.JoinPath(document.URI, "..")
			if err == nil {
				directoryInputs[directory] = document.Domain
			}
		}
	}
	sources := make([]compiler.Source, 0, len(documents))
	paths := make(map[string]uri.URI, len(documents))
	for documentURI, document := range documents {
		path := document.AnalysisPath()
		root := path
		directory, _ := uri.JoinPath(document.URI, "..")
		expected, directoryInput := directoryInputs[directory]
		if directoryInput {
			root = string(directory)
			if document.URI.IsFile() {
				root = filepath.Dir(document.Path)
			}
		}
		sources = append(sources, compiler.Source{
			Path: path, Domain: document.Domain, Root: root, ExpectedDomain: expected, DirectoryInput: directoryInput,
			Document: document.Revision, Bindings: document.Bindings, Parsed: document.Parsed,
			ParseDiagnostics: document.ParseDiagnostics,
		})
		paths[filepath.Clean(path)] = documentURI
	}
	return sources, paths
}

// SemanticDiagnostics analyzes sources and converts compiler diagnostics to
// their LSP representation.
func SemanticDiagnostics(ctx context.Context, workspaceAnalyzer *compiler.WorkspaceAnalyzer, sources []compiler.Source, paths map[string]uri.URI) (map[uri.URI][]protocol.Diagnostic, error) {
	result, _, err := SemanticWorkspace(ctx, workspaceAnalyzer, sources, paths, false)
	return result, err
}

// SemanticWorkspace analyzes sources and returns both protocol diagnostics and
// every semantic domain that compiled successfully.
func SemanticWorkspace(
	ctx context.Context,
	workspaceAnalyzer *compiler.WorkspaceAnalyzer,
	sources []compiler.Source,
	paths map[string]uri.URI,
	strict bool,
) (map[uri.URI][]protocol.Diagnostic, []compiler.WorkspaceDomain, error) {
	result := map[uri.URI][]protocol.Diagnostic{}
	contents := newSourceBuffers(sources)
	diagnostics, domains, err := workspaceAnalyzer.AnalyzeDomainsContext(ctx, sources)
	if err != nil {
		return nil, nil, err
	}
	if strict {
		compiler.ApplyStrictMode(diagnostics)
	}
	for _, diagnostic := range diagnostics {
		if strings.HasPrefix(diagnostic.Code, "syntax.") {
			continue
		}
		documentURI, ok := paths[filepath.Clean(diagnostic.Position.File)]
		if !ok {
			continue
		}
		buffer := contents.get(diagnostic.Position.File)
		result[documentURI] = append(result[documentURI], lspdiagnostic.ToProtocolBuffer(diagnostic, buffer, func(path string) (uri.URI, source.Buffer, bool) {
			cleaned := filepath.Clean(path)
			relatedURI, exists := paths[cleaned]
			return relatedURI, contents.get(cleaned), exists
		}))
	}
	return result, domains, nil
}

// FilesystemDomain adapts source identities for schema baseline filesystem access.
func FilesystemDomain(domain compiler.WorkspaceDomain) compiler.WorkspaceDomain {
	if !strings.Contains(domain.Root, "://") {
		return domain
	}
	domain.Root = uri.URI(domain.Root).FsPath()
	domain.Sources = append([]compiler.Source{}, domain.Sources...)
	for i := range domain.Sources {
		domain.Sources[i].Path = uri.URI(domain.Sources[i].Path).FsPath()
	}
	return domain
}
