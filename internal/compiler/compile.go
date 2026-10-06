package compiler

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/schema"
)

type Option struct {
	SkelIn      string
	SkelImports map[string]string
	Strict      bool
}

type Result struct {
	// Imports retains direct source declarations, including repeated imports
	// across files. Their Domain pointers are intentionally unset.
	Imports []*schema.Import
	Domain  *schema.Domain

	Diagnostics Diagnostics
}

// Compile loads, resolves, analyzes, and hashes one complete Skel input graph.
func Compile(option Option) (Result, error) { return CompileContext(context.Background(), option) }

func CompileContext(ctx context.Context, option Option) (Result, error) {
	return CompileFrom(ctx, loader.FileSystem{}, option)
}

// CompileFrom compiles a complete graph from the supplied immutable provider.
func CompileFrom(ctx context.Context, provider loader.Provider, option Option) (Result, error) {
	return compileFrom(ctx, provider, option, false)
}

// CompileImport allows unresolved dependencies for symbol and schema tooling.
func CompileImport(option Option) (Result, error) {
	return CompileImportContext(context.Background(), option)
}

func CompileImportContext(ctx context.Context, option Option) (Result, error) {
	return CompileImportFrom(ctx, loader.FileSystem{}, option)
}

// CompileImportFrom uses the same pipeline for filesystem and frozen inputs.
func CompileImportFrom(ctx context.Context, provider loader.Provider, option Option) (Result, error) {
	return compileFrom(ctx, provider, option, true)
}

func compileFrom(ctx context.Context, provider loader.Provider, option Option, unresolved bool) (Result, error) {
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
			diagnostics = append(diagnostics, LoaderWarningDiagnostics(loaded.Warnings)...)
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
	diagnostics = append(diagnostics, LoaderWarningDiagnostics(loaded.Warnings)...)
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
			imports := make([]*schema.Import, 0)
			for _, source := range sources {
				for _, declaration := range source.Parsed.Imports {
					item := &schema.Import{Name: declaration.Domain.String(), Pos: parser.SourcePosition(declaration.Pos)}
					if declaration.Alias != nil {
						item.Alias = declaration.Alias.Value
						item.ExplicitAlias = true
					}
					imports = append(imports, item)
				}
			}
			return Result{Domain: domain.Schema, Imports: imports, Diagnostics: diagnostics}, nil
		}
	}
	return Result{}, fmt.Errorf("no complete domain in %s", option.SkelIn)
}

// LoaderWarningDiagnostics converts source discovery warnings to diagnostics.
func LoaderWarningDiagnostics(warnings []loader.Warning) Diagnostics {
	diagnostics := make(Diagnostics, 0, len(warnings))
	for _, warning := range warnings {
		position := schema.Position{File: warning.Path, Line: 1, Column: 1}
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
