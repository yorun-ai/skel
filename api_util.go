package skelc

import (
	"fmt"
	gotoken "go/token"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/codegen/golang"
	"go.yorun.ai/skelc/internal/codegen/skeleton"
	"go.yorun.ai/skelc/internal/codegen/typescript"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/optionvalidation"
	"go.yorun.ai/skelc/internal/util/nameutil"
	"go.yorun.ai/skelc/model"
)

func normalizeInput(input Input) (compiler.Option, error) {
	if strings.TrimSpace(input.SkelIn) == "" {
		return compiler.Option{}, optionvalidation.NewValidationError(optionvalidation.FieldSkelInput, optionvalidation.RuleRequired, "skel input is required")
	}
	skelIn, err := absolutePath(input.SkelIn)
	if err != nil {
		return compiler.Option{}, err
	}
	imports, err := normalizePathMap(input.SkelImports)
	if err != nil {
		return compiler.Option{}, err
	}
	return compiler.Option{SkelIn: skelIn, SkelImports: imports, Strict: input.Strict}, nil
}

func normalizeGolangOption(option GolangOption) (golang.ResolvedOption, error) {
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
		apiDomains, err = apiImportDomains(domain)
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
		AsModule:    option.AsModule,
		Out:         out,
		Module:      module,
		Imports:     imports,
		ModuleScope: moduleScope,
	}, nil
}

func validateTypeScriptImports(domain *model.Domain, option typescript.Option) error {
	apiDomains, err := apiImportDomains(domain)
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

func absolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", path, err)
	}
	return absPath, nil
}

func optionalAbsolutePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	return absolutePath(path)
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

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func apiImportDomains(domain *model.Domain) (map[string]bool, error) {
	if err := common.ValidateDomain(domain); err != nil {
		return nil, err
	}
	view := common.BuildApiView(domain)
	domains := map[string]bool{}
	common.VisitTypes(common.ApiTypeRoots(view.Data, view.Services), func(kind *model.Type) {
		if kind.ExternalDomain != "" {
			domains[kind.ExternalDomain] = true
		}
	})
	return domains, nil
}
