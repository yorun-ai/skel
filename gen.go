package skelc

import (
	"context"
	"fmt"
	gotoken "go/token"
	"path/filepath"
	"strings"

	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/codegen/golang"
	"go.yorun.ai/skelc/internal/codegen/output"
	"go.yorun.ai/skelc/internal/codegen/skeleton"
	"go.yorun.ai/skelc/internal/codegen/typescript"
	"go.yorun.ai/skelc/internal/optionvalidation"
	"go.yorun.ai/skelc/internal/util/nameutil"
	"go.yorun.ai/skelc/model"
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
	// ApiOnly generates standalone portal clients.
	ApiOnly bool
	// ApiFilter selects API services/types and optionally prunes unused public types. Requires ApiOnly.
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
	// ApiFilter selects API services/types and optionally prunes unused public types. Requires ApiOnly.
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
	parsed, err := compileInput(context.Background(), input, compilerOption, false)
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
	parsed, err := compileInput(context.Background(), input, compilerOption, false)
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
	parsed, err := compileInput(context.Background(), input, compilerOption, false)
	if err != nil {
		return CompileResult{}, err
	}
	if err := generateSkeleton(parsed.Domain, codegenOption); err != nil {
		return CompileResult{}, err
	}
	return CompileResult{Diagnostics: parsed.Diagnostics}, nil
}

func normalizeGolangOption(option GolangOption) (golang.ResolvedOption, error) {
	if err := common.ValidateApiFilterMode(option.ApiFilter, option.ApiOnly); err != nil {
		return golang.ResolvedOption{}, err
	}
	apiFilter, err := common.NormalizeApiFilter(option.ApiFilter)
	if err != nil {
		return golang.ResolvedOption{}, err
	}
	if option.ApiOnly && option.PubOnly {
		return golang.ResolvedOption{}, fmt.Errorf("api and pub are mutually exclusive")
	}
	if (option.ApiOnly || option.PubOnly) && (option.PubOut != "" || option.PubModule != "") {
		return golang.ResolvedOption{}, fmt.Errorf("api/pub mode cannot be combined with go-pub-out or go-pub-module")
	}
	if strings.TrimSpace(option.Out) == "" {
		return golang.ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoOutput, optionvalidation.RuleRequired, "Go output is required")
	}
	modulePrefix := strings.TrimSpace(option.ModulePrefix)
	module := strings.TrimSpace(option.Module)
	pubOutValue := strings.TrimSpace(option.PubOut)
	pubModule := strings.TrimSpace(option.PubModule)
	if option.AsModule {
		if module == "" && modulePrefix == "" {
			return golang.ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoModuleIdentity, optionvalidation.RuleRequired, "Go module or module prefix is required")
		}
	} else {
		invalidFields := []struct {
			field   optionvalidation.Field
			value   string
			message string
		}{
			{optionvalidation.FieldGoPublicOutput, pubOutValue, "Go public output requires module generation"},
			{optionvalidation.FieldGoPublicModule, pubModule, "Go public module requires module generation"},
			{optionvalidation.FieldGoModule, module, "Go module requires module generation"},
			{optionvalidation.FieldGoModulePrefix, modulePrefix, "Go module prefix requires module generation"},
		}
		for _, field := range invalidFields {
			if field.value != "" {
				return golang.ResolvedOption{}, optionvalidation.NewValidationError(field.field, optionvalidation.RuleRequiresModule, field.message)
			}
		}
	}
	if pubModule != "" && pubOutValue == "" {
		return golang.ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoPublicModule, optionvalidation.RuleRequiresPublicOutput, "Go public module requires public output")
	}

	moduleFields := []struct {
		value string
		name  string
		field optionvalidation.Field
	}{
		{modulePrefix, "Go module prefix", optionvalidation.FieldGoModulePrefix},
		{module, "Go module", optionvalidation.FieldGoModule},
		{pubModule, "Go public module", optionvalidation.FieldGoPublicModule},
	}
	for _, field := range moduleFields {
		if err := checkNoTrailingSlash(field.value, field.name, field.field); err != nil {
			return golang.ResolvedOption{}, err
		}
		if field.value != "" {
			if err := golang.ValidateModulePath(field.value, field.field); err != nil {
				return golang.ResolvedOption{}, err
			}
		}
	}
	out, err := absolutePath(option.Out)
	if err != nil {
		return golang.ResolvedOption{}, err
	}
	if !option.AsModule {
		name := filepath.Base(out)
		if !nameutil.IsSnakeCase(name) || gotoken.Lookup(name).IsKeyword() {
			return golang.ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoOutput, optionvalidation.RuleInvalid,
				fmt.Sprintf("go output directory name %q is not a valid package name", name))
		}
	}
	pubOut, err := optionalAbsolutePath(pubOutValue)
	if err != nil {
		return golang.ResolvedOption{}, err
	}
	imports, err := normalizeImportMap(option.Imports)
	if err != nil {
		return golang.ResolvedOption{}, err
	}
	for _, domain := range sortedMapKeys(imports) {
		path := imports[domain]
		if _, err := golang.ImportPath(path); err != nil {
			return golang.ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleInvalid, err.Error())
		}
	}

	resolved, err := golang.ResolveOption(golang.Option{
		ApiFilter:       apiFilter,
		PubOnly:         option.PubOnly,
		ApiOnly:         option.ApiOnly,
		VrpcVersion:     option.VrpcVersion,
		CompilerVersion: strings.TrimSpace(option.CompilerVersion),
		AsModule:        option.AsModule,
		Out:             out,
		Module:          module,
		PubOut:          pubOut,
		PubModule:       pubModule,
		Imports:         imports,
		ModulePrefix:    modulePrefix,
		VineVersion:     strings.TrimSpace(option.VineVersion),
	})
	if err != nil {
		return golang.ResolvedOption{}, err
	}
	return resolved, nil
}

func validateGolangImports(domain *model.Domain, option golang.Option) error {
	var apiDomains map[string]bool
	if option.ApiOnly {
		var err error
		apiDomains, err = apiImportDomains(domain, option.ApiFilter)
		if err != nil {
			return err
		}
	}
	for _, domainImport := range domain.Imports() {
		if domainImport == nil {
			return fmt.Errorf("generated model contains nil import")
		}
		if option.ApiOnly && !apiDomains[domainImport.Name] {
			continue
		}
		if option.Imports[domainImport.Name] == "" && option.ModulePrefix == "" {
			return optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleRequired,
				fmt.Sprintf("missing Go import for domain %s; set Imports[%q] or ModulePrefix", domainImport.Name, domainImport.Name))
		}
	}
	return nil
}

func normalizeTypeScriptOption(option TypeScriptOption) (typescript.Option, error) {
	if err := common.ValidateApiFilterMode(option.ApiFilter, option.ApiOnly); err != nil {
		return typescript.Option{}, err
	}
	apiFilter, err := common.NormalizeApiFilter(option.ApiFilter)
	if err != nil {
		return typescript.Option{}, err
	}
	if !option.ApiOnly {
		return typescript.Option{}, fmt.Errorf("TypeScript generation requires api")
	}
	if strings.TrimSpace(option.Out) == "" {
		return typescript.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptOutput, optionvalidation.RuleRequired, "TypeScript output is required")
	}
	moduleScope := strings.TrimRight(strings.TrimSpace(option.ModuleScope), "/")
	module := strings.TrimRight(strings.TrimSpace(option.Module), "/")
	if option.AsModule {
		if module == "" && moduleScope == "" {
			return typescript.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModuleIdentity, optionvalidation.RuleRequired, "TypeScript module or module scope is required")
		}
	} else {
		if module != "" {
			return typescript.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModule, optionvalidation.RuleRequiresModule, "TypeScript module requires module generation")
		}
		if moduleScope != "" {
			return typescript.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModuleScope, optionvalidation.RuleRequiresModule, "TypeScript module scope requires module generation")
		}
	}
	out, err := absolutePath(option.Out)
	if err != nil {
		return typescript.Option{}, err
	}
	imports, err := normalizeTypeScriptImportMap(option.Imports)
	if err != nil {
		return typescript.Option{}, err
	}
	for _, domain := range sortedMapKeys(imports) {
		path := imports[domain]
		if err := validateVersionedImport(path, "TypeScript"); err != nil {
			return typescript.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleInvalid, err.Error())
		}
	}

	return typescript.Option{
		ApiFilter:   apiFilter,
		AsModule:    option.AsModule,
		Out:         out,
		Module:      module,
		Imports:     imports,
		ModuleScope: moduleScope,
	}, nil
}

func validateTypeScriptImports(domain *model.Domain, option typescript.Option) error {
	apiDomains, err := apiImportDomains(domain, option.ApiFilter)
	if err != nil {
		return err
	}
	for _, domainImport := range domain.Imports() {
		if domainImport == nil {
			return fmt.Errorf("generated model contains nil import")
		}
		if !apiDomains[domainImport.Name] {
			continue
		}
		if option.Imports[domainImport.Name] == "" && option.ModuleScope == "" {
			return optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleRequired,
				fmt.Sprintf("missing TypeScript import for domain %s; set Imports[%q] or ModuleScope", domainImport.Name, domainImport.Name))
		}
	}
	return nil
}

func normalizeSkeletonOption(option SkeletonOption) (skeleton.Option, error) {
	if strings.TrimSpace(option.Out) == "" {
		return skeleton.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldSkeletonOutput, optionvalidation.RuleRequired, "Skel output is required")
	}
	if !option.PubOnly {
		return skeleton.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldSkeletonPublicOnly, optionvalidation.RuleRequired, "Skel generation requires PubOnly")
	}
	out, err := absolutePath(option.Out)
	if err != nil {
		return skeleton.Option{}, err
	}
	return skeleton.Option{PubOnly: option.PubOnly, Out: out}, nil
}

func validateVersionedImport(path, kind string) error {
	index := strings.LastIndex(path, "@")
	if index < 0 {
		return nil
	}
	if index == 0 && kind == "go" {
		return fmt.Errorf("invalid %s import %q: missing module", kind, path)
	}
	if index == 0 {
		return nil
	}
	if index == len(path)-1 {
		return fmt.Errorf("invalid %s import %q: missing version", kind, path)
	}
	return nil
}

func checkNoTrailingSlash(value, label string, field optionvalidation.Field) error {
	if strings.HasSuffix(value, "/") {
		return optionvalidation.NewValidationError(field, optionvalidation.RuleNoTrailingSlash, label+" must not end with /")
	}
	return nil
}

func normalizeImportMap(values map[string]string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(values))
	for _, key := range sortedMapKeys(values) {
		value := values[key]
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleInvalid, "Go import domain is required")
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleRequired,
				fmt.Sprintf("Go import path for domain %s is required", normalizedKey))
		}
		if err := checkNoTrailingSlash(value, "Go import", optionvalidation.FieldGoImport); err != nil {
			return nil, err
		}
		if _, exists := normalized[normalizedKey]; exists {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleInvalid,
				fmt.Sprintf("duplicate Go import domain %s", normalizedKey))
		}
		normalized[normalizedKey] = value
	}
	return normalized, nil
}

func normalizeTypeScriptImportMap(values map[string]string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(values))
	for _, key := range sortedMapKeys(values) {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleInvalid, "TypeScript import domain is required")
		}
		value := strings.TrimRight(strings.TrimSpace(values[key]), "/")
		if value == "" {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleRequired,
				fmt.Sprintf("TypeScript import path for domain %s is required", normalizedKey))
		}
		if _, exists := normalized[normalizedKey]; exists {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleInvalid,
				fmt.Sprintf("duplicate TypeScript import domain %s", normalizedKey))
		}
		normalized[normalizedKey] = value
	}
	return normalized, nil
}

func apiImportDomains(domain *model.Domain, selection common.ApiFilter) (map[string]bool, error) {
	if err := common.ValidateDomain(domain); err != nil {
		return nil, err
	}
	view, err := common.BuildApiView(domain, selection)
	if err != nil {
		return nil, err
	}
	domains := map[string]bool{}
	common.VisitTypes(common.ApiTypeRoots(view.Data, view.Services), func(kind *model.Type) {
		if kind.ExternalDomain != "" {
			domains[kind.ExternalDomain] = true
		}
	})
	return domains, nil
}
