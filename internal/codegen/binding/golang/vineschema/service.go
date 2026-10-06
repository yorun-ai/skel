package vineschema

import (
	"go.yorun.ai/skel/internal/model"
	contractschema "go.yorun.ai/skel/internal/schema"
)

func (g *_Gen) buildServiceSchema(value *model.Service, projected *contractschema.Declaration) *_ServiceSchema {
	result := &_ServiceSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason,
		Pub:              projected.Pub,
		Api:              projected.Service.Api,
		Ext:              projected.Service.Ext,
		AuthMode:         _AuthMode(projected.Service.Auth),
		Audiences:        g.buildActorAudienceSchemas(projected.Service.Audiences),
		Require:          g.buildPermRequireSchema(value.Require, projected.Service.Require),
		Methods:          make([]*_MethodSchema, 0, len(projected.Service.Methods)),
	}
	for index, method := range projected.Service.Methods {
		result.Methods = append(result.Methods, g.buildMethodSchema(value.Methods[index], method))
	}
	return result
}

func (g *_Gen) buildMethodSchema(value *model.Method, projected *contractschema.Method) *_MethodSchema {
	return &_MethodSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason, Example: projected.Example,
		AuthMode: _AuthMode(projected.Auth), Require: g.buildPermRequireSchema(value.Require, projected.Require),
		InputDescription: projected.InputDescription, ArgumentsSensitive: projected.ArgumentsSensitive,
		OutputDescription: projected.OutputDescription, OutputExample: projected.OutputExample,
		ResultSensitive: projected.ResultSensitive, Arguments: g.buildArgumentSchemas(projected.Arguments),
		ResultType: g.buildTypeSchema(projected.Result),
	}
}

func (g *_Gen) buildActorAudienceSchemas(values []*contractschema.Audience) []*_ActorAudienceSchema {
	result := make([]*_ActorAudienceSchema, 0, len(values))
	for _, value := range values {
		name, skelName := localAndSkelName(value.Actor)
		result = append(result, &_ActorAudienceSchema{Name: name, SkelName: skelName, Via: actorVia(value.Via)})
	}
	return result
}
