package source

import (
	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
)

type _Gen struct {
	types  _Types
	Domain *model.Domain

	mode          view.Mode
	pkgName       string
	pubImportPath string
	view          *view.Domain

	Renderer *common.Renderer
}

type Option struct {
	Bindings      common.TypeBindings
	Domain        *model.Domain
	View          *view.Domain
	Mode          view.Mode
	PackageName   string
	PubImportPath string
	Out           string
}

// GenerateValidated renders a domain already checked by common.ValidateDomain.
func GenerateValidated(domain common.ValidatedDomain, option Option) error {
	option.Domain = domain.Model()
	gen := newGen(option)
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
		Renderer:      common.NewRenderer(option.Out),
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
