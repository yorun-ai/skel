package source

import (
	"strings"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/schema"
)

const facadeGoFilename = "pub.go"

var facadeGoTemplate = loadTemplate("facade.go.tpl")

func importPackageName(domainName string, usePubPackage bool) string {
	parts := strings.Split(domainName, ".")
	name := parts[len(parts)-1]
	if usePubPackage {
		return name + "pub"
	}
	return name
}

type FacadeGoPayload struct {
	PackageName        string
	PubImport          *Import
	PubPackageName     string
	Enums              []*Enum
	Data               []*Data
	Configs            []*Data
	Actors             []*Actor
	AuthCredentialData []*Data
	AuthServices       []*Service
	Resources          []*Resource
	Services           []*Service
	Events             []*Event
}

func (g *_Gen) genFacadeGo() {
	if !g.isSplitRegular() || !g.hasPubSymbols() {
		return
	}

	payload := &FacadeGoPayload{
		PackageName: g.pkgName,
		PubImport: &Import{
			Path: g.pubImportPath,
		},
		PubPackageName:     importPackageName(g.Domain.Name(), true),
		Enums:              make([]*Enum, 0),
		Data:               make([]*Data, 0),
		Configs:            make([]*Data, 0),
		Actors:             make([]*Actor, 0),
		AuthCredentialData: make([]*Data, 0),
		AuthServices:       make([]*Service, 0),
		Resources:          make([]*Resource, 0),
		Services:           make([]*Service, 0),
		Events:             make([]*Event, 0),
	}
	public, err := codegen.BuildPublicView(g.Domain)
	if err != nil {
		g.Renderer.Fail(err)
		return
	}
	for _, enum := range public.Enums {
		payload.Enums = append(payload.Enums, castEnum(enum))
	}
	for _, data := range public.Data {
		payload.Data = append(payload.Data, g.types.castData(data))
	}
	for _, config := range g.Domain.Configs() {
		if config.Pub {
			payload.Configs = append(payload.Configs, g.types.castData(config))
		}
	}
	for _, actor := range g.Domain.Actors() {
		if actor.Pub {
			payload.Actors = append(payload.Actors, castActor(actor))
			if actor.Auth != nil {
				payload.AuthCredentialData = append(payload.AuthCredentialData, g.types.castData(actor.Auth.Credential), g.types.castData(actor.Auth.Info))
				payload.AuthServices = append(payload.AuthServices, g.types.castActorAuthService(actor.Auth.Service))
			}
			if actor.Permission != nil {
				payload.AuthServices = append(payload.AuthServices, g.types.castActorAuthService(actor.Permission.Service))
			}
		}
	}
	for _, resource := range g.Domain.Resources() {
		if resource.Pub {
			casted := castResource(resource)
			payload.Resources = append(payload.Resources, casted)
			if resource.CheckService != nil {
				payload.AuthServices = append(payload.AuthServices, g.castService(resource.CheckService, false, true))
			}
		}
	}
	for _, service := range g.Domain.Services() {
		if service.Public() {
			payload.Services = append(payload.Services, g.castService(service, true, false))
		}
	}
	for _, event := range g.Domain.Events() {
		if event.Public() {
			payload.Events = append(payload.Events, g.castEvent(event, true, false))
		}
	}

	g.renderGo(facadeGoFilename, facadeGoTemplate, payload)
}

func (g *_Gen) hasPubSymbols() bool {
	return hasPubEnum(g.Domain.Enums()) ||
		hasPubData(g.Domain.Data()) ||
		hasPubData(g.Domain.Configs()) ||
		hasPubActor(g.Domain.Actors()) ||
		hasPubResource(g.Domain.Resources()) ||
		hasPubService(g.Domain.Services()) ||
		hasPubData(g.Domain.Events())
}

func hasPubEnum(enums []*schema.Enum) bool {
	for _, enum := range enums {
		if enum.Pub {
			return true
		}
	}
	return false
}

func hasPubData(dataList []*schema.Data) bool {
	for _, data := range dataList {
		if data.Public() {
			return true
		}
	}
	return false
}

func hasPubResource(resources []*schema.Resource) bool {
	for _, resource := range resources {
		if resource.Pub {
			return true
		}
	}
	return false
}

func hasPubActor(actors []*schema.Actor) bool {
	for _, actor := range actors {
		if actor.Pub {
			return true
		}
	}
	return false
}

func hasPubService(services []*schema.Service) bool {
	for _, service := range services {
		if service.Public() {
			return true
		}
	}
	return false
}
