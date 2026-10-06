package vineschema

import (
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/projection"
	contractschema "go.yorun.ai/skel/internal/schema"
)

func (g *_Gen) buildActorSchema(value *model.Actor, projected *contractschema.Declaration) *_ActorSchema {
	result := &_ActorSchema{
		IdentifierField: projected.Actor.IdentifierField, Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason, Vias: make([]_ActorVia, 0, len(projected.Actor.Vias)),
		AuthEnabled: projected.Actor.AuthEnabled, PermEnabled: projected.Actor.PermEnabled,
	}
	for _, via := range projected.Actor.Vias {
		result.Vias = append(result.Vias, actorVia(via.Name))
	}
	if value.AuthEnabled {
		result.AuthCredential = g.buildDataSchema(value.AuthCredential, projection.ProjectDataDeclaration(g.Domain, value.AuthCredential))
		result.AuthInfo = g.buildDataSchema(value.AuthInfo, projection.ProjectDataDeclaration(g.Domain, value.AuthInfo))
		result.AuthService = g.buildGeneratedServiceSchema(value.AuthService)
		result.AuthMethod = g.buildGeneratedMethodSchema(value.AuthMethod)
	}
	result.PermService = g.buildGeneratedServiceSchema(value.PermService)
	result.PermMethod = g.buildGeneratedMethodSchema(value.PermMethod)
	return result
}

func (g *_Gen) buildWebSchema(value *model.Web, projected *contractschema.Declaration) *_WebSchema {
	return &_WebSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason,
		Audiences:        g.buildActorAudienceSchemas(projected.Web.Audiences),
		MountPath:        projected.Web.MountPath,
		AuthMode:         _AuthMode(projected.Web.Auth),
	}
}

func (g *_Gen) buildGeneratedServiceSchema(value *model.Service) *_ServiceSchema {
	if value == nil {
		return nil
	}
	return g.buildServiceSchema(value, projection.ProjectServiceDeclaration(g.Domain, value))
}

func (g *_Gen) buildGeneratedMethodSchema(value *model.Method) *_MethodSchema {
	if value == nil {
		return nil
	}
	return g.buildMethodSchema(value, projection.ProjectMethodSchema(g.Domain, value))
}
