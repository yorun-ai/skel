package source

import "go.yorun.ai/skel/schema"

const actorGoFilename = "actor.go"

var actorImports = []*Import{
	{Path: "go.yorun.ai/vine/core/meta"},
}

var actorInfoImports = []*Import{
	{Path: "reflect"},
}

var actorGoTemplate = joinTemplates(
	"imports.go.tpl",
	"actor.go.tpl",
	"service/info.go.tpl",
	"service/arguments.go.tpl",
	"service/server.go.tpl",
	"service/er_server.go.tpl",
	"service/client.go.tpl",
	"service/er_client.go.tpl",
)

type ActorGoPayload struct {
	PackageName    string
	StdImports     []*Import
	ModuleImports  []*Import
	Actors         []*Actor
	CredentialData []*Data
	AuthServices   []*Service
	HasActorInfo   bool
}

type Actor struct {
	Name             string
	SkelName         string
	Hash             string
	AuthInfoName     string
	AuthInfoSkelName string
	HasInfo          bool
}

func (g *_Gen) genActorGo() {
	payload := &ActorGoPayload{
		PackageName:    g.pkgName,
		Actors:         make([]*Actor, 0, len(g.view.Actors)),
		CredentialData: make([]*Data, 0),
		AuthServices:   make([]*Service, 0),
	}
	for _, tokenActor := range g.view.Actors {
		actor := castActor(tokenActor)
		payload.Actors = append(payload.Actors, actor)
		payload.HasActorInfo = payload.HasActorInfo || actor.HasInfo
	}
	for _, tokenActor := range g.authServiceActors() {
		if tokenActor.Auth != nil {
			info := g.types.castData(tokenActor.Auth.Info)
			for _, member := range info.Members {
				member.Identifier = member.SkelName == tokenActor.Auth.IdentifierField
			}
			payload.CredentialData = append(
				payload.CredentialData,
				g.types.castData(tokenActor.Auth.Credential),
				info,
			)
			payload.AuthServices = append(payload.AuthServices, g.types.castActorAuthService(tokenActor.Auth.Service))
		}
		if tokenActor.Permission != nil {
			payload.AuthServices = append(payload.AuthServices, g.types.castActorAuthService(tokenActor.Permission.Service))
		}
	}
	if len(payload.Actors) == 0 && len(payload.AuthServices) == 0 {
		return
	}

	imports := newImportSet()
	if len(payload.Actors) > 0 {
		imports.addMany(actorImports)
		if payload.HasActorInfo {
			imports.addMany(actorInfoImports)
		}
	}
	if len(payload.AuthServices) > 0 {
		imports.addMany(serviceImports)
		imports.addMany(buildDataImports(payload.CredentialData))
		imports.addMany(buildServiceImports(payload.AuthServices))
	}
	payload.StdImports, payload.ModuleImports = splitImports(imports.sortedValues())
	g.renderGo(actorGoFilename, actorGoTemplate, payload)
}

func (g *_Gen) authServiceActors() []*schema.Actor {
	if g.isSplitPub() || g.isSplitRegular() {
		return g.view.Actors
	}
	return g.Domain.Actors()
}

func castActor(p *schema.Actor) *Actor {
	actor := &Actor{
		Name:     p.Name,
		SkelName: p.SkelName,
		Hash:     p.Hash,
	}
	if p.Auth != nil {
		actor.AuthInfoName = p.Auth.Info.Name
		actor.AuthInfoSkelName = p.Auth.Info.SkelName
		actor.HasInfo = true
	}
	return actor
}
