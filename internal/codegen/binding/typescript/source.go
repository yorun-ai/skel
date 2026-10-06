package typescript

import (
	"strings"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/schema"
)

const packageScope = "@yorun-ai/skeled"

type _Gen struct {
	types    _Types
	domain   *schema.Domain
	bindings binding.TypeBindings

	moduleScope string
	pkgName     string
	tsImports   map[string]string
	outputDir   string
	err         error
	apiView     *codegen.PublicView

	renderer *binding.Renderer
}

type _SourceResult struct {
	PackageName     string
	ResolvedImports map[string]string
}

// generateSource renders a domain already checked by codegen.ValidateDomain.
func generateSource(domain codegen.Input, outputDir string, option Option, sink binding.FileSink) (_SourceResult, error) {
	selected, err := domain.Select(codegen.Selection{Surface: codegen.SurfaceAPI, API: option.ApiFilter})
	if err != nil {
		return _SourceResult{}, err
	}
	gen := newGen(selected, outputDir, option, sink)
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

func newGen(input codegen.Input, outputDir string, option Option, sink binding.FileSink) *_Gen {
	domain := input.Schema()
	g := &_Gen{
		domain:      domain,
		moduleScope: strings.TrimRight(option.ModuleScope, "/"),
		pkgName:     strings.TrimRight(option.Module, "/"),
		tsImports:   option.Imports,
		outputDir:   outputDir,
		renderer:    binding.NewRendererWithSink(outputDir, sink),
	}
	if g.pkgName == "" {
		scope := g.moduleScope
		if scope == "" {
			scope = packageScope
		}
		g.pkgName = buildPackageName(scope, g.domain.Name())
	}
	decls := input.Declarations()
	g.apiView = &codegen.PublicView{Enums: decls.Enums, Data: decls.Data, Services: decls.Services}
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
