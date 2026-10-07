package api

import (
	internalapi "go.yorun.ai/skel/internal/api"
	"go.yorun.ai/skel/schema"
)

// MinimumGolangVineVersion is the minimum Vine module version supported by
// generated Go code.
const MinimumGolangVineVersion = internalapi.MinimumGolangVineVersion

// DefaultGolangVineVersion is the Vine module version used for generated Go
// modules when GolangOption.VineVersion is empty.
const DefaultGolangVineVersion = internalapi.DefaultGolangVineVersion

// CompileResult contains structured non-fatal diagnostics produced while loading and parsing Skel sources.
type CompileResult = internalapi.CompileResult

// ApiFilter selects API services and optional local type roots for pruning.
type ApiFilter = internalapi.ApiFilter

// GolangOption configures Go generation.
type GolangOption = internalapi.GolangOption

// TypeScriptOption configures TypeScript generation.
type TypeScriptOption = internalapi.TypeScriptOption

// SkeletonOption configures Skel source generation.
type SkeletonOption = internalapi.SkeletonOption

// GenerateGolang generates Go source or a standalone Go module from a parsed
// domain. Stale files carrying the skelc generated marker may be removed.
func GenerateGolang(domain *schema.Domain, option GolangOption) error {
	return internalapi.GenerateGolang(domain, option)
}

// CompileGolang parses input and generates Go source or a standalone Go module.
// Parsing completes before any generated output is committed.
func CompileGolang(input Input, option GolangOption) (CompileResult, error) {
	return internalapi.CompileGolang(input, option)
}

// GenerateTypeScript generates TypeScript source from a parsed domain. Stale
// files carrying the skelc generated marker may be removed.
func GenerateTypeScript(domain *schema.Domain, option TypeScriptOption) error {
	return internalapi.GenerateTypeScript(domain, option)
}

// CompileTypeScript parses input and generates TypeScript source. Parsing
// completes before generated output is committed.
func CompileTypeScript(input Input, option TypeScriptOption) (CompileResult, error) {
	return internalapi.CompileTypeScript(input, option)
}

// GenerateSkeleton generates a Skel contract from a parsed domain. Stale files
// carrying the skelc generated marker may be removed.
func GenerateSkeleton(domain *schema.Domain, option SkeletonOption) error {
	return internalapi.GenerateSkeleton(domain, option)
}

// CompileSkeleton parses input and generates a Skel contract. Parsing completes
// before generated output is committed.
func CompileSkeleton(input Input, option SkeletonOption) (CompileResult, error) {
	return internalapi.CompileSkeleton(input, option)
}
