package skeleton

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/schema"
)

type _SkelPayload struct {
	Domain      *schema.Domain
	Imports     []*schema.Import
	Actors      []*schema.Actor
	Enums       []*schema.Enum
	Data        []*schema.Data
	Configs     []*schema.Data
	Resources   []*schema.Resource
	Events      []*schema.Data
	Services    []*schema.Service
	Description string
}

func (g *_Gen) buildDomainPayload(imports []*schema.Import) *_SkelPayload {
	return &_SkelPayload{
		Domain:      g.domain,
		Imports:     imports,
		Description: g.domain.Description(),
	}
}

func (g *_Gen) buildActorPayload(actors []*schema.Actor) *_SkelPayload {
	payload := g.buildDomainPayload(collectActorImports(g.domain, actors))
	payload.Actors = actors
	return payload
}

func (g *_Gen) buildTypesPayload(view *codegen.PublicView) *_SkelPayload {
	payload := g.buildDomainPayload(collectTypeImports(g.domain, view))
	payload.Enums = view.Enums
	payload.Data = view.Data
	payload.Configs = view.Configs
	payload.Resources = view.Resources
	return payload
}

func (g *_Gen) buildEventPayload(events []*schema.Data) *_SkelPayload {
	payload := g.buildDomainPayload(collectDataImports(g.domain, events))
	payload.Events = events
	return payload
}

func (g *_Gen) buildServicePayload(services []*schema.Service) *_SkelPayload {
	payload := g.buildDomainPayload(collectServiceImports(g.domain, services))
	payload.Services = services
	return payload
}
