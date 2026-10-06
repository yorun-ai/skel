package analyzer

import (
	"context"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

type Analysis struct {
	name        string
	description string
	schema      *schema.Domain

	imports []*schema.Import

	enums     []*schema.Enum
	dataList  []*schema.Data
	configs   []*schema.Data
	events    []*schema.Data
	actors    []*schema.Actor
	resources []*schema.Resource
	webs      []*schema.Web
	services  []*schema.Service
	tasks     []*schema.Task

	warnings []string

	content *grammar.SkelContent

	enumsMap     map[string]*schema.Enum
	dataMap      map[string]*schema.Data
	actorsMap    map[string]*schema.Actor
	resourcesMap map[string]*schema.Resource
	websMap      map[string]*schema.Web
	servicesMap  map[string]*schema.Service
	tasksMap     map[string]*schema.Task
	importsMap   map[string]*_DomainImport

	reporter    *_DiagnosticReporter
	invalidData map[*schema.Data]bool
	unavailable map[string]bool
}

type _DomainImport struct {
	Domain *Analysis
	Schema *schema.Import
}

// Analyze analyzes a domain and explicitly reports independent
// validation failures. Invalid declarations are excluded from later global
// validation stages to avoid dependent cascade diagnostics.
func Analyze(content *grammar.SkelContent, importedDomains []*Analysis) (*Analysis, []error) {
	domain, diagnostics, _ := AnalyzeContext(context.Background(), content, importedDomains)
	return domain, diagnostics
}

// AnalyzeContext returns no partial analysis when cancelled.
func AnalyzeContext(ctx context.Context, content *grammar.SkelContent, importedDomains []*Analysis) (*Analysis, []error, error) {
	return analyzeContext(ctx, content, importedDomains, false)
}

func analyzeContext(ctx context.Context, content *grammar.SkelContent, importedDomains []*Analysis, allowMissingImports bool) (*Analysis, []error, error) {
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	domain := newAnalysis(content)
	domain.reporter.ctx = ctx
	domainByName := map[string]*Analysis{}
	for _, importedDomain := range importedDomains {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		domainByName[importedDomain.Schema().Name()] = importedDomain
	}
	if !domain.load() {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		domain.reporter.ctx = nil
		return domain, domain.reporter.result(), nil
	}
	diagnosticsBeforeImports := len(domain.reporter.errors)
	if allowMissingImports {
		domain.loadUnresolvedImports()
	} else {
		domain.loadImports(domainByName)
	}
	if ctx.Err() == nil && len(domain.reporter.errors) == diagnosticsBeforeImports {
		domain.normalizeWithMissingImports(allowMissingImports)
	}
	if ctx.Err() == nil && len(domain.reporter.errors) == 0 {
		domain.finalize()
	}
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	domain.reporter.ctx = nil
	return domain, domain.reporter.result(), nil
}

func AnalyzeImport(content *grammar.SkelContent) (*Analysis, []error) {
	domain, diagnostics, _ := AnalyzeImportContext(context.Background(), content)
	return domain, diagnostics
}

// AnalyzeImportContext analyzes unresolved imports without publishing cancelled work.
func AnalyzeImportContext(ctx context.Context, content *grammar.SkelContent) (*Analysis, []error, error) {
	return analyzeContext(ctx, content, nil, true)
}

// ImportNames returns the domains directly imported by this analysis's source.
func (p *Analysis) ImportNames() []string {
	names := make([]string, 0, len(p.content.Imports))
	for _, importDecl := range p.content.Imports {
		names = append(names, importDecl.Domain.String())
	}
	return names
}

func newAnalysis(content *grammar.SkelContent) *Analysis {
	return &Analysis{
		name: "",

		content: content,

		enumsMap:     map[string]*schema.Enum{},
		dataMap:      map[string]*schema.Data{},
		actorsMap:    map[string]*schema.Actor{},
		resourcesMap: map[string]*schema.Resource{},
		websMap:      map[string]*schema.Web{},
		servicesMap:  map[string]*schema.Service{},
		tasksMap:     map[string]*schema.Task{},
		importsMap:   map[string]*_DomainImport{},

		reporter:    newDiagnosticReporter(),
		invalidData: map[*schema.Data]bool{},
		unavailable: map[string]bool{},
	}
}

func (p *Analysis) Schema() *schema.Domain {
	if p.schema != nil {
		return p.schema
	}
	p.schema = schema.NewDomainFromSpec(schema.DomainSpec{
		Name: p.name, Description: p.description,
		Imports: p.imports, Enums: p.enums, Data: p.dataList, Configs: p.configs, Events: p.events,
		Actors: p.actors, Resources: p.resources, Webs: p.webs, Services: p.services, Tasks: p.tasks,
	})
	return p.schema
}

func (p *Analysis) Warnings() []string {
	return append([]string{}, p.warnings...)
}
