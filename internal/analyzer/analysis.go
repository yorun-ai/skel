package analyzer

import (
	"context"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

type Analysis struct {
	name        string
	description string
	model       *model.Domain

	imports []*model.Import

	enums     []*model.Enum
	dataList  []*model.Data
	configs   []*model.Data
	events    []*model.Data
	actors    []*model.Actor
	resources []*model.Resource
	webs      []*model.Web
	services  []*model.Service
	tasks     []*model.Task

	warnings []string

	content *grammar.SkelContent

	enumsMap     map[string]*model.Enum
	dataMap      map[string]*model.Data
	actorsMap    map[string]*model.Actor
	resourcesMap map[string]*model.Resource
	websMap      map[string]*model.Web
	servicesMap  map[string]*model.Service
	tasksMap     map[string]*model.Task
	importsMap   map[string]*_DomainImport

	reporter    *_DiagnosticReporter
	invalidData map[*model.Data]bool
	unavailable map[string]bool
}

type _DomainImport struct {
	Domain *Analysis
	Model  *model.Import
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
		domainByName[importedDomain.Model().Name()] = importedDomain
	}
	if !domain.load() {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		domain.reporter.ctx = nil
		return domain, domain.reporter.result(), nil
	}
	diagnosticsBeforeImports := len(domain.reporter.errors)
	domain.loadImports(domainByName)
	if ctx.Err() == nil && len(domain.reporter.errors) == diagnosticsBeforeImports {
		domain.normalize()
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
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	domain := newAnalysis(content)
	domain.reporter.ctx = ctx
	if domain.load() && ctx.Err() == nil {
		domain.normalizeImport()
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

// ResolveImports reanalyzes an import-only domain with its complete set of
// direct dependencies. Callers use this after loading the transitive graph.
func (p *Analysis) ResolveImports(importedDomains []*Analysis) (*Analysis, []error) {
	return Analyze(p.content, importedDomains)
}

// ImportNames returns the domains directly imported by this analysis's source.
func (p *Analysis) ImportNames() []string {
	names := make([]string, 0, len(p.content.Imports))
	for _, importDecl := range p.content.Imports {
		names = append(names, importDecl.Domain.String())
	}
	return names
}

// ImportAliases returns source qualifiers mapped to fully qualified imported
// domain names without requiring those domains to be loaded.
func (p *Analysis) ImportAliases() map[string]string {
	aliases := make(map[string]string, len(p.content.Imports))
	for _, importDecl := range p.content.Imports {
		domainName := importDecl.Domain.String()
		alias := domainName
		if importDecl.Alias != nil {
			alias = importDecl.Alias.Value
		}
		aliases[alias] = domainName
	}
	return aliases
}

func newAnalysis(content *grammar.SkelContent) *Analysis {
	return &Analysis{
		name: "",

		content: content,

		enumsMap:     map[string]*model.Enum{},
		dataMap:      map[string]*model.Data{},
		actorsMap:    map[string]*model.Actor{},
		resourcesMap: map[string]*model.Resource{},
		websMap:      map[string]*model.Web{},
		servicesMap:  map[string]*model.Service{},
		tasksMap:     map[string]*model.Task{},
		importsMap:   map[string]*_DomainImport{},

		reporter:    newDiagnosticReporter(),
		invalidData: map[*model.Data]bool{},
		unavailable: map[string]bool{},
	}
}

func (p *Analysis) Model() *model.Domain {
	if p.model != nil {
		return p.model
	}
	p.model = model.NewDomainFromSpec(model.DomainSpec{
		Name: p.name, Description: p.description,
		Imports: p.imports, Enums: p.enums, Data: p.dataList, Configs: p.configs, Events: p.events,
		Actors: p.actors, Resources: p.resources, Webs: p.webs, Services: p.services, Tasks: p.tasks,
	})
	return p.model
}

func (p *Analysis) Warnings() []string {
	return append([]string{}, p.warnings...)
}
