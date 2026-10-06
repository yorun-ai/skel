package source

import "strings"

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
	if !g.isSplitRegular() {
		return
	}
	public := g.view.Reexports
	if len(public.Enums)+len(public.Data)+len(public.Configs)+len(public.Actors)+
		len(public.Resources)+len(public.Services)+len(public.Events) == 0 {
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
	for _, enum := range public.Enums {
		payload.Enums = append(payload.Enums, castEnum(enum))
	}
	for _, data := range public.Data {
		payload.Data = append(payload.Data, g.types.castData(data))
	}
	for _, config := range public.Configs {
		payload.Configs = append(payload.Configs, g.types.castData(config))
	}
	for _, actor := range public.Actors {
		payload.Actors = append(payload.Actors, castActor(actor))
		if actor.Auth != nil {
			payload.AuthCredentialData = append(payload.AuthCredentialData, g.types.castData(actor.Auth.Credential), g.types.castData(actor.Auth.Info))
			payload.AuthServices = append(payload.AuthServices, g.types.castActorAuthService(actor.Auth.Service))
		}
		if actor.Permission != nil {
			payload.AuthServices = append(payload.AuthServices, g.types.castActorAuthService(actor.Permission.Service))
		}
	}
	for _, resource := range public.Resources {
		casted := castResource(resource)
		payload.Resources = append(payload.Resources, casted)
		if resource.CheckService != nil {
			payload.AuthServices = append(payload.AuthServices, g.castService(resource.CheckService, false, true))
		}
	}
	for _, service := range public.Services {
		payload.Services = append(payload.Services, g.castService(service, true, false))
	}
	for _, event := range public.Events {
		payload.Events = append(payload.Events, g.castEvent(event, true, false))
	}

	g.renderGo(facadeGoFilename, facadeGoTemplate, payload)
}
