package compiler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skelc/internal/loader"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/source"
)

type Option struct {
	SkelIn      string
	SkelImports map[string]string
	Strict      bool
}

type Result struct {
	Domain        *model.Domain
	ImportAliases map[string]string
	Diagnostics   Diagnostics
}

// Compile loads, resolves, analyzes, and hashes one complete Skel input graph.
func Compile(option Option) (Result, error) { return CompileContext(context.Background(), option) }

func CompileContext(ctx context.Context, option Option) (Result, error) {
	return compileFrom(ctx, source.FileSystem{}, option, false)
}

// CompileImport allows unresolved dependencies for symbol and schema tooling.
func CompileImport(option Option) (Result, error) {
	return CompileImportContext(context.Background(), option)
}

func CompileImportContext(ctx context.Context, option Option) (Result, error) {
	return CompileImportFrom(ctx, source.FileSystem{}, option)
}

// CompileImportFrom uses the same pipeline for filesystem and frozen inputs.
func CompileImportFrom(ctx context.Context, provider source.Provider, option Option) (Result, error) {
	return compileFrom(ctx, provider, option, true)
}

func compileFrom(ctx context.Context, provider source.Provider, option Option, unresolved bool) (Result, error) {
	inputs := []Source{}
	diagnostics := Diagnostics{}
	if !unresolved {
		names := make([]string, 0, len(option.SkelImports))
		for name := range option.SkelImports {
			names = append(names, name)
		}
		slices.Sort(names)
		for _, name := range names {
			loaded, err := loader.LoadFrom(ctx, provider, option.SkelImports[name])
			if err != nil {
				return Result{}, err
			}
			sources, err := prepareInput(ctx, loaded, false)
			if err != nil {
				return Result{}, err
			}
			actual := sources[0].ExpectedDomain
			if actual != name {
				return Result{}, fmt.Errorf("skel import %s has domain %s", name, actual)
			}
			inputs = append(inputs, sources...)
			diagnostics = append(diagnostics, loaderWarningDiagnostics(loaded.Warnings)...)
		}
	}
	loaded, err := loader.LoadFrom(ctx, provider, option.SkelIn)
	if err != nil {
		return Result{}, err
	}
	sources, err := prepareInput(ctx, loaded, false)
	if err != nil {
		return Result{}, err
	}
	inputs = append(inputs, sources...)
	diagnostics = append(diagnostics, loaderWarningDiagnostics(loaded.Warnings)...)
	engine := NewWorkspaceAnalyzer()
	analyzed, domains, err := engine.AnalyzeWithOptionsContext(ctx, inputs, AnalysisOptions{AllowUnresolvedImports: unresolved, ResolveIsolatedImports: true, IncludeWarnings: true})
	if err != nil {
		return Result{}, err
	}
	// Preserve the CLI spelling while sharing cycle detection with editor analysis.
	for i := range analyzed {
		if analyzed[i].Code == DiagnosticCodeImportCycle {
			analyzed[i].Message = strings.Replace(analyzed[i].Message, "cyclic domain import", "cyclic skel import", 1)
		}
	}
	diagnostics = append(diagnostics, analyzed...)
	slices.SortFunc(diagnostics, compareDiagnostics)
	if diagnostics.HasErrors() {
		return Result{}, errors.Join(diagnostics.Errors()...)
	}
	if option.Strict {
		ApplyStrictMode(diagnostics)
		if diagnostics.HasErrors() {
			return Result{}, diagnostics
		}
	}
	for _, domain := range domains {
		if domain.Root == sources[0].Root && domain.Name == sources[0].ExpectedDomain {
			return Result{Domain: domain.Model, ImportAliases: domain.ImportAliases, Diagnostics: diagnostics}, nil
		}
	}
	return Result{}, fmt.Errorf("no complete domain in %s", option.SkelIn)
}

func loaderWarningDiagnostics(warnings []loader.Warning) Diagnostics {
	diagnostics := make(Diagnostics, 0, len(warnings))
	for _, warning := range warnings {
		position := model.Position{File: warning.Path, Line: 1, Column: 1}
		diagnostics = append(diagnostics, Diagnostic{
			Code: warning.Code, Severity: DiagnosticSeverityWarning, Position: position,
			Range: SourceRange{Start: position, End: position}, Message: warning.Message,
		})
	}
	return diagnostics
}

func appendAnalysisWarnings(diagnostics Diagnostics, warnings []string) Diagnostics {
	for _, warning := range warnings {
		diagnostics = append(diagnostics, Diagnostic{
			Code: DiagnosticCodeSemanticWarning, Severity: DiagnosticSeverityWarning, Message: warning,
		})
	}
	return diagnostics
}

func findDomainFile(sourceFiles []*loader.SourceFile) (*loader.SourceFile, error) {
	for _, sourceFile := range sourceFiles {
		if filepath.Base(sourceFile.FilePath) == loader.DomainFileName {
			return sourceFile, nil
		}
	}
	return nil, fmt.Errorf("%s not found", loader.DomainFileName)
}
