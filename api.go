// Package skelc provides the supported programmatic API for parsing Skel
// contracts and generating Go, TypeScript, and public Skel output.
//
// Use [Parse] when several generators or a custom generator need to share one
// validated [model.Domain]. The Compile functions combine parsing and one
// generation step for callers that only need a single target.
//
// Generated files carry an ownership marker. Existing files without that
// marker are preserved.
package skelc

import (
	"fmt"

	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/codegen/golang"
	"go.yorun.ai/skelc/internal/codegen/output"
	"go.yorun.ai/skelc/internal/codegen/skeleton"
	"go.yorun.ai/skelc/internal/codegen/typescript"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/model"
)

// MinimumGolangVineVersion is the minimum Vine module version supported by
// generated Go code.
const MinimumGolangVineVersion = golang.MinimumVineVersion

// DefaultGolangVineVersion is the Vine module version used for generated Go
// modules when GolangOption.VineVersion is empty.
const DefaultGolangVineVersion = golang.DefaultVineVersion

// Input identifies the primary Skel source and any imported domains.
type Input struct {
	// SkelIn is the path to a Skel source file or domain directory.
	SkelIn string
	// SkelImports maps the complete transitive import closure to Skel source
	// files or directories. Only SkelIn is a generation target.
	SkelImports map[string]string
	// Strict rejects legacy declarations accepted with migration warnings.
	Strict bool
}

// ParseResult contains a validated semantic model and non-fatal diagnostics.
type ParseResult struct {
	// Domain is the validated semantic model, including compatibility hashes.
	Domain *model.Domain
	// Diagnostics contains structured non-fatal diagnostics produced while parsing.
	Diagnostics []Diagnostic
}

type Diagnostic = diagnostic.Diagnostic
type Diagnostics = diagnostic.Diagnostics
type DiagnosticSeverity = diagnostic.Severity
type SourceRange = diagnostic.SourceRange
type DiagnosticRelatedInformation = diagnostic.RelatedInformation
type DiagnosticSuggestion = diagnostic.Suggestion

const (
	DiagnosticSeverityError   = diagnostic.SeverityError
	DiagnosticSeverityWarning = diagnostic.SeverityWarning

	DiagnosticCodeSyntaxUnexpected   = diagnostic.CodeSyntaxUnexpected
	DiagnosticCodeSyntaxEOF          = diagnostic.CodeSyntaxEOF
	DiagnosticCodeSyntaxFinalize     = diagnostic.CodeSyntaxFinalize
	DiagnosticCodeSemanticValidation = diagnostic.CodeSemanticValidation
	DiagnosticCodeSemanticDuplicate  = diagnostic.CodeSemanticDuplicate
	DiagnosticCodeSemanticNaming     = diagnostic.CodeSemanticNaming
	DiagnosticCodeSemanticReference  = diagnostic.CodeSemanticReference
	DiagnosticCodeSemanticWarning    = diagnostic.CodeSemanticWarning
	DiagnosticCodeServiceModifier    = diagnostic.CodeServiceModifier
	DiagnosticCodeServiceClientRules = diagnostic.CodeServiceClientRules
	DiagnosticCodeImportMissing      = diagnostic.CodeImportMissing
	DiagnosticCodeImportCycle        = diagnostic.CodeImportCycle
	DiagnosticCodeDomainMissing      = diagnostic.CodeDomainMissing
	DiagnosticCodeDomainMismatch     = diagnostic.CodeDomainMismatch
	DiagnosticCodeDomainFileContent  = diagnostic.CodeDomainFileContent
	DiagnosticCodeDomainDecorator    = diagnostic.CodeDomainDecorator
	DiagnosticCodeLoaderDirectory    = diagnostic.CodeLoaderDirectory
	DiagnosticCodeLoaderHiddenFile   = diagnostic.CodeLoaderHiddenFile
	DiagnosticCodeLoaderUnsupported  = diagnostic.CodeLoaderUnsupported
)

// CompileResult contains structured non-fatal diagnostics produced while loading and parsing Skel sources.
type CompileResult struct {
	// Diagnostics contains non-fatal diagnostics produced while parsing.
	Diagnostics []Diagnostic
}

// Parse loads and validates a Skel contract for use by custom generators and
// tools. All direct and transitive imported domains must be declared in
// Input.SkelImports.
func Parse(input Input) (ParseResult, error) {
	option, err := normalizeInput(input)
	if err != nil {
		return ParseResult{}, err
	}
	parsed, parseErr := compiler.Compile(option)
	if parseErr != nil {
		return ParseResult{}, parseErr
	}
	return ParseResult{Domain: parsed.Domain, Diagnostics: parsed.Diagnostics}, nil
}

// ApiFilter selects API services by fully qualified actor names.
type ApiFilter = common.ApiFilter

// GolangOption configures Go generation.
type GolangOption struct {
	// CompilerVersion identifies the actual skelc version embedded in generated metadata.
	// Required for backend output and must be at least v0.17.1.
	// Use v0.0.0-dev only for development builds.
	CompilerVersion string
	// AsModule generates a standalone Go module instead of package source for an
	// existing module.
	AsModule bool
	// PubOnly generates backend public contracts. It is mutually exclusive with ApiOnly.
	PubOnly bool
	// ApiOnly generates standalone Portal clients.
	ApiOnly bool
	// ApiFilter limits API clients to services declaring one of its actors. Requires ApiOnly.
	ApiFilter ApiFilter
	// Out is the output directory for generated Go files.
	Out string
	// Module is the module path used when AsModule is true.
	Module string
	// PubOut is the optional output directory for a separate public Go module.
	PubOut string
	// PubModule is the module path for PubOut.
	PubModule string
	// Imports maps imported Skel domain names to Go import paths.
	Imports map[string]string
	// ModulePrefix derives module paths from domain names when Module or an
	// imported-domain mapping is omitted.
	ModulePrefix string
	// VineVersion selects the go.yorun.ai/vine version written to generated module
	// metadata. It must not be lower than [MinimumGolangVineVersion]. An empty
	// value uses [DefaultGolangVineVersion].
	VineVersion string
	// VrpcVersion selects the standalone client module version.
	VrpcVersion string
}

// TypeScriptOption configures TypeScript generation.
type TypeScriptOption struct {
	// ApiOnly is required for TypeScript client generation.
	ApiOnly bool
	// ApiFilter limits API clients to services declaring one of its actors. Requires ApiOnly.
	ApiFilter ApiFilter
	// AsModule emits package metadata for a standalone npm package.
	AsModule bool
	// Out is the output directory for generated TypeScript files.
	Out string
	// Module is the npm package name used when AsModule is true.
	Module string
	// Imports maps imported Skel domain names to npm package specifiers.
	Imports map[string]string
	// ModuleScope derives npm package names for the current and imported domains.
	ModuleScope string
}

// SkeletonOption configures Skel source generation.
type SkeletonOption struct {
	// PubOnly limits output to declarations in the public contract.
	PubOnly bool
	// Out is the output directory for generated Skel files.
	Out string
}

// GenerateGolang generates Go source or a standalone Go module from a parsed
// domain. Stale files carrying the skelc generated marker may be removed.
func GenerateGolang(domain *model.Domain, option GolangOption) error {
	if domain == nil {
		return fmt.Errorf("parsed domain is required")
	}
	codegenOption, err := normalizeGolangOption(option)
	if err != nil {
		return err
	}
	return generateGolang(domain, codegenOption)
}

func generateGolang(domain *model.Domain, resolved golang.ResolvedOption) error {
	option := resolved.Options()
	if err := validateGolangImports(domain, option); err != nil {
		return err
	}
	return output.RunManagedOutputs([]string{option.Out, option.PubOut}, func(staged []string) error {
		return golang.Generate(domain, resolved.WithOutputs(staged[0], staged[1]))
	})
}

// CompileGolang parses input and generates Go source or a standalone Go module.
// Parsing completes before any generated output is committed.
func CompileGolang(input Input, option GolangOption) (CompileResult, error) {
	compilerOption, err := normalizeInput(input)
	if err != nil {
		return CompileResult{}, err
	}
	codegenOption, err := normalizeGolangOption(option)
	if err != nil {
		return CompileResult{}, err
	}
	parsed, err := compiler.Compile(compilerOption)
	if err != nil {
		return CompileResult{}, err
	}
	if err := generateGolang(parsed.Domain, codegenOption); err != nil {
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
	codegenOption, err := normalizeTypeScriptOption(option)
	if err != nil {
		return err
	}
	return generateTypeScript(domain, codegenOption)
}

func generateTypeScript(domain *model.Domain, option typescript.Option) error {
	if err := validateTypeScriptImports(domain, option); err != nil {
		return err
	}
	return output.RunManagedOutputs([]string{option.Out}, func(staged []string) error {
		option.Out = staged[0]
		return typescript.Generate(domain, option)
	})
}

// CompileTypeScript parses input and generates TypeScript source. Parsing
// completes before generated output is committed.
func CompileTypeScript(input Input, option TypeScriptOption) (CompileResult, error) {
	compilerOption, err := normalizeInput(input)
	if err != nil {
		return CompileResult{}, err
	}
	codegenOption, err := normalizeTypeScriptOption(option)
	if err != nil {
		return CompileResult{}, err
	}
	parsed, err := compiler.Compile(compilerOption)
	if err != nil {
		return CompileResult{}, err
	}
	if err := generateTypeScript(parsed.Domain, codegenOption); err != nil {
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
	codegenOption, err := normalizeSkeletonOption(option)
	if err != nil {
		return err
	}
	return generateSkeleton(domain, codegenOption)
}

func generateSkeleton(domain *model.Domain, option skeleton.Option) error {
	return output.RunManagedOutputs([]string{option.Out}, func(staged []string) error {
		option.Out = staged[0]
		return skeleton.Generate(domain, option)
	})
}

// CompileSkeleton parses input and generates a Skel contract. Parsing completes
// before generated output is committed.
func CompileSkeleton(input Input, option SkeletonOption) (CompileResult, error) {
	compilerOption, err := normalizeInput(input)
	if err != nil {
		return CompileResult{}, err
	}
	codegenOption, err := normalizeSkeletonOption(option)
	if err != nil {
		return CompileResult{}, err
	}
	parsed, err := compiler.Compile(compilerOption)
	if err != nil {
		return CompileResult{}, err
	}
	if err := generateSkeleton(parsed.Domain, codegenOption); err != nil {
		return CompileResult{}, err
	}
	return CompileResult{Diagnostics: parsed.Diagnostics}, nil
}
