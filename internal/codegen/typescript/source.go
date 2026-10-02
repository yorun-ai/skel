package typescript

import (
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

const packageScope = "@yorun-ai/skeled"

type _Gen struct {
	types    _Types
	domain   *model.Domain
	bindings common.TypeBindings

	moduleScope string
	pkgName     string
	tsImports   map[string]string
	outputDir   string
	err         error
	apiView     *common.PublicView

	renderer *common.Renderer
}

type _SourceResult struct {
	PackageName     string
	ResolvedImports map[string]string
}

// generateSource renders a domain already checked by common.ValidateDomain.
func generateSource(domain common.ValidatedDomain, outputDir string, option Option) (_SourceResult, error) {
	gen := newGen(domain.Model(), outputDir, option)
	if gen.err != nil {
		return _SourceResult{}, gen.err
	}
	gen.generate()
	resolvedImports := gen.resolvedModuleImports()
	if gen.err != nil {
		return _SourceResult{}, gen.err
	}
	return _SourceResult{
		PackageName:     gen.pkgName,
		ResolvedImports: resolvedImports,
	}, gen.renderer.Err()
}

func newGen(domain *model.Domain, outputDir string, options ...Option) *_Gen {
	option := Option{}
	if len(options) > 0 {
		option = options[0]
	}
	g := &_Gen{
		domain:      domain,
		moduleScope: strings.TrimRight(option.ModuleScope, "/"),
		pkgName:     strings.TrimRight(option.Module, "/"),
		tsImports:   option.Imports,
		outputDir:   outputDir,
		renderer:    common.NewRenderer(outputDir),
	}
	if g.pkgName == "" {
		scope := g.moduleScope
		if scope == "" {
			scope = packageScope
		}
		g.pkgName = buildPackageName(scope, g.domain.Name())
	}
	g.apiView = common.BuildApiView(domain)
	g.resolveExternalTypeImports()
	g.types = _Types{bindings: g.bindings}
	return g
}

func (g *_Gen) generate() {
	if g.hasApiDeclarations() {
		g.genDataTs()
		g.genSpecTs()
		g.genServiceTs()
	}
	g.genIndex()
}

func (g *_Gen) hasApiDeclarations() bool {
	return len(g.apiView.Enums) > 0 || len(g.apiView.Data) > 0 || len(g.apiView.Services) > 0
}
