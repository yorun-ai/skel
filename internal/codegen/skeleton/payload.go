package skeleton

import (
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
)

type _SkelPayload struct {
	Domain      *model.Domain
	Imports     []*model.Import
	Actors      []*model.Actor
	Enums       []*model.Enum
	Data        []*model.Data
	Configs     []*model.Data
	Resources   []*model.Resource
	Events      []*model.Data
	Services    []*model.Service
	Description string
}

func (g *_Gen) buildDomainPayload(imports []*model.Import) *_SkelPayload {
	return &_SkelPayload{
		Domain:      g.domain,
		Imports:     imports,
		Description: g.domain.Description(),
	}
}

func (g *_Gen) buildActorPayload(actors []*model.Actor) *_SkelPayload {
	payload := g.buildDomainPayload(collectActorImports(g.domain, actors))
	payload.Actors = actors
	return payload
}

func (g *_Gen) buildTypesPayload(view *common.PublicView) *_SkelPayload {
	payload := g.buildDomainPayload(collectTypeImports(g.domain, view))
	payload.Enums = view.Enums
	payload.Data = view.Data
	payload.Configs = view.Configs
	payload.Resources = view.Resources
	return payload
}

func (g *_Gen) buildEventPayload(events []*model.Data) *_SkelPayload {
	payload := g.buildDomainPayload(collectDataImports(g.domain, events))
	payload.Events = events
	return payload
}

func (g *_Gen) buildServicePayload(services []*model.Service) *_SkelPayload {
	payload := g.buildDomainPayload(collectServiceImports(g.domain, services))
	payload.Services = services
	return payload
}
