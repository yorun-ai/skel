package vineschema

import (
	"fmt"
	"go.yorun.ai/skel/internal/projection"
	contractschema "go.yorun.ai/skel/internal/schema"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
)

func (g *_Gen) buildDomainSchema() (*_DomainSchema, error) {
	document, err := projection.Project(g.Domain, nil)
	if err != nil {
		return nil, err
	}
	domainView := g.view
	if g.isSplitRegular() {
		domainView = view.Full(g.Domain)
	}
	result := &_DomainSchema{
		Domain: g.Domain.Name(), Description: document.Description, Hash: g.Domain.Hash(), Full: !g.isSplitPub(),
		Enums: make([]*_EnumSchema, 0, len(domainView.Enums)), Data: make([]*_DataSchema, 0, len(domainView.Data)),
		Configs: make([]*_ConfigSchema, 0, len(domainView.Configs)), Webs: make([]*_WebSchema, 0, len(domainView.Webs)),
		Events: make([]*_EventSchema, 0, len(domainView.Events)), Actors: make([]*_ActorSchema, 0, len(domainView.Actors)),
		Resources: make([]*_ResourceSchema, 0, len(domainView.Resources)), Services: make([]*_ServiceSchema, 0, len(domainView.Services)),
		Tasks: make([]*_TaskSchema, 0, len(domainView.Tasks)), Generated: &_GeneratedInfo{CompilerVersion: g.compilerVersion},
	}
	for _, value := range domainView.Enums {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeEnum, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Enums = append(result.Enums, g.buildEnumSchema(value, projected))
	}
	for _, value := range domainView.Data {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeData, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Data = append(result.Data, g.buildDataSchema(value, projected))
	}
	for _, value := range domainView.Configs {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeConfig, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Configs = append(result.Configs, g.buildConfigSchema(value, projected))
	}
	for _, value := range domainView.Webs {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeWeb, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Webs = append(result.Webs, g.buildWebSchema(value, projected))
	}
	for _, value := range domainView.Events {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeEvent, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Events = append(result.Events, g.buildEventSchema(value, projected))
	}
	for _, value := range domainView.Actors {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeActor, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Actors = append(result.Actors, g.buildActorSchema(value, projected))
	}
	for _, value := range domainView.Resources {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeResource, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Resources = append(result.Resources, g.buildResourceSchema(value, projected))
	}
	for _, value := range domainView.Services {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeService, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Services = append(result.Services, g.buildServiceSchema(value, projected))
	}
	for _, value := range domainView.Tasks {
		projected, findErr := findProjected(document, contractschema.DeclarationTypeTask, value.SkelName, value.Name)
		if findErr != nil {
			return nil, findErr
		}
		result.Tasks = append(result.Tasks, g.buildTaskSchema(value, projected))
	}
	return result, nil
}

func findProjected(document *contractschema.Document, kind contractschema.DeclarationType, skelName, name string) (*contractschema.Declaration, error) {
	if skelName != "" {
		if declaration := contractschema.Find(document, string(kind), skelName); declaration != nil {
			return declaration, nil
		}
	}
	for _, declaration := range document.Declarations {
		if declaration.Kind == kind && declaration.Name == name {
			return declaration, nil
		}
	}
	return nil, fmt.Errorf("normalized schema has no %s declaration %s", kind, name)
}
