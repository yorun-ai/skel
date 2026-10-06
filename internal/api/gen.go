package api

import (
	"context"
	"fmt"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/codegen/golang"
	"go.yorun.ai/skel/internal/codegen/skeleton"
	"go.yorun.ai/skel/internal/codegen/typescript"
	"go.yorun.ai/skel/model"
)

// MinimumGolangVineVersion is the minimum Vine module version supported by
// generated Go code.
const MinimumGolangVineVersion = golang.MinimumVineVersion

// DefaultGolangVineVersion is the Vine module version used for generated Go
// modules when GolangOption.VineVersion is empty.
const DefaultGolangVineVersion = golang.DefaultVineVersion

// CompileResult contains structured non-fatal diagnostics produced while loading and parsing Skel sources.
type CompileResult struct {
	// Diagnostics contains non-fatal diagnostics produced while parsing.
	Diagnostics []diagnostic.Diagnostic
}

// ApiFilter selects API services and optional local type roots for pruning.
type ApiFilter = common.ApiFilter

// GolangOption configures Go generation.
type GolangOption = golang.Option

// TypeScriptOption configures TypeScript generation.
type TypeScriptOption = typescript.RequestOption

// SkeletonOption configures Skel source generation.
type SkeletonOption = skeleton.Option

// GenerateGolang generates Go source or a standalone Go module from a parsed
// domain. Stale files carrying the skelc generated marker may be removed.
func GenerateGolang(domain *model.Domain, option GolangOption) error {
	if domain == nil {
		return fmt.Errorf("parsed domain is required")
	}
	codegenOption, err := golang.NormalizeOption(option)
	if err != nil {
		return err
	}
	return golang.GenerateManaged(domain, codegenOption)
}

// CompileGolang parses input and generates Go source or a standalone Go module.
// Parsing completes before any generated output is committed.
func CompileGolang(input Input, option GolangOption) (CompileResult, error) {
	compilerOption, err := normalizeInput(input)
	if err != nil {
		return CompileResult{}, err
	}
	codegenOption, err := golang.NormalizeOption(option)
	if err != nil {
		return CompileResult{}, err
	}
	parsed, err := compileInput(context.Background(), input, compilerOption, false)
	if err != nil {
		return CompileResult{}, err
	}
	if err := golang.GenerateManaged(parsed.Domain, codegenOption); err != nil {
		return CompileResult{}, err
	}
	return CompileResult{Diagnostics: parsed.Diagnostics}, nil
}

// GenerateTypeScript generates TypeScript source from a parsed domain. Stale
// files carrying the skelc generated marker may be removed.
func GenerateTypeScript(domain *model.Domain, option TypeScriptOption) error {
	if domain == nil {
		return fmt.Errorf("parsed domain is required")
	}
	codegenOption, err := typescript.NormalizeOption(option)
	if err != nil {
		return err
	}
	return typescript.GenerateManaged(domain, codegenOption)
}

// CompileTypeScript parses input and generates TypeScript source. Parsing
// completes before generated output is committed.
func CompileTypeScript(input Input, option TypeScriptOption) (CompileResult, error) {
	compilerOption, err := normalizeInput(input)
	if err != nil {
		return CompileResult{}, err
	}
	codegenOption, err := typescript.NormalizeOption(option)
	if err != nil {
		return CompileResult{}, err
	}
	parsed, err := compileInput(context.Background(), input, compilerOption, false)
	if err != nil {
		return CompileResult{}, err
	}
	if err := typescript.GenerateManaged(parsed.Domain, codegenOption); err != nil {
		return CompileResult{}, err
	}
	return CompileResult{Diagnostics: parsed.Diagnostics}, nil
}

// GenerateSkeleton generates a Skel contract from a parsed domain. Stale files
// carrying the skelc generated marker may be removed.
func GenerateSkeleton(domain *model.Domain, option SkeletonOption) error {
	if domain == nil {
		return fmt.Errorf("parsed domain is required")
	}
	codegenOption, err := skeleton.NormalizeOption(option)
	if err != nil {
		return err
	}
	return skeleton.GenerateManaged(domain, codegenOption)
}

// CompileSkeleton parses input and generates a Skel contract. Parsing completes
// before generated output is committed.
func CompileSkeleton(input Input, option SkeletonOption) (CompileResult, error) {
	compilerOption, err := normalizeInput(input)
	if err != nil {
		return CompileResult{}, err
	}
	codegenOption, err := skeleton.NormalizeOption(option)
	if err != nil {
		return CompileResult{}, err
	}
	parsed, err := compileInput(context.Background(), input, compilerOption, false)
	if err != nil {
		return CompileResult{}, err
	}
	if err := skeleton.GenerateManaged(parsed.Domain, codegenOption); err != nil {
		return CompileResult{}, err
	}
	return CompileResult{Diagnostics: parsed.Diagnostics}, nil
}
