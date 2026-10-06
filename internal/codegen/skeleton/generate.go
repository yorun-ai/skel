package skeleton

import (
	"fmt"
	"text/template"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
)

type _Gen struct {
	domain *model.Domain

	renderer *common.Renderer
	template *template.Template
}

func Generate(domain *model.Domain, option Option) error {
	if !option.PubOnly {
		return fmt.Errorf("Skel generation requires public-only output")
	}
	validated, err := common.PrepareDomain(domain)
	if err != nil {
		return fmt.Errorf("validate Skel generation model: %w", err)
	}
	gen, err := newGen(validated.Model(), option.Out)
	if err != nil {
		return err
	}
	if err := gen.generate(); err != nil {
		return err
	}
	return gen.renderer.Err()
}

func newGen(domain *model.Domain, outputDir string) (*_Gen, error) {
	tpl, err := newSkelTemplate()
	if err != nil {
		return nil, err
	}
	return &_Gen{
		domain:   domain,
		renderer: common.NewRenderer(outputDir),
		template: tpl,
	}, nil
}

func (g *_Gen) generate() error {
	view, err := common.BuildPublicView(g.domain)
	if err != nil {
		return err
	}
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
