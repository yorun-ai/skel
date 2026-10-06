package golang

import (
	"fmt"
	gotoken "go/token"
	"path/filepath"
	"strings"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/optionvalidation"
	"go.yorun.ai/skel/internal/util/nameutil"
)

// NormalizeOption validates and normalizes generation options before source compilation.
func NormalizeOption(option Option) (ResolvedOption, error) {
	if err := codegen.ValidateApiFilterMode(option.ApiFilter, option.ApiOnly); err != nil {
		return ResolvedOption{}, err
	}
	apiFilter, err := codegen.NormalizeApiFilter(option.ApiFilter)
	if err != nil {
		return ResolvedOption{}, err
	}
	if option.ApiOnly && option.PubOnly {
		return ResolvedOption{}, fmt.Errorf("api and pub are mutually exclusive")
	}
	if (option.ApiOnly || option.PubOnly) && (option.PubOut != "" || option.PubModule != "") {
		return ResolvedOption{}, fmt.Errorf("api/pub mode cannot be combined with go-pub-out or go-pub-module")
	}
	if strings.TrimSpace(option.Out) == "" {
		return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoOutput, optionvalidation.RuleRequired, "Go output is required")
	}
	modulePrefix := strings.TrimSpace(option.ModulePrefix)
	module := strings.TrimSpace(option.Module)
	pubOutValue := strings.TrimSpace(option.PubOut)
	pubModule := strings.TrimSpace(option.PubModule)
	if option.AsModule {
		if module == "" && modulePrefix == "" {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoModuleIdentity, optionvalidation.RuleRequired, "Go module or module prefix is required")
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
				return ResolvedOption{}, optionvalidation.NewValidationError(field.field, optionvalidation.RuleRequiresModule, field.message)
			}
		}
	}
	if pubModule != "" && pubOutValue == "" {
		return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoPublicModule, optionvalidation.RuleRequiresPublicOutput, "Go public module requires public output")
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
			return ResolvedOption{}, err
		}
		if field.value != "" {
			if err := ValidateModulePath(field.value, field.field); err != nil {
				return ResolvedOption{}, err
			}
		}
	}
	out, err := binding.AbsolutePath(option.Out)
	if err != nil {
		return ResolvedOption{}, err
	}
	if !option.AsModule {
		name := filepath.Base(out)
		if !nameutil.IsSnakeCase(name) || gotoken.Lookup(name).IsKeyword() {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoOutput, optionvalidation.RuleInvalid,
				fmt.Sprintf("go output directory name %q is not a valid package name", name))
		}
	}
	pubOut, err := binding.OptionalAbsolutePath(pubOutValue)
	if err != nil {
		return ResolvedOption{}, err
	}
	imports, err := normalizeImportMap(option.Imports)
	if err != nil {
		return ResolvedOption{}, err
	}
	for _, domain := range binding.SortedMapKeys(imports) {
		path := imports[domain]
		if _, err := ImportPath(path); err != nil {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleInvalid, err.Error())
		}
	}

	resolved, err := ResolveOption(Option{
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
		return ResolvedOption{}, err
	}
	return resolved, nil
}

func validateGolangImports(domain *model.Domain, option Option) error {
	var apiDomains map[string]bool
	if option.ApiOnly {
		var err error
		apiDomains, err = codegen.ApiImportDomains(domain, option.ApiFilter)
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
	for _, key := range binding.SortedMapKeys(values) {
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
