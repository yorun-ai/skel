package schema

import (
	"fmt"
	"reflect"
	"slices"

	"go.yorun.ai/skelc/internal/model"
)

func (c *_Diff) compareService(owner string, baseline, candidate *ServiceSchema) {
	if baseline.Ext != candidate.Ext {
		c.add(ImpactBreaking, "service.ext.changed", owner, "service contract direction changed", model.Position{}, model.Position{})
	}
	if baseline.Api != candidate.Api {
		c.add(ImpactBreaking, "service.api.changed", owner, "service invocation boundary changed", model.Position{}, model.Position{})
	}
	c.compareAudiences(owner, "service.audience", baseline.Audiences, candidate.Audiences)
	c.compareAuth(owner, "service", baseline.Auth, candidate.Auth, model.Position{}, model.Position{})
	c.compareRequirement(owner, "service", baseline.Require, candidate.Require, model.Position{}, model.Position{})
	baselineByName := methodsByName(baseline.Methods)
	candidateByName := methodsByName(candidate.Methods)
	for _, method := range baseline.Methods {
		other := candidateByName[method.Name]
		symbol := owner + "." + method.Name
		if other == nil {
			c.add(ImpactBreaking, "service.method.removed", symbol, fmt.Sprintf("service method %s was removed", method.Name), method.Pos, model.Position{})
			continue
		}
		c.compareMethod(symbol, method, other)
	}
	for _, method := range candidate.Methods {
		if baselineByName[method.Name] == nil {
			c.add(ImpactCompatible, "service.method.added", owner+"."+method.Name,
				fmt.Sprintf("service method %s was added", method.Name), model.Position{}, method.Pos)
		}
	}
}

func (c *_Diff) compareMethod(owner string, baseline, candidate *Method) {
	c.compareAuth(owner, "method", baseline.Auth, candidate.Auth, baseline.Pos, candidate.Pos)
	c.compareRequirement(owner, "method", baseline.Require, candidate.Require, baseline.Pos, candidate.Pos)
	c.compareArguments(owner, "method.argument", baseline.Arguments, candidate.Arguments)
	if !reflect.DeepEqual(baseline.Result, candidate.Result) {
		c.add(ImpactBreaking, "method.result.changed", owner,
			fmt.Sprintf("method result changed from %s to %s", typeDisplay(baseline.Result), typeDisplay(candidate.Result)), baseline.Pos, candidate.Pos)
	}
	if baseline.ArgumentsSensitive != candidate.ArgumentsSensitive || baseline.ResultSensitive != candidate.ResultSensitive {
		c.add(ImpactDangerous, "method.sensitive.changed", owner, "method sensitivity changed", baseline.Pos, candidate.Pos)
	}
	if baseline.Example != candidate.Example || baseline.InputDescription != candidate.InputDescription ||
		baseline.OutputDescription != candidate.OutputDescription || baseline.OutputExample != candidate.OutputExample {
		c.add(ImpactCompatible, "method.documentation.changed", owner, "method documentation or examples changed", baseline.Pos, candidate.Pos)
	}
	c.compareMetadata("method", owner, baseline.Metadata, candidate.Metadata, baseline.Pos, candidate.Pos)
}

func (c *_Diff) compareArguments(owner, prefix string, baseline, candidate []*Argument) {
	baselineByName := argumentsByName(baseline)
	candidateByName := argumentsByName(candidate)
	for _, argument := range baseline {
		other := candidateByName[argument.Name]
		symbol := owner + "." + argument.Name
		if other == nil {
			c.add(ImpactBreaking, prefix+".removed", symbol, fmt.Sprintf("argument %s was removed", argument.Name), argument.Pos, model.Position{})
			continue
		}
		if !reflect.DeepEqual(argument.Type, other.Type) {
			c.add(ImpactBreaking, prefix+".type.changed", symbol,
				fmt.Sprintf("argument type changed from %s to %s", typeDisplay(argument.Type), typeDisplay(other.Type)), argument.Pos, other.Pos)
		}
		if argument.Sensitive != other.Sensitive {
			c.add(ImpactDangerous, prefix+".sensitive.changed", symbol, "argument sensitivity changed", argument.Pos, other.Pos)
		}
		if argument.Example != other.Example {
			c.add(ImpactCompatible, prefix+".example.changed", symbol, "argument example changed", argument.Pos, other.Pos)
		}
		c.compareMetadata(prefix, symbol, argument.Metadata, other.Metadata, argument.Pos, other.Pos)
	}
	for _, argument := range candidate {
		if baselineByName[argument.Name] == nil {
			c.add(ImpactBreaking, prefix+".added", owner+"."+argument.Name,
				fmt.Sprintf("required argument %s was added", argument.Name), model.Position{}, argument.Pos)
		}
	}
	if sameNamedSet(argumentNames(baseline), argumentNames(candidate)) && !slices.Equal(argumentNames(baseline), argumentNames(candidate)) {
		c.add(ImpactBreaking, prefix+".order.changed", owner, "argument order changed", model.Position{}, model.Position{})
	}
}

func (c *_Diff) compareAudiences(owner, prefix string, baseline, candidate []*Audience) {
	baselineByKey := audiencesByKey(baseline)
	candidateByKey := audiencesByKey(candidate)
	for key, audience := range baselineByKey {
		if candidateByKey[key] == nil {
			c.add(ImpactBreaking, prefix+".removed", owner,
				fmt.Sprintf("audience %s via %s was removed", audience.Actor, audience.Via), audience.Pos, model.Position{})
		}
	}
	for key, audience := range candidateByKey {
		if baselineByKey[key] == nil {
			c.add(ImpactCompatible, prefix+".added", owner,
				fmt.Sprintf("audience %s via %s was added", audience.Actor, audience.Via), model.Position{}, audience.Pos)
		}
	}
}

func (c *_Diff) compareAuth(owner, prefix string, baseline, candidate AuthMode, baselinePos, candidatePos model.Position) {
	if baseline == candidate {
		return
	}
	code := prefix + ".auth.changed"
	impact := ImpactDangerous
	before, after := authComparisonMode(baseline, prefix), authComparisonMode(candidate, prefix)
	switch {
	case before == after:
		impact = ImpactCompatible
	case before == AuthModeOptional && (after == AuthModeRequired || after == AuthModeAnonymous):
		code = prefix + ".auth.tightened"
	case after == AuthModeOptional && (before == AuthModeRequired || before == AuthModeAnonymous):
		code = prefix + ".auth.relaxed"
	}
	c.add(impact, code, owner, fmt.Sprintf("authentication changed from %s to %s", baseline, candidate), baselinePos, candidatePos)
}

func (c *_Diff) compareRequirement(owner, prefix string, baseline, candidate *Requirement, baselinePos, candidatePos model.Position) {
	if reflect.DeepEqual(baseline, candidate) {
		return
	}
	code := prefix + ".require.changed"
	message := "permission requirement changed"
	if baseline == nil {
		code, message = prefix+".require.added", "permission requirement was added"
	} else if candidate == nil {
		code, message = prefix+".require.removed", "permission requirement was removed"
	}
	c.add(ImpactDangerous, code, owner, message, baselinePos, candidatePos)
}

// Legacy spellings are compared by their runtime meaning without changing schema values.
func authComparisonMode(mode AuthMode, declaration string) AuthMode {
	switch mode {
	case AuthModeAuth:
		return AuthModeRequired
	case AuthModeNoAuth:
		if declaration == "web" {
			return AuthModeOff
		}
		return AuthModeOptional
	case "", AuthModeUnset:
		switch declaration {
		case "service":
			return AuthModeRequired
		case "method":
			return AuthModeInherit
		default:
			// Legacy web defaults depended on actor configuration, so no explicit
			// mode is unconditionally equivalent.
			return AuthModeUnset
		}
	default:
		return mode
	}
}
