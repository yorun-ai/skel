package compiler

import (
	"context"
	"errors"
	"path/filepath"

	"go.yorun.ai/skelc/internal/loader"
	"go.yorun.ai/skelc/internal/source"
)

// AnalysisOptions makes entry-point differences explicit. Recovery happens in
// the syntax stage; a recovered domain is never returned as a complete model.
type AnalysisOptions struct {
	AllowUnresolvedImports bool
	ResolveIsolatedImports bool
	IncludeWarnings        bool
}

// prepareInput normalizes one compiler input. Strict and recovering frontends
// share source identity, expected-domain discovery and directory validation.
func prepareInput(ctx context.Context, loaded loader.Result, recoverSyntax bool) ([]Source, error) {
	sources := make([]Source, 0, len(loaded.Files))
	for _, file := range loaded.Files {
		input := Source{Document: file.Document, Path: file.FilePath, Content: file.Content, DirectoryInput: loaded.IsDir}
		input.Root = file.FilePath
		if loaded.IsDir {
			input.Root = filepath.Dir(file.FilePath)
		}
		var err error
		if recoverSyntax {
			input.Parsed, input.ParseDiagnostics, err = ParseSourceRecoveringContext(ctx, file.FilePath, file.Content)
		} else {
			input.Parsed, err = parseContentContext(ctx, file, true)
		}
		if err != nil {
			return nil, err
		}
		sources = append(sources, input)
	}
	expected := checkExpectedDomain(loaded, sources)
	for i := range sources {
		sources[i].ExpectedDomain = expected
		sources[i].Domain = expected
		if loaded.IsDir && !recoverSyntax {
			if issue := inspectDirectorySource(sources[i].Path, expected, sources[i].Parsed); issue != nil {
				return nil, errors.New(issue.strictMessage)
			}
		}
	}
	return sources, nil
}

// AnalyzeInputFromContext compiles an immutable workspace input using the same
// parse, binding, validation and hashing stages as strict compilation.
func AnalyzeInputFromContext(ctx context.Context, provider source.Provider, path string, requireDomainFile bool) (Diagnostics, []WorkspaceDomain, error) {
	var loaded loader.Result
	var err error
	if requireDomainFile {
		loaded, err = loader.LoadFrom(ctx, provider, path)
	} else {
		loaded, err = loader.LoadWorkspaceFrom(ctx, provider, path)
	}
	if err != nil {
		return nil, nil, err
	}
	sources, err := prepareInput(ctx, loaded, true)
	if err != nil {
		return nil, nil, err
	}
	if !requireDomainFile && len(sources) > 0 && sources[0].ExpectedDomain == "" {
		for i := range sources {
			sources[i].DirectoryInput = false
		}
	}
	diagnostics, domains, err := NewWorkspaceAnalyzer().AnalyzeWithOptionsContext(ctx, sources, AnalysisOptions{AllowUnresolvedImports: true})
	return append(diagnostics, loaderWarningDiagnostics(loaded.Warnings)...), domains, err
}
