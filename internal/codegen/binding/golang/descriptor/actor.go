package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildActorDescriptor(value *schema.Actor) *descriptor.Actor {
	result := &descriptor.Actor{
		IdentifierField: value.IdentifierField, Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Vias: make([]descriptor.ActorVia, 0, len(value.Vias)),
		AuthEnabled: value.AuthEnabled, PermissionEnabled: value.PermissionEnabled,
	}
	for _, via := range value.Vias {
		result.Vias = append(result.Vias, actorVia(via.Name))
	}
	if value.AuthEnabled {
		result.AuthCredential = g.buildDataDescriptor(value.AuthCredential)
		result.AuthInfo = g.buildDataDescriptor(value.AuthInfo)
		result.AuthService = g.buildGeneratedServiceDescriptor(value.AuthService)
		result.AuthMethod = g.buildGeneratedMethodDescriptor(value.AuthMethod)
	}
	result.PermissionService = g.buildGeneratedServiceDescriptor(value.PermissionService)
	result.PermissionMethod = g.buildGeneratedMethodDescriptor(value.PermissionMethod)
	return result
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

func (g *_Gen) buildGeneratedMethodDescriptor(value *schema.Method) *descriptor.Method {
	if value == nil {
		return nil
	}
	return g.buildMethodDescriptor(value)
}
