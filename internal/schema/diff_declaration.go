package schema

import (
	"fmt"
	"reflect"
	"slices"

	"go.yorun.ai/skelc/internal/model"
)

func (c *_Diff) compareDeclaration(baseline, candidate *Declaration) {
	if baseline.Kind != candidate.Kind {
		c.add(ImpactBreaking, "declaration.type.changed", candidate.SkelName,
			fmt.Sprintf("declaration type changed from %s to %s", baseline.Kind, candidate.Kind), baseline.Pos, candidate.Pos)
		return
	}
	if baseline.Pub != candidate.Pub {
		impact := ImpactCompatible
		code := "declaration.visibility.increased"
		message := "declaration became public"
		if baseline.Pub {
			impact = ImpactBreaking
			code = "declaration.visibility.reduced"
			message = "public declaration became non-public"
		}
		c.add(impact, code, candidate.SkelName, message, baseline.Pos, candidate.Pos)
	}
	c.compareMetadata("declaration", candidate.SkelName, baseline.Metadata, candidate.Metadata, baseline.Pos, candidate.Pos)
	switch baseline.Kind {
	case DeclarationTypeEnum:
		c.compareEnum(candidate.SkelName, baseline.Enum, candidate.Enum)
	case DeclarationTypeData, DeclarationTypeConfig, DeclarationTypeEvent:
		c.compareData(candidate.SkelName, baseline.Data, candidate.Data)
	case DeclarationTypeActor:
		c.compareActor(candidate.SkelName, baseline.Actor, candidate.Actor)
	case DeclarationTypeResource:
		c.compareResource(candidate.SkelName, baseline.Resource, candidate.Resource)
	case DeclarationTypeService:
		c.compareService(candidate.SkelName, baseline.Service, candidate.Service)
	case DeclarationTypeWeb:
		c.compareAudiences(candidate.SkelName, "web.audience", baseline.Web.Audiences, candidate.Web.Audiences)
		if baseline.Web.MountPath != candidate.Web.MountPath {
			c.add(ImpactBreaking, "web.mount-path.changed", candidate.SkelName,
				fmt.Sprintf("web mount path changed from %q to %q", baseline.Web.MountPath, candidate.Web.MountPath), baseline.Pos, candidate.Pos)
		}
	case DeclarationTypeTask:
		c.compareTask(candidate.SkelName, baseline.Task, candidate.Task)
	}
}

func (c *_Diff) compareEnum(owner string, baseline, candidate *EnumSchema) {
	baselineByName := enumItemsByName(baseline.Items)
	candidateByName := enumItemsByName(candidate.Items)
	for _, item := range baseline.Items {
		other := candidateByName[item.Name]
		symbol := owner + "." + item.Name
		if other == nil {
			c.add(ImpactBreaking, "enum.item.removed", symbol, fmt.Sprintf("enum item %s was removed", item.Name), item.Pos, model.Position{})
			continue
		}
		c.compareMetadata("enum.item", symbol, item.Metadata, other.Metadata, item.Pos, other.Pos)
	}
	for _, item := range candidate.Items {
		if baselineByName[item.Name] == nil {
			c.add(ImpactDangerous, "enum.item.added", owner+"."+item.Name,
				fmt.Sprintf("enum item %s was added", item.Name), model.Position{}, item.Pos)
		}
	}
}

func (c *_Diff) compareData(owner string, baseline, candidate *DataSchema) {
	if baseline.Lifecycle != candidate.Lifecycle {
		c.add(ImpactDangerous, "config.lifecycle.changed", owner,
			fmt.Sprintf("config lifecycle changed from %s to %s", baseline.Lifecycle, candidate.Lifecycle), model.Position{}, model.Position{})
	}
	if baseline.Sensitive != candidate.Sensitive {
		c.add(ImpactDangerous, "data.sensitive.changed", owner, "data sensitivity changed", model.Position{}, model.Position{})
	}
	if !slices.Equal(baseline.TypeParameters, candidate.TypeParameters) {
		c.add(ImpactBreaking, "data.type-parameters.changed", owner, "data type parameters changed", model.Position{}, model.Position{})
	}
	c.compareMembers(owner, "data.member", baseline.Members, candidate.Members, ImpactDangerous)
}

func (c *_Diff) compareMembers(owner, prefix string, baseline, candidate []*Member, reorderImpact ImpactLevel) {
	baselineByName := membersByName(baseline)
	candidateByName := membersByName(candidate)
	for _, member := range baseline {
		other := candidateByName[member.Name]
		symbol := owner + "." + member.Name
		if other == nil {
			c.add(ImpactBreaking, prefix+".removed", symbol, fmt.Sprintf("member %s was removed", member.Name), member.Pos, model.Position{})
			continue
		}
		if !reflect.DeepEqual(member.Type, other.Type) {
			c.add(ImpactBreaking, prefix+".type.changed", symbol,
				fmt.Sprintf("member type changed from %s to %s", typeDisplay(member.Type), typeDisplay(other.Type)), member.Pos, other.Pos)
		}
		if member.Sensitive != other.Sensitive {
			c.add(ImpactDangerous, prefix+".sensitive.changed", symbol, "member sensitivity changed", member.Pos, other.Pos)
		}
		if member.Example != other.Example {
			c.add(ImpactCompatible, prefix+".example.changed", symbol, "member example changed", member.Pos, other.Pos)
		}
		c.compareMetadata(prefix, symbol, member.Metadata, other.Metadata, member.Pos, other.Pos)
	}
	for _, member := range candidate {
		if baselineByName[member.Name] == nil {
			c.add(ImpactBreaking, prefix+".added", owner+"."+member.Name,
				fmt.Sprintf("required member %s was added", member.Name), model.Position{}, member.Pos)
		}
	}
	if sameNamedSet(memberNames(baseline), memberNames(candidate)) && !slices.Equal(memberNames(baseline), memberNames(candidate)) {
		c.add(reorderImpact, prefix+".order.changed", owner, "member order changed", model.Position{}, model.Position{})
	}
}

func (c *_Diff) compareActor(owner string, baseline, candidate *ActorSchema) {
	if baseline.IdentifierField != candidate.IdentifierField {
		c.add(ImpactDangerous, "actor.identifier.changed", owner, "actor identifier field changed", model.Position{}, model.Position{})
	}
	c.compareStringSet(owner, "actor.via", actorViaNames(baseline.Vias), actorViaNames(candidate.Vias), ImpactBreaking, ImpactCompatible)
	if baseline.AuthEnabled != candidate.AuthEnabled {
		code := "actor.auth.added"
		message := "actor authentication was added"
		if baseline.AuthEnabled {
			code, message = "actor.auth.removed", "actor authentication was removed"
		}
		c.add(ImpactDangerous, code, owner, message, model.Position{}, model.Position{})
	}
	if baseline.AuthEnabled && candidate.AuthEnabled {
		c.compareMembers(owner+".credential", "actor.auth-credential.member", baseline.AuthCredential.Members, candidate.AuthCredential.Members, ImpactDangerous)
		c.compareMembers(owner+".info", "actor.auth-info.member", baseline.AuthInfo.Members, candidate.AuthInfo.Members, ImpactDangerous)
	}
	if baseline.PermEnabled != candidate.PermEnabled {
		code := "actor.permission.added"
		message := "actor permission support was added"
		if baseline.PermEnabled {
			code, message = "actor.permission.removed", "actor permission support was removed"
		}
		c.add(ImpactDangerous, code, owner, message, model.Position{}, model.Position{})
	}
}

func (c *_Diff) compareMetadata(prefix, symbol string, baseline, candidate Metadata, baselinePos, candidatePos model.Position) {
	if baseline.Description != candidate.Description {
		c.add(ImpactCompatible, prefix+".description.changed", symbol, "description changed", baselinePos, candidatePos)
	}
	if baseline.Deprecated != candidate.Deprecated {
		message := "deprecation was removed"
		if candidate.Deprecated {
			message = "deprecation was added"
		}
		c.add(ImpactCompatible, prefix+".deprecated.changed", symbol, message, baselinePos, candidatePos)
	}
	if baseline.DeprecatedReason != candidate.DeprecatedReason {
		c.add(ImpactCompatible, prefix+".deprecated-reason.changed", symbol, "deprecation reason changed", baselinePos, candidatePos)
	}
}

func (c *_Diff) compareStringSet(owner, prefix string, baseline, candidate []string, removedImpact, addedImpact ImpactLevel) {
	baselineSet := stringSet(baseline)
	candidateSet := stringSet(candidate)
	for _, value := range baseline {
		if !candidateSet[value] {
			c.add(removedImpact, prefix+".removed", owner, fmt.Sprintf("%s %s was removed", prefix, value), model.Position{}, model.Position{})
		}
	}
	for _, value := range candidate {
		if !baselineSet[value] {
			c.add(addedImpact, prefix+".added", owner, fmt.Sprintf("%s %s was added", prefix, value), model.Position{}, model.Position{})
		}
	}
}

func (c *_Diff) compareResource(owner string, baseline, candidate *ResourceSchema) {
	c.compareResourceChecks(owner, baseline.Checks, candidate.Checks)
	baselineByName := resourceActionsByName(baseline.Actions)
	candidateByName := resourceActionsByName(candidate.Actions)
	for _, action := range baseline.Actions {
		other := candidateByName[action.Name]
		symbol := owner + "." + action.Name
		if other == nil {
			c.add(ImpactBreaking, "resource.action.removed", symbol, fmt.Sprintf("resource action %s was removed", action.Name), action.Pos, model.Position{})
			continue
		}
		if action.PermissionCode != other.PermissionCode {
			c.add(ImpactDangerous, "resource.action.code.changed", symbol, "resource action permission code changed", action.Pos, other.Pos)
		}
		c.compareMetadata("resource.action", symbol, action.Metadata, other.Metadata, action.Pos, other.Pos)
		c.compareResourceChecks(symbol, action.Checks, other.Checks)
	}
	for _, action := range candidate.Actions {
		if baselineByName[action.Name] == nil {
			c.add(ImpactCompatible, "resource.action.added", owner+"."+action.Name,
				fmt.Sprintf("resource action %s was added", action.Name), model.Position{}, action.Pos)
		}
	}
}

func (c *_Diff) compareResourceChecks(owner string, baseline, candidate []*ResourceCheck) {
	baselineByName := resourceChecksByName(baseline)
	candidateByName := resourceChecksByName(candidate)
	for _, check := range baseline {
		other := candidateByName[check.Name]
		symbol := owner + "." + check.Name
		if other == nil {
			c.add(ImpactBreaking, "resource.check.removed", symbol, fmt.Sprintf("resource check %s was removed", check.Name), check.Pos, model.Position{})
			continue
		}
		c.compareArguments(symbol, "resource.check.argument", check.Arguments, other.Arguments)
		c.compareMetadata("resource.check", symbol, check.Metadata, other.Metadata, check.Pos, other.Pos)
	}
	for _, check := range candidate {
		if baselineByName[check.Name] == nil {
			c.add(ImpactCompatible, "resource.check.added", owner+"."+check.Name,
				fmt.Sprintf("resource check %s was added", check.Name), model.Position{}, check.Pos)
		}
	}
}

func (c *_Diff) compareTask(owner string, baseline, candidate *TaskSchema) {
	baselineByName := triggersByName(baseline.Triggers)
	candidateByName := triggersByName(candidate.Triggers)
	for _, trigger := range baseline.Triggers {
		other := candidateByName[trigger.Name]
		symbol := owner + "." + trigger.Name
		if other == nil {
			c.add(ImpactBreaking, "task.trigger.removed", symbol, fmt.Sprintf("task trigger %s was removed", trigger.Name), trigger.Pos, model.Position{})
			continue
		}
		c.compareArguments(symbol, "task.trigger.argument", trigger.Arguments, other.Arguments)
		if trigger.ArgumentsSensitive != other.ArgumentsSensitive {
			c.add(ImpactDangerous, "task.trigger.sensitive.changed", symbol, "task trigger sensitivity changed", trigger.Pos, other.Pos)
		}
		if trigger.InputDescription != other.InputDescription {
			c.add(ImpactCompatible, "task.trigger.documentation.changed", symbol, "task trigger documentation changed", trigger.Pos, other.Pos)
		}
		c.compareMetadata("task.trigger", symbol, trigger.Metadata, other.Metadata, trigger.Pos, other.Pos)
	}
	for _, trigger := range candidate.Triggers {
		if baselineByName[trigger.Name] == nil {
			c.add(ImpactCompatible, "task.trigger.added", owner+"."+trigger.Name,
				fmt.Sprintf("task trigger %s was added", trigger.Name), model.Position{}, trigger.Pos)
		}
	}
}
