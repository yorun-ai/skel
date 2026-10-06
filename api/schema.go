package api

import (
	"context"

	internalapi "go.yorun.ai/skel/internal/api"
	"go.yorun.ai/skel/model"
	"go.yorun.ai/skel/schema"
)

// ApiTypeDependency identifies an external data or enum required by an API view.
type ApiTypeDependency = internalapi.ApiTypeDependency

// ApiDependencyReport lists selected local declarations and external type references.
type ApiDependencyReport = internalapi.ApiDependencyReport

// ApiDependencyResult contains the dependency report and source diagnostics.
type ApiDependencyResult = internalapi.ApiDependencyResult

// QueryApiDependencies uses the same selection as API code generation. It follows
// local types but leaves foreign declarations for the caller to query separately.
func QueryApiDependencies(input Input, selection ApiFilter) (ApiDependencyResult, error) {
	return internalapi.QueryApiDependencies(input, selection)
}

// QueryApiDependenciesContext is QueryApiDependencies with cancellation support.
func QueryApiDependenciesContext(ctx context.Context, input Input, selection ApiFilter) (ApiDependencyResult, error) {
	return internalapi.QueryApiDependenciesContext(ctx, input, selection)
}

// SchemaDependencyOption selects the complete domain, backend public contract, or API view.
type SchemaDependencyOption = internalapi.SchemaDependencyOption

// SchemaDependencyReport contains selected local declarations and external references.
type SchemaDependencyReport = internalapi.SchemaDependencyReport

// SchemaDeclarationDependency identifies one external declaration.
type SchemaDeclarationDependency = internalapi.SchemaDeclarationDependency

// SchemaDependencyResult includes the report and compiler diagnostics.
type SchemaDependencyResult = internalapi.SchemaDependencyResult

// QuerySchemaDependencies shares declaration selection with list and generation.
// It fully resolves imports but reports only the selected domain's references.
func QuerySchemaDependencies(input Input, selection SchemaDependencyOption) (SchemaDependencyResult, error) {
	return internalapi.QuerySchemaDependencies(input, selection)
}

// QuerySchemaDependenciesContext is QuerySchemaDependencies with cancellation support.
func QuerySchemaDependenciesContext(ctx context.Context, input Input, selection SchemaDependencyOption) (SchemaDependencyResult, error) {
	return internalapi.QuerySchemaDependenciesContext(ctx, input, selection)
}

// SchemaQueryOption selects a complete, public, or API schema view.
type SchemaQueryOption = internalapi.SchemaQueryOption

// SchemaQueryResult includes the normalized document and non-fatal diagnostics.
type SchemaQueryResult = internalapi.SchemaQueryResult

// QuerySchema creates a snapshot for inspection, encoding, listing or lookup.
// The default view keeps unresolved imports, matching schema snapshot/list/get.
func QuerySchema(input Input, selection SchemaQueryOption) (SchemaQueryResult, error) {
	return internalapi.QuerySchema(input, selection)
}

// QuerySchemaContext is QuerySchema with cancellation support.
func QuerySchemaContext(ctx context.Context, input Input, selection SchemaQueryOption) (SchemaQueryResult, error) {
	return internalapi.QuerySchemaContext(ctx, input, selection)
}

// ErrGitHistoryUnavailable identifies an unavailable implicit Git baseline.
var ErrGitHistoryUnavailable = internalapi.ErrGitHistoryUnavailable

// ErrSchemaSourceCompilation identifies invalid source in a schema comparison.
var ErrSchemaSourceCompilation = internalapi.ErrSchemaSourceCompilation

// SchemaDiffOption selects an explicit source baseline or Git HEAD.
type SchemaDiffOption = internalapi.SchemaDiffOption

// DiffSchemaSources compares candidate source with an explicit baseline or Git HEAD.
func DiffSchemaSources(candidate Input, option SchemaDiffOption) (*schema.Report, error) {
	return internalapi.DiffSchemaSources(candidate, option)
}

// DiffSchemaSourcesContext is DiffSchemaSources with cancellation support.
func DiffSchemaSourcesContext(ctx context.Context, candidate Input, option SchemaDiffOption) (*schema.Report, error) {
	return internalapi.DiffSchemaSourcesContext(ctx, candidate, option)
}

// ProjectSchema converts a semantic domain into a canonical schema snapshot.
// Optional aliases normalize unresolved imported references.
func ProjectSchema(domain *model.Domain, aliases map[string]string) (*schema.Document, error) {
	return internalapi.ProjectSchema(domain, aliases)
}
