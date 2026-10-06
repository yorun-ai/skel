package vineschema

import (
	"go.yorun.ai/skel/internal/model"
	contractschema "go.yorun.ai/skel/internal/schema"
)

func (g *_Gen) buildResourceSchema(value *model.Resource, projected *contractschema.Declaration) *_ResourceSchema {
	result := &_ResourceSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason,
		Checks:           g.buildResourceCheckSchemas(value.Checks, projected.Resource.Checks),
		Actions:          make([]*_ResourceActionSchema, 0, len(projected.Resource.Actions)),
	}
	for index, action := range projected.Resource.Actions {
		result.Actions = append(result.Actions, &_ResourceActionSchema{
			Name: action.Name, PermissionCode: action.PermissionCode,
			Description: action.Description, Deprecated: action.Deprecated,
			DeprecatedReason: action.DeprecatedReason,
			Checks:           g.buildResourceCheckSchemas(value.Actions[index].Checks, action.Checks),
		})
	}
	result.CheckService = g.buildGeneratedServiceSchema(value.CheckService)
	return result
}

func (g *_Gen) buildResourceCheckSchemas(values []*model.ResourceCheck, projected []*contractschema.ResourceCheck) []*_ResourceCheckSchema {
	result := make([]*_ResourceCheckSchema, 0, len(projected))
	for index, check := range projected {
		value := values[index]
		result = append(result, &_ResourceCheckSchema{
			Name: check.Name, Deprecated: check.Deprecated, DeprecatedReason: check.DeprecatedReason,
			Method: g.buildGeneratedMethodSchema(value.Method), Arguments: g.buildArgumentSchemas(check.Arguments),
		})
	}
	return result
}
