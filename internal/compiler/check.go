package compiler

import (
	"context"
	"path/filepath"
	"slices"

	"go.yorun.ai/skelc/internal/loader"
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
	loadResult, err := loader.Load(option.SkelIn)
	if err != nil {
		return CheckResult{}, err
	}
	sources, err := prepareInput(context.Background(), loadResult, true)
	if err != nil {
		return CheckResult{}, err
	}
	diagnostics, _, err := workspaceAnalyzer.AnalyzeWithOptionsContext(context.Background(), sources, AnalysisOptions{AllowUnresolvedImports: true})
	if err != nil {
		return CheckResult{}, err
	}
	filtered := append(Diagnostics{}, diagnostics...)
	filtered = append(filtered, loaderWarningDiagnostics(loadResult.Warnings)...)
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
