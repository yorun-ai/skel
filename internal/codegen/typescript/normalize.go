package typescript

import (
	"fmt"
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/optionvalidation"
	"strings"
)

// RequestOption describes TypeScript generation before normalization.
type RequestOption struct {
	// ApiOnly is required for TypeScript client generation.
	ApiOnly bool
	// ApiFilter selects API services/types and optionally prunes unused public types. Requires ApiOnly.
	ApiFilter common.ApiFilter
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

// NormalizeOption validates and normalizes generation options before source compilation.
func NormalizeOption(option RequestOption) (Option, error) {
	if err := common.ValidateApiFilterMode(option.ApiFilter, option.ApiOnly); err != nil {
		return Option{}, err
	}
	apiFilter, err := common.NormalizeApiFilter(option.ApiFilter)
	if err != nil {
		return Option{}, err
	}
	if !option.ApiOnly {
		return Option{}, fmt.Errorf("TypeScript generation requires api")
	}
	if strings.TrimSpace(option.Out) == "" {
		return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptOutput, optionvalidation.RuleRequired, "TypeScript output is required")
	}
	moduleScope := strings.TrimRight(strings.TrimSpace(option.ModuleScope), "/")
	module := strings.TrimRight(strings.TrimSpace(option.Module), "/")
	if option.AsModule {
		if module == "" && moduleScope == "" {
			return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModuleIdentity, optionvalidation.RuleRequired, "TypeScript module or module scope is required")
		}
	} else {
		if module != "" {
			return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModule, optionvalidation.RuleRequiresModule, "TypeScript module requires module generation")
		}
		if moduleScope != "" {
			return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModuleScope, optionvalidation.RuleRequiresModule, "TypeScript module scope requires module generation")
		}
	}
	out, err := common.AbsolutePath(option.Out)
	if err != nil {
		return Option{}, err
	}
	imports, err := normalizeTypeScriptImportMap(option.Imports)
	if err != nil {
		return Option{}, err
	}
	for _, domain := range common.SortedMapKeys(imports) {
		path := imports[domain]
		if err := validateVersionedImport(path, "TypeScript"); err != nil {
			return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleInvalid, err.Error())
		}
	}

	return Option{
		ApiFilter:   apiFilter,
		AsModule:    option.AsModule,
		Out:         out,
		Module:      module,
		Imports:     imports,
		ModuleScope: moduleScope,
	}, nil
}

func validateTypeScriptImports(domain *model.Domain, option Option) error {
	apiDomains, err := common.ApiImportDomains(domain, option.ApiFilter)
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

func normalizeTypeScriptImportMap(values map[string]string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(values))
	for _, key := range common.SortedMapKeys(values) {
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
