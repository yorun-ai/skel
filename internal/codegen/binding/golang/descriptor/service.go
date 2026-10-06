package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildServiceDescriptor(value *schema.Service) *descriptor.Service {
	result := &descriptor.Service{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason,
		Pub:              value.Pub,
		Api:              value.Api,
		Ext:              value.Ext,
		AuthMode:         descriptor.AuthMode(value.NormalizedAuth()),
		Audiences:        g.buildActorAudienceDescriptors(value.Audiences),
		Require:          g.buildPermissionRequireDescriptor(value.Require),
		Methods:          make([]*descriptor.Method, 0, len(value.Methods)),
	}
	for _, method := range value.Methods {
		result.Methods = append(result.Methods, g.buildMethodDescriptor(method))
	}
	return result
}

func (g *_Gen) buildMethodDescriptor(value *schema.Method) *descriptor.Method {
	return &descriptor.Method{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Example: value.Example,
		AuthMode: descriptor.AuthMode(value.NormalizedAuth()), Require: g.buildPermissionRequireDescriptor(value.Require),
		EffectiveAuthMode: descriptor.AuthMode(value.EffectiveAuthMode), EffectiveRequire: g.buildPermissionRequireDescriptor(value.EffectiveRequire),
		InputDescription: value.InputDescription, ArgumentsSensitive: value.ArgumentsSensitive,
		OutputDescription: value.OutputDescription, OutputExample: value.OutputExample,
		ResultSensitive: value.ResultSensitive, Arguments: g.buildArgumentDescriptors(value.Arguments),
		ResultType: g.buildTypeDescriptor(value.ResultType),
	}
}

func (g *_Gen) buildActorAudienceDescriptors(values []*schema.ActorAudience) []*descriptor.ActorAudience {
	result := make([]*descriptor.ActorAudience, 0, len(values))
	for _, value := range values {
		name, skelName := localAndSkelName(g.Domain.ReferenceName(value.Actor))
		result = append(result, &descriptor.ActorAudience{Name: name, SkelName: skelName, Via: actorVia(value.Via)})
	}
	return result
}
