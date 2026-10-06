package skel

import (
	"context"
	"errors"
	"fmt"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/optionvalidation"
	internalschema "go.yorun.ai/skel/internal/schema"
	"go.yorun.ai/skel/internal/schema/sourcediff"
	"go.yorun.ai/skel/schema"
)

// ApiTypeDependency identifies an external data or enum required by an API view.
type ApiTypeDependency = common.ApiTypeDependency

// ApiDependencyReport lists selected local declarations and external type references.
type ApiDependencyReport = common.ApiDependencyReport

// ApiDependencyResult contains the dependency report and source diagnostics.
type ApiDependencyResult struct {
	Report      *ApiDependencyReport
	Diagnostics []diagnostic.Diagnostic
}

// QueryApiDependencies uses the same selection as API code generation. It follows
// local types but leaves foreign declarations for the caller to query separately.
func QueryApiDependencies(input Input, selection ApiFilter) (ApiDependencyResult, error) {
	return QueryApiDependenciesContext(context.Background(), input, selection)
}

// QueryApiDependenciesContext is QueryApiDependencies with cancellation support.
func QueryApiDependenciesContext(ctx context.Context, input Input, selection ApiFilter) (ApiDependencyResult, error) {
	normalized, err := common.NormalizeApiFilter(selection)
	if err != nil {
		return ApiDependencyResult{}, err
	}
	parsed, err := ParseContext(ctx, input)
	if err != nil {
		return ApiDependencyResult{}, err
	}
	report, err := common.ApiDependencies(parsed.Domain, normalized)
	if err != nil {
		return ApiDependencyResult{}, err
	}
	return ApiDependencyResult{Report: report, Diagnostics: parsed.Diagnostics}, nil
}

// SchemaDependencyOption selects the complete domain, backend public contract, or API view.
type SchemaDependencyOption struct {
	Pub       bool
	Api       bool
	ApiFilter ApiFilter
}

// SchemaDependencyReport contains selected local declarations and external references.
type SchemaDependencyReport = internalschema.DependencyReport

// SchemaDeclarationDependency identifies one external declaration.
type SchemaDeclarationDependency = internalschema.Dependency

// SchemaDependencyResult includes the report and compiler diagnostics.
type SchemaDependencyResult struct {
	Report      *SchemaDependencyReport
	Diagnostics []diagnostic.Diagnostic
}

// QuerySchemaDependencies shares declaration selection with list and generation.
// It fully resolves imports but reports only the selected domain's references.
func QuerySchemaDependencies(input Input, selection SchemaDependencyOption) (SchemaDependencyResult, error) {
	return QuerySchemaDependenciesContext(context.Background(), input, selection)
}

// QuerySchemaDependenciesContext is QuerySchemaDependencies with cancellation support.
func QuerySchemaDependenciesContext(ctx context.Context, input Input, selection SchemaDependencyOption) (SchemaDependencyResult, error) {
	if selection.Api && selection.Pub {
		return SchemaDependencyResult{}, optionvalidation.NewValidationError(optionvalidation.FieldSchemaView, optionvalidation.RuleInvalid, "flags api and pub are mutually exclusive")
	}
	if err := common.ValidateApiFilterMode(selection.ApiFilter, selection.Api); err != nil {
		return SchemaDependencyResult{}, err
	}
	if selection.Api {
		result, err := QueryApiDependenciesContext(ctx, input, selection.ApiFilter)
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
	result, err := QuerySchemaContext(ctx, input, SchemaQueryOption{Pub: selection.Pub, ResolveImports: true})
	if err != nil {
		return SchemaDependencyResult{}, err
	}
	return SchemaDependencyResult{Report: internalschema.Dependencies(result.Document), Diagnostics: result.Diagnostics}, nil
}

// SchemaQueryOption selects a complete, public, or API schema view.
type SchemaQueryOption struct {
	Pub       bool
	Api       bool
	ApiFilter ApiFilter
	// ResolveImports loads the complete import graph in the default view.
	// Public and API views always resolve imports.
	ResolveImports bool
}

// SchemaQueryResult includes the normalized document and non-fatal diagnostics.
type SchemaQueryResult struct {
	Document    *schema.Document
	Diagnostics []diagnostic.Diagnostic
}

// QuerySchema creates a snapshot for inspection, encoding, listing or lookup.
// The default view keeps unresolved imports, matching schema snapshot/list/get.
func QuerySchema(input Input, selection SchemaQueryOption) (SchemaQueryResult, error) {
	return QuerySchemaContext(context.Background(), input, selection)
}

// QuerySchemaContext is QuerySchema with cancellation support.
func QuerySchemaContext(ctx context.Context, input Input, selection SchemaQueryOption) (SchemaQueryResult, error) {
	if !selection.ResolveImports && !selection.Pub && !selection.Api && len(input.SkelImports) != 0 {
		return SchemaQueryResult{}, optionvalidation.NewValidationError(optionvalidation.FieldSchemaView, optionvalidation.RuleInvalid, "dependency mappings require resolved schema inspection")
	}
	if selection.Api && selection.Pub {
		return SchemaQueryResult{}, optionvalidation.NewValidationError(optionvalidation.FieldSchemaView, optionvalidation.RuleInvalid, "flags api and pub are mutually exclusive")
	}
	if err := common.ValidateApiFilterMode(selection.ApiFilter, selection.Api); err != nil {
		return SchemaQueryResult{}, err
	}
	filter, err := common.NormalizeApiFilter(selection.ApiFilter)
	if err != nil {
		return SchemaQueryResult{}, err
	}
	option, err := normalizeInput(input)
	if err != nil {
		return SchemaQueryResult{}, err
	}
	compiled, err := compileInput(ctx, input, option, !selection.ResolveImports && !selection.Pub && !selection.Api)
	if err != nil {
		return SchemaQueryResult{}, err
	}
	var document *schema.Document
	if selection.Pub || selection.Api {
		var view *common.PublicView
		if selection.Api {
			view, err = common.BuildApiView(compiled.Domain, filter)
		} else {
			view, err = common.BuildPublicView(compiled.Domain)
		}
		if err != nil {
			return SchemaQueryResult{}, err
		}
		document, err = view.ProjectSchema(compiled.Domain)
	} else {
		document, err = schema.Project(compiled.Domain, compiled.ImportAliases)
	}
	if err != nil {
		return SchemaQueryResult{}, err
	}
	return SchemaQueryResult{Document: document, Diagnostics: compiled.Diagnostics}, nil
}

// ErrGitHistoryUnavailable identifies an unavailable implicit Git baseline.
var ErrGitHistoryUnavailable = sourcediff.ErrGitHistoryUnavailable

// ErrSchemaSourceCompilation identifies invalid source in a schema comparison.
var ErrSchemaSourceCompilation = sourcediff.ErrSourceCompilation

// SchemaDiffOption selects an explicit source baseline or Git HEAD.
type SchemaDiffOption struct {
	// Baseline may use disk or frozen sources. Nil selects the candidate path at
	// Git HEAD, which requires a filesystem candidate and a usable repository.
	// Historical baseline declarations are accepted without strict-mode rejection.
	Baseline *Input
}

// DiffSchemaSources compares candidate source with an explicit baseline or Git HEAD.
func DiffSchemaSources(candidate Input, option SchemaDiffOption) (*schema.Report, error) {
	return DiffSchemaSourcesContext(context.Background(), candidate, option)
}

// DiffSchemaSourcesContext is DiffSchemaSources with cancellation support.
func DiffSchemaSourcesContext(ctx context.Context, candidate Input, option SchemaDiffOption) (*schema.Report, error) {
	if option.Baseline == nil {
		normalized, err := normalizeInput(candidate)
		if err != nil {
			return nil, err
		}
		if len(candidate.SkelImports) != 0 {
			return nil, fmt.Errorf("schema source diff does not accept dependency mappings")
		}
		if candidate.Sources != nil {
			return nil, fmt.Errorf("a frozen candidate requires an explicit schema baseline")
		}
		return sourcediff.DiffSource(ctx, normalized.SkelIn, sourcediff.Option{Strict: candidate.Strict})
	}
	selected, err := QuerySchemaContext(ctx, candidate, SchemaQueryOption{})
	if err != nil {
		return nil, schemaSourceError(err)
	}
	baseline := *option.Baseline
	baseline.Strict = false
	previous, err := QuerySchemaContext(ctx, baseline, SchemaQueryOption{})
	if err != nil {
		return nil, schemaSourceError(err)
	}
	return schema.Diff(previous.Document, selected.Document)
}

func schemaSourceError(err error) error {
	var validation *optionvalidation.ValidationError
	if errors.As(err, &validation) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrSchemaSourceCompilation, err)
}
