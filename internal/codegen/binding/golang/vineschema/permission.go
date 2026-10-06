package vineschema

import (
	"go.yorun.ai/skel/internal/model"
	contractschema "go.yorun.ai/skel/schema"
)

func (g *_Gen) buildPermRequireSchema(semantic *model.PermissionRequire, projected *contractschema.Requirement) *_PermRequire {
	if semantic == nil || projected == nil {
		return nil
	}
	return &_PermRequire{Expr: g.buildPermExprSchema(semantic.Expr, projected)}
}

func (g *_Gen) buildPermExprSchema(semantic *model.PermissionExpr, projected *contractschema.Requirement) *_PermExpr {
	if semantic == nil || projected == nil {
		return nil
	}
	result := &_PermExpr{Mode: permissionRequireMode(projected.Mode), Code: projected.Code}
	if projected.Check != nil && semantic.Check != nil {
		result.Check = &_PermCheckInvocation{
			ResourceSkelName: projected.Check.Resource,
			ActionName:       projected.Check.Action,
			CheckName:        projected.Check.Check,
			ServiceSkelName:  semantic.Check.ServiceSkelName,
			MethodSkelName:   semantic.Check.MethodSkelName,
			CodeArgumentName: semantic.Check.CodeArgumentName,
			Arguments:        g.buildPermCheckArgumentSchemas(projected.Check.Arguments),
		}
	}
	for index, child := range projected.Children {
		result.Children = append(result.Children, g.buildPermExprSchema(semantic.Children[index], child))
	}
	return result
}

func (g *_Gen) buildPermCheckArgumentSchemas(arguments []*contractschema.RequirementCheckArgument) []*_PermCheckArgument {
	result := make([]*_PermCheckArgument, 0, len(arguments))
	for _, argument := range arguments {
		result = append(result, &_PermCheckArgument{
			Name: argument.Name, JsonPath: argument.JSONPath, Type: g.buildTypeSchema(argument.Type),
		})
	}
	return result
}

func permissionRequireMode(mode contractschema.RequirementMode) _PermRequireMode {
	switch mode {
	case contractschema.RequirementModeCode:
		return permRequireModeCode
	case contractschema.RequirementModeCheck, contractschema.RequirementModeReference:
		return permRequireModeCheck
	case contractschema.RequirementModeAll:
		return permRequireModeAll
	case contractschema.RequirementModeAny:
		return permRequireModeAny
	default:
		return ""
	}
}
