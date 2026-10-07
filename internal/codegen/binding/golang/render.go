// Package golang coordinates Go generation and owns target options, imports and
// module metadata. Source rendering and Vine schema adaptation are subpackages.
package golang

import (
	"strings"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/binding/golang/descriptor"
	"go.yorun.ai/skel/internal/codegen/binding/golang/source"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/schema"
)

type _Gen struct {
	domain   *schema.Domain
	view     *view.Domain
	bindings binding.TypeBindings

	mode              view.Mode
	modName           string
	pkgName           string
	asModule          bool
	compilerVersion   string
	vineVersion       string
	vrpcVersion       string
	modulePrefix      string
	goImports         map[string]string
	pubImportPath     string
	extraDependencies []string
	out               string
	sink              binding.FileSink
}

func render(validated codegen.Input, resolved ResolvedOption, files *binding.FileCollector) error {
	option := resolved.option
	domain := validated.Schema()
	if option.ApiOnly || option.PubOnly {
		mode := view.ModePub
		if option.ApiOnly {
			mode = view.ModeApi
		}
		g, err := newGen(_GenOption{Input: validated, Sink: files.Sink(""), ApiFilter: option.ApiFilter, Mode: mode, Out: option.Out, AsModule: option.AsModule, Module: option.Module, ModulePrefix: option.ModulePrefix, Imports: option.Imports, VineVersion: option.VineVersion, VrpcVersion: option.VrpcVersion, CompilerVersion: option.CompilerVersion})
		if err != nil {
			return err
		}
		return g.gen(validated)
	}
	if option.PubOut == "" {
		gen, err := newGen(_GenOption{
			Input: validated, Sink: files.Sink(""),
			CompilerVersion: option.CompilerVersion,
			ModulePrefix:    option.ModulePrefix,
			Module:          option.Module,
			VineVersion:     option.VineVersion,
			Imports:         option.Imports,
			Mode:            view.ModeFull,
			Out:             option.Out,
			AsModule:        option.AsModule,
		})
		if err != nil {
			return err
		}
		return gen.gen(validated)
	}

	pubModule := option.PubModule
	if pubModule == "" {
		pubModule = DefaultPubModuleName(option.ModulePrefix, option.Module, domain.Name())
	}
	pubGen, err := newGen(_GenOption{
		Input: validated, Sink: files.Sink("pub"),
		CompilerVersion: option.CompilerVersion,
		ModulePrefix:    option.ModulePrefix,
		Module:          pubModule,
		VineVersion:     option.VineVersion,
		Imports:         option.Imports,
		Mode:            view.ModePub,
		Out:             option.PubOut,
		AsModule:        true,
	})
	if err != nil {
		return err
	}
	if err := pubGen.gen(validated); err != nil {
		return err
	}
	regularGen, err := newGen(_GenOption{
		Input: validated, Sink: files.Sink(""),
		CompilerVersion:   option.CompilerVersion,
		ModulePrefix:      option.ModulePrefix,
		Module:            option.Module,
		VineVersion:       option.VineVersion,
		Imports:           option.Imports,
		Out:               option.Out,
		AsModule:          true,
		Mode:              view.ModeRegular,
		PubImportPath:     pubModule,
		ExtraDependencies: []string{pubModule},
	})
	if err != nil {
		return err
	}
	return regularGen.gen(validated)
}

func newGen(option _GenOption) (*_Gen, error) {
	// Backend modes historically validate all mappings. API mode validates
	// only its selected dependencies when generating the module below.
	if option.AsModule && option.Mode != view.ModeApi {
		if _, err := moduleDependencies(_ModuleOption{
			Imports: option.Imports, VineVersion: option.VineVersion,
			CompilerVersion: option.CompilerVersion,
			Api:             option.Mode == view.ModeApi, VrpcVersion: option.VrpcVersion,
			ExtraDependencies: option.ExtraDependencies,
		}); err != nil {
			return nil, err
		}
	}
	g := &_Gen{
		domain:            option.Input.Schema(),
		sink:              option.Sink,
		mode:              option.Mode,
		asModule:          option.AsModule,
		compilerVersion:   option.CompilerVersion,
		vineVersion:       option.VineVersion,
		vrpcVersion:       option.VrpcVersion,
		modulePrefix:      option.ModulePrefix,
		goImports:         option.Imports,
		pubImportPath:     option.PubImportPath,
		extraDependencies: option.ExtraDependencies,
		out:               option.Out,
	}
	var err error
	g.view, err = view.FromInput(option.Mode, option.Input, option.ApiFilter)
	if err != nil {
		return nil, err
	}
	domainParts := strings.Split(g.domain.Name(), ".")
	if option.AsModule {
		g.modName = option.Module
		if g.modName == "" {
			g.modName = buildModuleName(g.modulePrefix, domainParts, option.Mode == view.ModePub)
			if option.Mode == view.ModeApi {
				g.modName += "api"
			}
		}
	}
	fallback := packageNameFallback(domainParts, option.Mode == view.ModePub)
	if option.Mode == view.ModeApi {
		fallback += "api"
	}
	g.pkgName, err = inferPackageName(option.Out, fallback, option.AsModule)
	if err != nil {
		return nil, err
	}
	if err := g.resolveExternalTypeImports(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *_Gen) gen(validated codegen.Input) error {
	if g.asModule {
		imports, err := g.usedModuleImports()
		if err != nil {
			return err
		}
		if err := generateModule(_ModuleOption{
			Out:               g.out,
			Sink:              g.sink,
			Module:            g.modName,
			VineVersion:       g.vineVersion,
			Api:               g.mode == view.ModeApi,
			VrpcVersion:       g.vrpcVersion,
			CompilerVersion:   g.compilerVersion,
			Imports:           imports,
			ExtraDependencies: g.extraDependencies,
		}); err != nil {
			return err
		}
	}
	if err := source.GenerateValidated(validated, source.Option{
		Domain:        g.domain,
		Bindings:      g.bindings,
		View:          g.view,
		Mode:          g.mode,
		PackageName:   g.pkgName,
		PubImportPath: g.pubImportPath,
		Out:           g.out,
		Sink:          g.sink,
	}); err != nil {
		return err
	}
	if g.mode == view.ModeApi {
		return nil
	}
	return descriptor.GenerateValidated(validated, descriptor.Option{
		Domain:          g.domain,
		View:            g.view,
		Mode:            g.mode,
		PackageName:     g.pkgName,
		CompilerVersion: g.compilerVersion,
		Out:             g.out,
		Sink:            g.sink,
	})
}
