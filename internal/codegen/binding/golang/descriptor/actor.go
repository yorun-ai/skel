package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildActorDescriptor(value *schema.Actor) *descriptor.Actor {
	result := new(descriptor.Actor{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Vias: make([]descriptor.ActorViaKind, 0, len(value.Vias)),
	})
	for _, via := range value.Vias {
		result.Vias = append(result.Vias, actorVia(via.Name))
	}
	if auth := value.Auth; auth != nil {
		result.Auth = new(descriptor.ActorAuth{
			Credential: g.buildDataDescriptor(auth.Credential), Info: g.buildDataDescriptor(auth.Info),
			IdentifierField: auth.IdentifierField,
			Service:         g.buildGeneratedServiceDescriptor(auth.Service), MethodName: generatedMethodName(auth.Method),
		})
	}
	if permission := value.Permission; permission != nil {
		result.Permission = new(descriptor.ActorPermission{
			Service: g.buildGeneratedServiceDescriptor(permission.Service), MethodName: generatedMethodName(permission.Method),
		})
	}
	return result
}

func generatedMethodName(value *schema.Method) string {
	if value == nil {
		return ""
	}
	return value.Name
}

func (g *_Gen) buildWebDescriptor(value *schema.Web) *descriptor.Web {
	return &descriptor.Web{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason,
		Audiences:        g.buildActorAudienceDescriptors(value.Audiences),
		MountPath:        value.MountPath,
		AuthMode:         descriptor.AuthMode(value.NormalizedAuth()),
	}
}

func (g *_Gen) buildGeneratedServiceDescriptor(value *schema.Service) *descriptor.Service {
	if value == nil {
		return nil
	}
	return g.buildServiceDescriptor(value)
}
