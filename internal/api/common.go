package api

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skel/diagnostic"
	internalcompiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/optionvalidation"
	"go.yorun.ai/skel/internal/source"
	"go.yorun.ai/skel/schema"
)

// Input identifies the primary Skel source and any imported domains.
type Input struct {
	// SkelIn is the path to a Skel source file or domain directory.
	SkelIn string
	// Sources is an optional complete in-memory snapshot, keyed by logical source
	// path. Nil reads the filesystem; non-nil never falls back to disk. Relative
	// paths resolve against the current working directory.
	Sources map[string][]byte
	// SkelImports maps the complete transitive import closure to Skel source
	// files or directories. Only SkelIn is a generation target.
	SkelImports map[string]string
	// Strict rejects legacy declarations accepted with migration warnings.
	Strict bool
}

// ParseResult contains a validated semantic schema and non-fatal diagnostics.
type ParseResult struct {
	// Domain is the validated semantic schema, including compatibility hashes.
	Domain *schema.Domain
	// Diagnostics contains structured non-fatal diagnostics produced while parsing.
	Diagnostics []diagnostic.Diagnostic
}

// Parse loads and validates a Skel contract for use by custom generators and
// tools. All direct and transitive imported domains must be declared in
// Input.SkelImports.
func Parse(input Input) (ParseResult, error) {
	return ParseContext(context.Background(), input)
}

// ParseContext is Parse with cancellation support, including frozen inputs.
func ParseContext(ctx context.Context, input Input) (ParseResult, error) {
	option, err := normalizeInput(input)
	if err != nil {
		return ParseResult{}, err
	}
	parsed, parseErr := compileInput(ctx, input, option, false)
	if parseErr != nil {
		return ParseResult{}, parseErr
	}
	return ParseResult{Domain: parsed.Domain, Diagnostics: parsed.Diagnostics}, nil
}

func normalizeInput(input Input) (internalcompiler.Option, error) {
	if strings.TrimSpace(input.SkelIn) == "" {
		return internalcompiler.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldSkelInput, optionvalidation.RuleRequired, "skel input is required")
	}
	skelIn, err := absolutePath(input.SkelIn)
	if err != nil {
		return internalcompiler.Option{}, err
	}
	imports, err := normalizePathMap(input.SkelImports)
	if err != nil {
		return internalcompiler.Option{}, err
	}
	return internalcompiler.Option{SkelIn: skelIn, SkelImports: imports, Strict: input.Strict}, nil
}

func absolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", path, err)
	}
	return absPath, nil
}

func normalizePathMap(values map[string]string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(values))
	for _, key := range sortedMapKeys(values) {
		value := values[key]
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldSkelImport, optionvalidation.RuleInvalid, "Skel import domain is required")
		}
		if strings.TrimSpace(value) == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldSkelImport, optionvalidation.RuleRequired,
				fmt.Sprintf("Skel import path for domain %s is required", normalizedKey))
		}
		path, err := absolutePath(value)
		if err != nil {
			return nil, err
		}
		if _, exists := normalized[normalizedKey]; exists {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldSkelImport, optionvalidation.RuleInvalid,
				fmt.Sprintf("duplicate Skel import domain %s", normalizedKey))
		}
		normalized[normalizedKey] = path
	}
	return normalized, nil
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func inputProvider(input Input) (loader.Provider, error) {
	if input.Sources == nil {
		return loader.FileSystem{}, nil
	}
	paths := make([]string, 0, len(input.Sources))
	for path := range input.Sources {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	documents := make([]*source.Document, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldSkelInput, optionvalidation.RuleInvalid, "source path is required")
		}
		absolute, err := absolutePath(path)
		if err != nil {
			return nil, err
		}
		if seen[absolute] {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldSkelInput, optionvalidation.RuleInvalid, fmt.Sprintf("duplicate source path %s", absolute))
		}
		seen[absolute] = true
		documents = append(documents, source.New(source.ID(absolute), absolute, 0, string(input.Sources[path])))
	}
	for _, document := range documents {
		path := document.AnalysisPath()
		for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
			if seen[parent] {
				return nil, optionvalidation.NewValidationError(optionvalidation.FieldSkelInput, optionvalidation.RuleInvalid, fmt.Sprintf("source path %s is also a directory containing %s", parent, path))
			}
			if parent == filepath.Dir(parent) {
				break
			}
		}
	}
	return loader.NewMemory(documents...), nil
}

func compileInput(ctx context.Context, input Input, option internalcompiler.Option, unresolved bool) (internalcompiler.Result, error) {
	provider, err := inputProvider(input)
	if err != nil {
		return internalcompiler.Result{}, err
	}
	if unresolved {
		return internalcompiler.CompileImportFrom(ctx, provider, option)
	}
	return internalcompiler.CompileFrom(ctx, provider, option)
}
