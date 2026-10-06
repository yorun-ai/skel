package vineschema

import (
	"go.yorun.ai/skel/internal/model"
	contractschema "go.yorun.ai/skel/internal/schema"
)

func (g *_Gen) buildTaskSchema(value *model.Task, projected *contractschema.Declaration) *_TaskSchema {
	result := &_TaskSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason,
		Triggers:         make([]*_TriggerSchema, 0, len(projected.Task.Triggers)),
	}
	for index, trigger := range projected.Task.Triggers {
		result.Triggers = append(result.Triggers, &_TriggerSchema{
			Name: trigger.Name, SkelName: trigger.SkelName, Hash: value.Triggers[index].Hash,
			Description: trigger.Description, Deprecated: trigger.Deprecated,
			DeprecatedReason: trigger.DeprecatedReason, InputDescription: trigger.InputDescription,
			ArgumentsSensitive: trigger.ArgumentsSensitive, Arguments: g.buildArgumentSchemas(trigger.Arguments),
		})
	}
	return result
}
