package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildResourceDescriptor(value *schema.Resource) *descriptor.Resource {
	result := &descriptor.Resource{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason,
		Checks:           g.buildResourceCheckDescriptors(value.Checks),
		Actions:          make([]*descriptor.ResourceAction, 0, len(value.Actions)),
	}
	for _, action := range value.Actions {
		result.Actions = append(result.Actions, &descriptor.ResourceAction{
			Name: action.Name, PermissionCode: action.PermissionCode,
			Description: action.Description, Deprecated: action.Deprecated,
			DeprecatedReason: action.DeprecatedReason,
			Checks:           g.buildResourceCheckDescriptors(action.Checks),
		})
	}
	result.CheckService = g.buildGeneratedServiceDescriptor(value.CheckService)
	return result
}

func (g *_Gen) buildResourceCheckDescriptors(values []*schema.ResourceCheck) []*descriptor.ResourceCheck {
	result := make([]*descriptor.ResourceCheck, 0, len(values))
	for _, value := range values {
		result = append(result, &descriptor.ResourceCheck{
			Name: value.Name, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason,
			MethodName: generatedMethodName(value.Method),
		})
	}
	return result
}
