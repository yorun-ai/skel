package skelc

import "go.yorun.ai/skelc/internal/codegen/common"

// ApiTypeDependency identifies an external data or enum required by an API view.
type ApiTypeDependency = common.ApiTypeDependency

// ApiDependencyReport lists selected local declarations and external type references.
type ApiDependencyReport = common.ApiDependencyReport

// ApiDependencyResult contains the dependency report and source diagnostics.
type ApiDependencyResult struct {
	Report      *ApiDependencyReport
	Diagnostics []Diagnostic
}

// QueryApiDependencies uses the same selection as API code generation. It follows
// local types but leaves foreign declarations for the caller to query separately.
func QueryApiDependencies(input Input, selection ApiFilter) (ApiDependencyResult, error) {
	normalized, err := common.NormalizeApiFilter(selection)
	if err != nil {
		return ApiDependencyResult{}, err
	}
	parsed, err := Parse(input)
	if err != nil {
		return ApiDependencyResult{}, err
	}
	report, err := common.ApiDependencies(parsed.Domain, normalized)
	if err != nil {
		return ApiDependencyResult{}, err
	}
	return ApiDependencyResult{Report: report, Diagnostics: parsed.Diagnostics}, nil
}
