package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
)

func (g *_Gen) buildDomainDescriptor() *descriptor.Domain {
	domainView := g.view
	if g.isSplitRegular() {
		domainView = view.Full(g.Domain)
	}
	result := &descriptor.Domain{
		Domain: g.Domain.Name(), Description: g.Domain.Description(), Hash: g.Domain.Hash(), Full: !g.isSplitPub(),
		Enums: make([]*descriptor.Enum, 0, len(domainView.Enums)), Data: make([]*descriptor.Data, 0, len(domainView.Data)),
		Configs: make([]*descriptor.Config, 0, len(domainView.Configs)), Webs: make([]*descriptor.Web, 0, len(domainView.Webs)),
		Events: make([]*descriptor.Event, 0, len(domainView.Events)), Actors: make([]*descriptor.Actor, 0, len(domainView.Actors)),
		Resources: make([]*descriptor.Resource, 0, len(domainView.Resources)), Services: make([]*descriptor.Service, 0, len(domainView.Services)),
		Tasks: make([]*descriptor.Task, 0, len(domainView.Tasks)), Generated: &descriptor.GeneratedInfo{CompilerVersion: g.compilerVersion},
	}
	for _, value := range domainView.Enums {
		result.Enums = append(result.Enums, g.buildEnumDescriptor(value))
	}
	for _, value := range domainView.Data {
		result.Data = append(result.Data, g.buildDataDescriptor(value))
	}
	for _, value := range domainView.Configs {
		result.Configs = append(result.Configs, g.buildConfigDescriptor(value))
	}
	for _, value := range domainView.Webs {
		result.Webs = append(result.Webs, g.buildWebDescriptor(value))
	}
	for _, value := range domainView.Events {
		result.Events = append(result.Events, g.buildEventDescriptor(value))
	}
	for _, value := range domainView.Actors {
		result.Actors = append(result.Actors, g.buildActorDescriptor(value))
	}
	for _, value := range domainView.Resources {
		result.Resources = append(result.Resources, g.buildResourceDescriptor(value))
	}
	for _, value := range domainView.Services {
		result.Services = append(result.Services, g.buildServiceDescriptor(value))
	}
	for _, value := range domainView.Tasks {
		result.Tasks = append(result.Tasks, g.buildTaskDescriptor(value))
	}
	return result
}
