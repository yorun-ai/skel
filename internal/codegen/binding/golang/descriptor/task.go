package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildTaskDescriptor(value *schema.Task) *descriptor.Task {
	result := &descriptor.Task{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason,
		Triggers:         make([]*descriptor.Trigger, 0, len(value.Triggers)),
	}
	for _, trigger := range value.Triggers {
		result.Triggers = append(result.Triggers, &descriptor.Trigger{
			Name: trigger.Name, SkelName: trigger.SkelName, Hash: trigger.Hash,
			Description: trigger.Description, Deprecated: trigger.Deprecated,
			DeprecatedReason: trigger.DeprecatedReason, InputDescription: trigger.InputDescription,
			ArgumentsSensitive: trigger.ArgumentsSensitive, Arguments: g.buildArgumentDescriptors(trigger.Arguments),
		})
	}
	return result
}
