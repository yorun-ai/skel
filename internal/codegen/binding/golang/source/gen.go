package source

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/schema"
)

type _Gen struct {
	types  _Types
	Domain *schema.Domain

	mode          view.Mode
	pkgName       string
	pubImportPath string
	view          *view.Domain

	Renderer *binding.Renderer
}

type Option struct {
	Sink          binding.FileSink
	Bindings      binding.TypeBindings
	Domain        *schema.Domain
	View          *view.Domain
	Mode          view.Mode
	PackageName   string
	PubImportPath string
	Out           string
}

// GenerateValidated renders a domain already checked by codegen.ValidateDomain.
func GenerateValidated(domain codegen.Input, option Option) error {
	option.Domain = domain.Schema()
	gen := newGen(option)
	if err := gen.validateSensitiveMembers(); err != nil {
		return err
	}
	gen.gen()
	return gen.Renderer.Err()
}

func newGen(option Option) *_Gen {
	return &_Gen{
		types:         _Types{bindings: option.Bindings},
		Domain:        option.Domain,
		mode:          option.Mode,
		pkgName:       option.PackageName,
		pubImportPath: option.PubImportPath,
		view:          option.View,
		Renderer:      binding.NewRendererWithSink(option.Out, option.Sink),
	}
}

func (g *_Gen) gen() {
	if g.mode == view.ModeApi {
		g.genApiGo()
		return
	}
	g.genDocGo()
	g.genEnumGo()
	g.genDataGo()
	g.genConfigGo()
	g.genActorGo()
	g.genWebGo()
	g.genEventGo()
	g.genResourceGo()
	g.genServiceGo()
	g.genTaskGo()
	g.genFacadeGo()
}

func (g *_Gen) isSplitPub() bool {
	return g.mode == view.ModePub
}

func (g *_Gen) isSplitRegular() bool {
	return g.mode == view.ModeRegular
}
