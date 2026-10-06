package diff

import "go.yorun.ai/skel/schema"

// Directions describe existing producer/consumer roles, not generated Go signatures.
type _Usage uint8

const (
	usageInput _Usage = 1 << iota
	usageOutput
	usageBoth = usageInput | usageOutput
)

func collectDiffUsage(documents ...*schema.Domain) map[string]_Usage {
	usage := map[string]_Usage{}
	for _, document := range documents {
		declarations := map[string]*schema.Declaration{}
		for _, declaration := range document.Declarations() {
			declarations[declaration.SkelName] = declaration
		}
		visited := map[string]_Usage{}
		var walk func(*schema.Type, _Usage)
		walk = func(kind *schema.Type, direction _Usage) {
			if kind == nil {
				return
			}
			if kind.List != nil {
				walk(kind.List.Value, direction)
			}
			if kind.Map != nil {
				walk(kind.Map.Key, direction)
				walk(kind.Map.Value, direction)
			}
			// Generic substitutions and independently exported types may be used
			// in either role; do not infer a one-way guarantee for them.
			for _, argument := range kind.TypeArguments {
				walk(argument, usageBoth)
			}
			declaration := declarations[document.TypeReferenceName(kind)]
			if declaration == nil || declaration.Data == nil {
				return
			}
			if declaration.Pub || len(declaration.Data.TypeParameters) != 0 {
				direction = usageBoth
			}
			usage[document.TypeReferenceName(kind)] |= direction
			if visited[document.TypeReferenceName(kind)]&direction == direction {
				return
			}
			visited[document.TypeReferenceName(kind)] |= direction
			for _, member := range declaration.Data.Members {
				walk(member.Type, direction)
			}
		}
		for _, declaration := range document.Declarations() {
			if service := declaration.Service; service != nil {
				for _, method := range service.Methods {
					for _, argument := range method.Arguments {
						walk(argument.Type, usageInput)
					}
					walk(method.ResultType, usageOutput)
				}
			}
			if data := declaration.Data; data != nil && (declaration.Pub || declaration.Kind != schema.DeclarationTypeData) {
				walk(&schema.Type{Kind: schema.TypeKindData, SkelName: declaration.SkelName, Data: declaration.Data}, usageBoth)
			}
			if actor := declaration.Actor; actor != nil && actor.Auth != nil {
				for _, data := range []*schema.Data{actor.Auth.Credential, actor.Auth.Info} {
					if data != nil {
						for _, member := range data.Members {
							walk(member.Type, usageBoth)
						}
					}
				}
			}
			if resource := declaration.Resource; resource != nil {
				checks := append([]*schema.ResourceCheck{}, resource.Checks...)
				for _, action := range resource.Actions {
					checks = append(checks, action.Checks...)
				}
				for _, check := range checks {
					for _, argument := range check.Method.Arguments {
						walk(argument.Type, usageInput)
					}
				}
			}
			if task := declaration.Task; task != nil {
				for _, trigger := range task.Triggers {
					for _, argument := range trigger.Arguments {
						walk(argument.Type, usageInput)
					}
				}
			}
		}
	}
	return usage
}

func (c *_Diff) typeChangeImpact(before, after *schema.Type, direction _Usage) ImpactLevel {
	if before != nil && after != nil && before.Nullable != after.Nullable {
		left, right := *before, *after
		left.Nullable, right.Nullable = false, false
		if c.sameType(&left, &right) && ((direction == usageInput && after.Nullable) || (direction == usageOutput && before.Nullable)) {
			return ImpactCompatible
		}
	}
	return ImpactBreaking
}
