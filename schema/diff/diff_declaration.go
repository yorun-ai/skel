package diff

import (
	"fmt"
	"slices"

	"go.yorun.ai/skel/schema"
)

func (c *_Diff) compareDeclaration(baseline, candidate *schema.Declaration) {
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
	c.compareMetadata("declaration", candidate.SkelName, metadata(baseline.Description, baseline.Deprecated, baseline.DeprecatedReason), metadata(candidate.Description, candidate.Deprecated, candidate.DeprecatedReason), baseline.Pos, candidate.Pos)
	switch baseline.Kind {
	case schema.DeclarationTypeEnum:
		c.compareEnum(candidate.SkelName, baseline.Enum, candidate.Enum)
	case schema.DeclarationTypeData, schema.DeclarationTypeConfig, schema.DeclarationTypeEvent:
		c.compareData(candidate.SkelName, baseline.Data, candidate.Data)
	case schema.DeclarationTypeActor:
		c.compareActor(candidate.SkelName, baseline.Actor, candidate.Actor)
	case schema.DeclarationTypeResource:
		c.compareResource(candidate.SkelName, baseline.Resource, candidate.Resource)
	case schema.DeclarationTypeService:
		c.compareService(candidate.SkelName, baseline.Service, candidate.Service)
	case schema.DeclarationTypeWeb:
		c.compareAuth(candidate.SkelName, "web", baseline.Web.NormalizedAuth(), candidate.Web.NormalizedAuth(), baseline.Pos, candidate.Pos)
		c.compareAudiences(candidate.SkelName, "web.audience", baseline.Web.Audiences, candidate.Web.Audiences)
		if baseline.Web.MountPath != candidate.Web.MountPath {
			c.add(ImpactBreaking, "web.mount-path.changed", candidate.SkelName,
				fmt.Sprintf("web mount path changed from %q to %q", baseline.Web.MountPath, candidate.Web.MountPath), baseline.Pos, candidate.Pos)
		}
	case schema.DeclarationTypeTask:
		c.compareTask(candidate.SkelName, baseline.Task, candidate.Task)
	}
}

func (c *_Diff) compareEnum(owner string, baseline, candidate *schema.Enum) {
	baselineByName := enumItemsByName(baseline.Items)
	candidateByName := enumItemsByName(candidate.Items)
	for _, item := range baseline.Items {
		other := candidateByName[item.Name]
		symbol := owner + "." + item.Name
		if other == nil {
			c.add(ImpactBreaking, "enum.item.removed", symbol, fmt.Sprintf("enum item %s was removed", item.Name), item.Pos, schema.Position{})
			continue
		}
		c.compareMetadata("enum.item", symbol, metadata(item.Description, item.Deprecated, item.DeprecatedReason), metadata(other.Description, other.Deprecated, other.DeprecatedReason), item.Pos, other.Pos)
	}
	for _, item := range candidate.Items {
		if baselineByName[item.Name] == nil {
			c.add(ImpactDangerous, "enum.item.added", owner+"."+item.Name,
				fmt.Sprintf("enum item %s was added", item.Name), schema.Position{}, item.Pos)
		}
	}
}

func (c *_Diff) compareData(owner string, baseline, candidate *schema.Data) {
	if baseline.Ext != candidate.Ext {
		c.add(ImpactBreaking, "event.ext.changed", owner, "event extension direction changed", schema.Position{}, schema.Position{})
	}
	if baseline.Lifecycle != candidate.Lifecycle {
		c.add(ImpactDangerous, "config.lifecycle.changed", owner,
			fmt.Sprintf("config lifecycle changed from %s to %s", baseline.Lifecycle, candidate.Lifecycle), schema.Position{}, schema.Position{})
	}
	if baseline.Sensitive != candidate.Sensitive {
		c.add(ImpactDangerous, "data.sensitive.changed", owner, "data sensitivity changed", schema.Position{}, schema.Position{})
	}
	if !slices.EqualFunc(baseline.TypeParameters, candidate.TypeParameters, func(a, b *schema.TypeParameter) bool { return a.Name == b.Name }) {
		c.add(ImpactBreaking, "data.type-parameters.changed", owner, "data type parameters changed", schema.Position{}, schema.Position{})
	}
	c.compareMembers(owner, "data.member", baseline.Members, candidate.Members, ImpactCompatible)
}

func (c *_Diff) compareMembers(owner, prefix string, baseline, candidate []*schema.DataMember, reorderImpact ImpactLevel) {
	baselineByName := membersByName(baseline)
	candidateByName := membersByName(candidate)
	for _, member := range baseline {
		other := candidateByName[member.Name]
		symbol := owner + "." + member.Name
		if other == nil {
			c.add(ImpactBreaking, prefix+".removed", symbol, fmt.Sprintf("member %s was removed", member.Name), member.Pos, schema.Position{})
			continue
		}
		if !c.sameType(member.Type, other.Type) {
			direction := c.usage[owner]
			if prefix == "actor.auth-credential.member" {
				direction = usageInput
			}
			if prefix == "actor.auth-info.member" {
				direction = usageOutput
			}
			c.add(c.typeChangeImpact(member.Type, other.Type, direction), prefix+".type.changed", symbol,
				fmt.Sprintf("member type changed from %s to %s", typeDisplay(c.baseline, member.Type), typeDisplay(c.candidate, other.Type)), member.Pos, other.Pos)
		}
		if member.Sensitive != other.Sensitive {
			c.add(ImpactDangerous, prefix+".sensitive.changed", symbol, "member sensitivity changed", member.Pos, other.Pos)
		}
		if member.Example != other.Example {
			c.add(ImpactCompatible, prefix+".example.changed", symbol, "member example changed", member.Pos, other.Pos)
		}
		c.compareMetadata(prefix, symbol, metadata(member.Description, member.Deprecated, member.DeprecatedReason), metadata(other.Description, other.Deprecated, other.DeprecatedReason), member.Pos, other.Pos)
	}
	for _, member := range candidate {
		if baselineByName[member.Name] == nil {
			impact := ImpactBreaking
			if (prefix == "actor.auth-credential.member" && member.Type.Nullable) || c.usage[owner] == usageOutput {
				impact = ImpactCompatible
			}
			c.add(impact, prefix+".added", owner+"."+member.Name,
				fmt.Sprintf("member %s was added", member.Name), schema.Position{}, member.Pos)
		}
	}
	if sameNamedSet(memberNames(baseline), memberNames(candidate)) && !slices.Equal(memberNames(baseline), memberNames(candidate)) {
		c.add(reorderImpact, prefix+".order.changed", owner, "member order changed", schema.Position{}, schema.Position{})
	}
}

func (c *_Diff) compareActor(owner string, baseline, candidate *schema.Actor) {
	var baselineIdentifier, candidateIdentifier string
	if baseline.Auth != nil {
		baselineIdentifier = baseline.Auth.IdentifierField
	}
	if candidate.Auth != nil {
		candidateIdentifier = candidate.Auth.IdentifierField
	}
	if baselineIdentifier != candidateIdentifier {
		c.add(ImpactDangerous, "actor.identifier.changed", owner, "actor identifier field changed", actorIdentifierPosition(baseline), actorIdentifierPosition(candidate))
	}
	c.compareStringSet(owner, "actor.via", actorViaNames(baseline.Vias), actorViaNames(candidate.Vias), ImpactBreaking, ImpactCompatible)
	if (baseline.Auth != nil) != (candidate.Auth != nil) {
		code := "actor.auth.added"
		message := "actor authentication was added"
		if baseline.Auth != nil {
			code, message = "actor.auth.removed", "actor authentication was removed"
		}
		impact := ImpactDangerous
		if baseline.Auth != nil {
			impact = ImpactBreaking
		}
		c.add(impact, code, owner, message, actorAuthPosition(baseline), actorAuthPosition(candidate))
	}
	if baseline.Auth != nil && candidate.Auth != nil {
		c.compareActorAuthData(owner+".credential", "actor.auth-credential", baseline.Auth.Credential, candidate.Auth.Credential, actorAuthPosition(baseline), actorAuthPosition(candidate))
		c.compareActorAuthData(owner+".info", "actor.auth-info", baseline.Auth.Info, candidate.Auth.Info, actorAuthPosition(baseline), actorAuthPosition(candidate))
	}
	if (baseline.Permission != nil) != (candidate.Permission != nil) {
		code := "actor.permission.added"
		message := "actor permission support was added"
		if baseline.Permission != nil {
			code, message = "actor.permission.removed", "actor permission support was removed"
		}
		impact := ImpactDangerous
		if baseline.Permission != nil {
			impact = ImpactBreaking
		}
		c.add(impact, code, owner, message, actorPermissionPosition(baseline), actorPermissionPosition(candidate))
	}
}

func (c *_Diff) compareActorAuthData(owner, prefix string, baseline, candidate *schema.Data, baselinePos, candidatePos schema.Position) {
	if baseline.Sensitive != candidate.Sensitive {
		if baseline.Pos != (schema.Position{}) {
			baselinePos = baseline.Pos
		}
		if candidate.Pos != (schema.Position{}) {
			candidatePos = candidate.Pos
		}
		c.add(ImpactDangerous, prefix+".sensitive.changed", owner, "data sensitivity changed", baselinePos, candidatePos)
	}
	c.compareMembers(owner, prefix+".member", baseline.Members, candidate.Members, ImpactCompatible)
}

func actorAuthPosition(actor *schema.Actor) schema.Position {
	if actor.Auth != nil && actor.Auth.Pos != (schema.Position{}) {
		return actor.Auth.Pos
	}
	return actor.Pos
}

func actorPermissionPosition(actor *schema.Actor) schema.Position {
	if actor.Permission != nil && actor.Permission.Pos != (schema.Position{}) {
		return actor.Permission.Pos
	}
	return actor.Pos
}

func actorIdentifierPosition(actor *schema.Actor) schema.Position {
	if auth := actor.Auth; auth != nil && auth.Info != nil {
		for _, member := range auth.Info.Members {
			if member.Name == auth.IdentifierField && member.Pos != (schema.Position{}) {
				return member.Pos
			}
		}
	}
	return actorAuthPosition(actor)
}

func (c *_Diff) compareMetadata(prefix, symbol string, baseline, candidate _Metadata, baselinePos, candidatePos schema.Position) {
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
			c.add(removedImpact, prefix+".removed", owner, fmt.Sprintf("%s %s was removed", prefix, value), schema.Position{}, schema.Position{})
		}
	}
	for _, value := range candidate {
		if !baselineSet[value] {
			c.add(addedImpact, prefix+".added", owner, fmt.Sprintf("%s %s was added", prefix, value), schema.Position{}, schema.Position{})
		}
	}
}

func (c *_Diff) compareResource(owner string, baseline, candidate *schema.Resource) {
	c.compareResourceChecks(owner, baseline.Checks, candidate.Checks)
	baselineByName := resourceActionsByName(baseline.Actions)
	candidateByName := resourceActionsByName(candidate.Actions)
	for _, action := range baseline.Actions {
		other := candidateByName[action.Name]
		symbol := owner + "." + action.Name
		if other == nil {
			c.add(ImpactBreaking, "resource.action.removed", symbol, fmt.Sprintf("resource action %s was removed", action.Name), action.Pos, schema.Position{})
			continue
		}
		if action.PermissionCode != other.PermissionCode {
			c.add(ImpactDangerous, "resource.action.code.changed", symbol, "resource action permission code changed", action.Pos, other.Pos)
		}
		c.compareMetadata("resource.action", symbol, metadata(action.Description, action.Deprecated, action.DeprecatedReason), metadata(other.Description, other.Deprecated, other.DeprecatedReason), action.Pos, other.Pos)
		c.compareResourceChecks(symbol, action.Checks, other.Checks)
	}
	for _, action := range candidate.Actions {
		if baselineByName[action.Name] == nil {
			c.add(ImpactCompatible, "resource.action.added", owner+"."+action.Name,
				fmt.Sprintf("resource action %s was added", action.Name), schema.Position{}, action.Pos)
		}
	}
}

func (c *_Diff) compareResourceChecks(owner string, baseline, candidate []*schema.ResourceCheck) {
	baselineByName := resourceChecksByName(baseline)
	candidateByName := resourceChecksByName(candidate)
	for _, check := range baseline {
		other := candidateByName[check.Name]
		symbol := owner + "." + check.Name
		if other == nil {
			c.add(ImpactBreaking, "resource.check.removed", symbol, fmt.Sprintf("resource check %s was removed", check.Name), check.Method.Pos, schema.Position{})
			continue
		}
		c.compareArguments(symbol, "resource.check.argument", check.Method.Arguments, other.Method.Arguments)
		c.compareMetadata("resource.check", symbol, metadata(check.Method.Description, check.Deprecated, check.DeprecatedReason), metadata(other.Method.Description, other.Deprecated, other.DeprecatedReason), check.Method.Pos, other.Method.Pos)
	}
	for _, check := range candidate {
		if baselineByName[check.Name] == nil {
			c.add(ImpactCompatible, "resource.check.added", owner+"."+check.Name,
				fmt.Sprintf("resource check %s was added", check.Name), schema.Position{}, check.Method.Pos)
		}
	}
}

func (c *_Diff) compareTask(owner string, baseline, candidate *schema.Task) {
	baselineByName := triggersByName(baseline.Triggers)
	candidateByName := triggersByName(candidate.Triggers)
	for _, trigger := range baseline.Triggers {
		other := candidateByName[trigger.Name]
		symbol := owner + "." + trigger.Name
		if other == nil {
			c.add(ImpactBreaking, "task.trigger.removed", symbol, fmt.Sprintf("task trigger %s was removed", trigger.Name), trigger.Pos, schema.Position{})
			continue
		}
		c.compareArguments(symbol, "task.trigger.argument", trigger.Arguments, other.Arguments)
		if trigger.ArgumentsSensitive != other.ArgumentsSensitive {
			c.add(ImpactDangerous, "task.trigger.sensitive.changed", symbol, "task trigger sensitivity changed", trigger.Pos, other.Pos)
		}
		if trigger.InputDescription != other.InputDescription {
			c.add(ImpactCompatible, "task.trigger.documentation.changed", symbol, "task trigger documentation changed", trigger.Pos, other.Pos)
		}
		c.compareMetadata("task.trigger", symbol, metadata(trigger.Description, trigger.Deprecated, trigger.DeprecatedReason), metadata(other.Description, other.Deprecated, other.DeprecatedReason), trigger.Pos, other.Pos)
	}
	for _, trigger := range candidate.Triggers {
		if baselineByName[trigger.Name] == nil {
			c.add(ImpactCompatible, "task.trigger.added", owner+"."+trigger.Name,
				fmt.Sprintf("task trigger %s was added", trigger.Name), schema.Position{}, trigger.Pos)
		}
	}
}
