package compiler

import (
	"context"
	"path/filepath"
	"slices"

	"go.yorun.ai/skel/internal/loader"
)

// CheckResult contains the diagnostics produced by a check operation.
type CheckResult struct {
	Diagnostics Diagnostics
}

// Check validates all discoverable source files and returns independent
// diagnostics. Unresolved imports remain allowed for compatibility with the
// check command, which does not accept import path mappings.
func Check(option Option) (CheckResult, error) {
	return checkWithAnalyzer(option, NewWorkspaceAnalyzer())
}

func checkWithAnalyzer(option Option, workspaceAnalyzer *WorkspaceAnalyzer) (CheckResult, error) {
	return checkFrom(context.Background(), loader.FileSystem{}, option, workspaceAnalyzer)
}

// CheckFrom validates filesystem or frozen sources with unresolved imports.
func CheckFrom(ctx context.Context, provider loader.Provider, option Option) (CheckResult, error) {
	return checkFrom(ctx, provider, option, NewWorkspaceAnalyzer())
}

func checkFrom(ctx context.Context, provider loader.Provider, option Option, workspaceAnalyzer *WorkspaceAnalyzer) (CheckResult, error) {
	loadResult, err := loader.LoadFrom(ctx, provider, option.SkelIn)
	if err != nil {
		return CheckResult{}, err
	}
	return checkLoaded(ctx, loadResult, option, workspaceAnalyzer)
}

// CheckLoaded validates the exact source revisions returned by the loader.
func CheckLoaded(ctx context.Context, loaded loader.Result, option Option) (CheckResult, error) {
	return checkLoaded(ctx, loaded, option, NewWorkspaceAnalyzer())
}

func checkLoaded(ctx context.Context, loadResult loader.Result, option Option, workspaceAnalyzer *WorkspaceAnalyzer) (CheckResult, error) {
	sources, err := prepareInput(ctx, loadResult, true)
	if err != nil {
		return CheckResult{}, err
	}
	diagnostics, _, err := workspaceAnalyzer.AnalyzeWithOptionsContext(ctx, sources, AnalysisOptions{AllowUnresolvedImports: true})
	if err != nil {
		return CheckResult{}, err
	}
	filtered := append(Diagnostics{}, diagnostics...)
	filtered = append(filtered, LoaderWarningDiagnostics(loadResult.Warnings)...)
	if option.Strict {
		ApplyStrictMode(filtered)
	}
	slices.SortFunc(filtered, compareDiagnostics)
	return CheckResult{Diagnostics: filtered}, nil
}

func checkExpectedDomain(loadResult loader.Result, sources []Source) string {
	var domainSource *Source
	if loadResult.IsDir {
		for index := range sources {
			if filepath.Base(sources[index].Path) == loader.DomainFileName {
				domainSource = &sources[index]
				break
			}
		}
	} else if len(sources) > 0 {
		domainSource = &sources[0]
	}
	if domainSource == nil {
		return ""
	}
	content := domainSource.Parsed
	if len(domainSource.ParseDiagnostics) > 0 || content == nil || content.Domain == nil || content.Domain.Name == nil {
		return ""
	}
	return content.Domain.Name.String()
}
