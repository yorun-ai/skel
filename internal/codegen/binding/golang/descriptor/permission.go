package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildPermissionRequireDescriptor(semantic *schema.PermissionRequire) *descriptor.PermissionRequire {
	if semantic == nil {
		return nil
	}
	return &descriptor.PermissionRequire{Expression: g.buildPermissionExpressionDescriptor(semantic.Expression)}
}

func (g *_Gen) buildPermissionExpressionDescriptor(semantic *schema.PermissionExpression) *descriptor.PermissionExpression {
	if semantic == nil {
		return nil
	}
	result := &descriptor.PermissionExpression{Mode: descriptor.PermissionRequireMode(semantic.Mode), Code: semantic.Code}
	if semantic.Check != nil {
		result.Check = &descriptor.PermissionCheckInvocation{
			ResourceSkelName: g.Domain.ReferenceName(semantic.Check.ResourceSkelName),
			ActionName:       semantic.Check.ActionName,
			CheckName:        semantic.Check.CheckName,
			ServiceSkelName:  semantic.Check.ServiceSkelName,
			MethodSkelName:   semantic.Check.MethodSkelName,
			CodeArgumentName: semantic.Check.CodeArgumentName,
			Arguments:        g.buildPermissionCheckArgumentDescriptors(semantic.Check.Arguments),
		}
	}
	for _, child := range semantic.Children {
		result.Children = append(result.Children, g.buildPermissionExpressionDescriptor(child))
	}
	return result
}

func (g *_Gen) buildPermissionCheckArgumentDescriptors(arguments []*schema.PermissionCheckArgument) []*descriptor.PermissionCheckArgument {
	result := make([]*descriptor.PermissionCheckArgument, 0, len(arguments))
	for _, argument := range arguments {
		result = append(result, &descriptor.PermissionCheckArgument{
			Name: argument.Name, JsonPath: argument.JsonPath, Type: g.buildTypeDescriptor(argument.Type),
		})
	}
	return result
}
