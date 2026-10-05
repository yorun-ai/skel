package skelc

import (
	"fmt"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/optionvalidation"
	"go.yorun.ai/skelc/internal/schema"
)

// SchemaDependencyOption selects the complete domain, backend public contract, or API view.
type SchemaDependencyOption struct {
	Pub       bool
	Api       bool
	ApiFilter ApiFilter
}

// SchemaDependencyReport contains selected local declarations and external references.
type SchemaDependencyReport = schema.DependencyReport

// SchemaDeclarationDependency identifies one external declaration.
type SchemaDeclarationDependency = schema.Dependency

// SchemaDependencyResult includes the report and compiler diagnostics.
type SchemaDependencyResult struct {
	Report      *SchemaDependencyReport
	Diagnostics []Diagnostic
}

// QuerySchemaDependencies shares declaration selection with list and generation.
// It fully resolves imports but reports only the selected domain's references.
func QuerySchemaDependencies(input Input, selection SchemaDependencyOption) (SchemaDependencyResult, error) {
	if selection.Api && selection.Pub {
		return SchemaDependencyResult{}, optionvalidation.NewValidationError(optionvalidation.FieldSchemaView, optionvalidation.RuleInvalid, "flags api and pub are mutually exclusive")
	}
	if err := common.ValidateApiFilterMode(selection.ApiFilter, selection.Api); err != nil {
		return SchemaDependencyResult{}, err
	}
	if selection.Api {
		result, err := QueryApiDependencies(input, selection.ApiFilter)
		if err != nil {
			return SchemaDependencyResult{}, err
		}
		report := new(SchemaDependencyReport{
			Domain:   result.Report.Domain,
			Services: result.Report.Services, Data: result.Report.Data, Enums: result.Report.Enums,
			Dependencies: result.Report.Dependencies,
		})
		return SchemaDependencyResult{Report: report, Diagnostics: result.Diagnostics}, nil
	}
	parsed, err := Parse(input)
	if err != nil {
		return SchemaDependencyResult{}, err
	}
	var document *schema.Document
	if selection.Pub {
		view, viewErr := common.BuildPublicView(parsed.Domain)
		if viewErr != nil {
			return SchemaDependencyResult{}, viewErr
		}
		document, err = view.ProjectSchema(parsed.Domain)
	} else {
		document, err = schema.Project(parsed.Domain, nil)
	}
	if err != nil {
		return SchemaDependencyResult{}, fmt.Errorf("project dependency schema: %w", err)
	}
	return SchemaDependencyResult{Report: schema.Dependencies(document), Diagnostics: parsed.Diagnostics}, nil
}
