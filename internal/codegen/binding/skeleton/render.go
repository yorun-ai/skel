package skeleton

import (
	"fmt"
	"text/template"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/schema"
)

type _Gen struct {
	domain *schema.Domain
	input  codegen.Input

	renderer *binding.Renderer
	template *template.Template
}

func render(validated codegen.Input, option Option, sink binding.FileSink) error {
	if !option.PubOnly {
		return fmt.Errorf("Skel generation requires public-only output")
	}

	selected, err := validated.Select(codegen.Selection{Surface: codegen.SurfacePublic})
	if err != nil {
		return err
	}
	gen, err := newGen(selected, option.Out, sink)
	if err != nil {
		return err
	}
	if err := gen.generate(); err != nil {
		return err
	}
	return gen.renderer.Err()
}

func newGen(input codegen.Input, outputDir string, sink binding.FileSink) (*_Gen, error) {
	tpl, err := newSkelTemplate()
	if err != nil {
		return nil, err
	}
	return &_Gen{
		domain:   input.Schema(),
		input:    input,
		renderer: binding.NewRendererWithSink(outputDir, sink),
		template: tpl,
	}, nil
}

func (g *_Gen) generate() error {
	d := g.input.Declarations()
	view := &codegen.PublicView{Enums: d.Enums, Data: d.Data, Configs: d.Configs, Events: d.Events, Actors: d.Actors, Resources: d.Resources, Services: d.Services}
	if err := g.render("domain.skel", "domain.skel.tpl", g.buildDomainPayload(nil)); err != nil {
		return err
	}
	if len(view.Actors) > 0 {
		if err := g.render("actor.skel", "actor.skel.tpl", g.buildActorPayload(view.Actors)); err != nil {
			return err
		}
	}
	if len(view.Enums) > 0 || len(view.Data) > 0 || len(view.Configs) > 0 || len(view.Resources) > 0 {
		if err := g.render("types.skel", "types.skel.tpl", g.buildTypesPayload(view)); err != nil {
			return err
		}
	}
	if len(view.Events) > 0 {
		if err := g.render("event.skel", "event.skel.tpl", g.buildEventPayload(view.Events)); err != nil {
			return err
		}
	}
	if len(view.Services) > 0 {
		if err := g.render("service.skel", "service.skel.tpl", g.buildServicePayload(view.Services)); err != nil {
			return err
		}
	}
	return nil
}

func (g *_Gen) render(file, templateName string, payload *_SkelPayload) error {
	content, err := g.renderSkel(templateName, payload)
	if err != nil {
		return err
	}
	g.renderer.Write(file, content)
	return g.renderer.Err()
}
