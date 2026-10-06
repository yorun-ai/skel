package schema

import "reflect"

// Directions describe existing producer/consumer roles, not generated Go signatures.
type _Usage uint8

const (
	usageInput _Usage = 1 << iota
	usageOutput
	usageBoth = usageInput | usageOutput
)

func collectDiffUsage(documents ...*Document) map[string]_Usage {
	usage := map[string]_Usage{}
	for _, document := range documents {
		declarations := map[string]*Declaration{}
		for _, declaration := range document.Declarations {
			declarations[declaration.SkelName] = declaration
		}
		visited := map[string]_Usage{}
		var walk func(*Type, _Usage)
		walk = func(kind *Type, direction _Usage) {
			if kind == nil {
				return
			}
			walk(kind.Element, direction)
			walk(kind.Key, direction)
			walk(kind.Value, direction)
			// Generic substitutions and independently exported types may be used
			// in either role; do not infer a one-way guarantee for them.
			for _, argument := range kind.Arguments {
				walk(argument, usageBoth)
			}
			declaration := declarations[kind.Name]
			if declaration == nil || declaration.Data == nil {
				return
			}
			if declaration.Pub || len(declaration.Data.TypeParameters) != 0 {
				direction = usageBoth
			}
			usage[kind.Name] |= direction
			if visited[kind.Name]&direction == direction {
				return
			}
			visited[kind.Name] |= direction
			for _, member := range declaration.Data.Members {
				walk(member.Type, direction)
			}
		}
		for _, declaration := range document.Declarations {
			if service := declaration.Service; service != nil {
				for _, method := range service.Methods {
					for _, argument := range method.Arguments {
						walk(argument.Type, usageInput)
					}
					walk(method.Result, usageOutput)
				}
			}
			if data := declaration.Data; data != nil && (declaration.Pub || declaration.Kind != DeclarationTypeData) {
				walk(&Type{Name: declaration.SkelName}, usageBoth)
			}
			if actor := declaration.Actor; actor != nil {
				for _, data := range []*DataSchema{actor.AuthCredential, actor.AuthInfo} {
					if data != nil {
						for _, member := range data.Members {
							walk(member.Type, usageBoth)
						}
					}
				}
			}
			if resource := declaration.Resource; resource != nil {
				checks := append([]*ResourceCheck{}, resource.Checks...)
				for _, action := range resource.Actions {
					checks = append(checks, action.Checks...)
				}
				for _, check := range checks {
					for _, argument := range check.Arguments {
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

func typeChangeImpact(before, after *Type, direction _Usage) ImpactLevel {
	if before != nil && after != nil && before.Nullable != after.Nullable {
		left, right := *before, *after
		left.Nullable, right.Nullable = false, false
		if reflect.DeepEqual(left, right) && ((direction == usageInput && after.Nullable) || (direction == usageOutput && before.Nullable)) {
			return ImpactCompatible
		}
	}
	return ImpactBreaking
}
