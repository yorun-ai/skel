package diff

import (
	"fmt"
	"slices"

	"go.yorun.ai/skel/schema"
)

func (c *_Diff) compareService(owner string, baseline, candidate *schema.Service) {
	if baseline.Ext != candidate.Ext {
		c.add(ImpactBreaking, "service.ext.changed", owner, "service contract direction changed", schema.Position{}, schema.Position{})
	}
	if baseline.Api != candidate.Api {
		c.add(ImpactBreaking, "service.api.changed", owner, "service invocation boundary changed", schema.Position{}, schema.Position{})
	}
	baselineByName := methodsByName(baseline.Methods)
	candidateByName := methodsByName(candidate.Methods)
	c.compareAudiences(owner, "service.audience", baseline.Audiences, candidate.Audiences)
	authChangeStart := len(c.report.Changes)
	c.compareAuth(owner, "service", baseline.NormalizedAuth(), candidate.NormalizedAuth(), schema.Position{}, schema.Position{})
	// A service default only affects methods that inherit it. Explicit method
	// policies can preserve every existing interaction despite a default change.
	if len(c.report.Changes) > authChangeStart && len(baseline.Methods) > 0 {
		impact := ImpactCompatible
		for _, method := range baseline.Methods {
			if other := candidateByName[method.Name]; other != nil {
				_, current := authImpact(c.effectivePolicy(baseline, method).AuthMode, c.effectivePolicy(candidate, other).AuthMode)
				if impactOrder(current) < impactOrder(impact) {
					impact = current
				}
			}
		}
		c.report.Changes[authChangeStart].Impact = impact
	}
	requirementStart := len(c.report.Changes)
	c.compareRequirement(owner, "service", baseline.Require, candidate.Require, schema.Position{}, schema.Position{})
	if len(c.report.Changes) > requirementStart && len(baseline.Methods) > 0 {
		impact := ImpactCompatible
		for _, method := range baseline.Methods {
			if other := candidateByName[method.Name]; other != nil {
				current := c.requirementImpact(baseline.Require, method.Require, candidate.Require, other.Require)
				if impactOrder(current) < impactOrder(impact) {
					impact = current
				}
			}
		}
		c.report.Changes[requirementStart].Impact = impact
	}
	for _, method := range baseline.Methods {
		other := candidateByName[method.Name]
		symbol := owner + "." + method.Name
		if other == nil {
			c.add(ImpactBreaking, "service.method.removed", symbol, fmt.Sprintf("service method %s was removed", method.Name), method.Pos, schema.Position{})
			continue
		}
		c.compareMethod(symbol, method, other, baseline, candidate)
	}
	for _, method := range candidate.Methods {
		if baselineByName[method.Name] == nil {
			c.add(ImpactCompatible, "service.method.added", owner+"."+method.Name,
				fmt.Sprintf("service method %s was added", method.Name), schema.Position{}, method.Pos)
		}
	}
}

func (c *_Diff) compareMethod(owner string, baseline, candidate *schema.Method, services ...*schema.Service) {
	before, after := baseline.NormalizedAuth(), candidate.NormalizedAuth()
	if len(services) == 2 {
		before = c.effectivePolicy(services[0], baseline).AuthMode
		after = c.effectivePolicy(services[1], candidate).AuthMode
	}
	if baseline.NormalizedAuth() != candidate.NormalizedAuth() || before != after {
		code, impact := authImpact(authComparisonMode(before, "method"), authComparisonMode(after, "method"))
		c.add(impact, "method.auth."+code, owner, fmt.Sprintf("effective authentication changed from %s to %s", before, after), baseline.Pos, candidate.Pos)
	}
	requirementStart := len(c.report.Changes)
	c.compareRequirement(owner, "method", baseline.Require, candidate.Require, baseline.Pos, candidate.Pos)
	if len(services) == 2 && len(c.report.Changes) > requirementStart {
		c.report.Changes[requirementStart].Impact = c.requirementImpact(services[0].Require, baseline.Require, services[1].Require, candidate.Require)
	}
	c.compareArguments(owner, "method.argument", baseline.Arguments, candidate.Arguments)
	if !c.sameType(baseline.ResultType, candidate.ResultType) {
		c.add(c.typeChangeImpact(baseline.ResultType, candidate.ResultType, usageOutput), "method.result.changed", owner,
			fmt.Sprintf("method result changed from %s to %s", typeDisplay(c.baseline, baseline.ResultType), typeDisplay(c.candidate, candidate.ResultType)), baseline.Pos, candidate.Pos)
	}
	if baseline.ArgumentsSensitive != candidate.ArgumentsSensitive || baseline.ResultSensitive != candidate.ResultSensitive {
		c.add(ImpactDangerous, "method.sensitive.changed", owner, "method sensitivity changed", baseline.Pos, candidate.Pos)
	}
	if baseline.Example != candidate.Example || baseline.InputDescription != candidate.InputDescription ||
		baseline.OutputDescription != candidate.OutputDescription || baseline.OutputExample != candidate.OutputExample {
		c.add(ImpactCompatible, "method.documentation.changed", owner, "method documentation or examples changed", baseline.Pos, candidate.Pos)
	}
	c.compareMetadata("method", owner, metadata(baseline.Description, baseline.Deprecated, baseline.DeprecatedReason), metadata(candidate.Description, candidate.Deprecated, candidate.DeprecatedReason), baseline.Pos, candidate.Pos)
}

func (c *_Diff) compareArguments(owner, prefix string, baseline, candidate []*schema.Argument) {
	baselineByName := argumentsByName(baseline)
	candidateByName := argumentsByName(candidate)
	for _, argument := range baseline {
		other := candidateByName[argument.Name]
		symbol := owner + "." + argument.Name
		if other == nil {
			c.add(ImpactBreaking, prefix+".removed", symbol, fmt.Sprintf("argument %s was removed", argument.Name), argument.Pos, schema.Position{})
			continue
		}
		if !c.sameType(argument.Type, other.Type) {
			c.add(c.typeChangeImpact(argument.Type, other.Type, usageInput), prefix+".type.changed", symbol,
				fmt.Sprintf("argument type changed from %s to %s", typeDisplay(c.baseline, argument.Type), typeDisplay(c.candidate, other.Type)), argument.Pos, other.Pos)
		}
		if argument.Sensitive != other.Sensitive {
			c.add(ImpactDangerous, prefix+".sensitive.changed", symbol, "argument sensitivity changed", argument.Pos, other.Pos)
		}
		if argument.Example != other.Example {
			c.add(ImpactCompatible, prefix+".example.changed", symbol, "argument example changed", argument.Pos, other.Pos)
		}
		c.compareMetadata(prefix, symbol, metadata(argument.Description, argument.Deprecated, argument.DeprecatedReason), metadata(other.Description, other.Deprecated, other.DeprecatedReason), argument.Pos, other.Pos)
	}
	for _, argument := range candidate {
		if baselineByName[argument.Name] == nil {
			c.add(ImpactBreaking, prefix+".added", owner+"."+argument.Name,
				fmt.Sprintf("required argument %s was added", argument.Name), schema.Position{}, argument.Pos)
		}
	}
	if sameNamedSet(argumentNames(baseline), argumentNames(candidate)) && !slices.Equal(argumentNames(baseline), argumentNames(candidate)) {
		c.add(ImpactBreaking, prefix+".order.changed", owner, "argument order changed", schema.Position{}, schema.Position{})
	}
}

func (c *_Diff) compareAudiences(owner, prefix string, baseline, candidate []*schema.ActorAudience) {
	baselineByKey := audiencesByKey(c.baseline, baseline)
	candidateByKey := audiencesByKey(c.candidate, candidate)
	for key, audience := range baselineByKey {
		if candidateByKey[key] == nil {
			c.add(ImpactBreaking, prefix+".removed", owner,
				fmt.Sprintf("audience %s via %s was removed", audience.Actor, audience.Via), audience.Pos, schema.Position{})
		}
	}
	for key, audience := range candidateByKey {
		if baselineByKey[key] == nil {
			c.add(ImpactCompatible, prefix+".added", owner,
				fmt.Sprintf("audience %s via %s was added", audience.Actor, audience.Via), schema.Position{}, audience.Pos)
		}
	}
}

func (c *_Diff) compareAuth(owner, prefix string, baseline, candidate schema.AuthMode, baselinePos, candidatePos schema.Position) {
	if baseline == candidate {
		return
	}
	code, impact := authImpact(authComparisonMode(baseline, prefix), authComparisonMode(candidate, prefix))
	c.add(impact, prefix+".auth."+code, owner, fmt.Sprintf("authentication changed from %s to %s", baseline, candidate), baselinePos, candidatePos)
}

func (c *_Diff) effectivePolicy(service *schema.Service, method *schema.Method) schema.EffectivePolicy {
	value, err := schema.ComputeEffectivePolicy(service, method)
	if err != nil {
		c.err = fmt.Errorf("service %s method %s effective policy: %w", service.Name, method.Name, err)
	}
	return value
}

func authImpact(before, after schema.AuthMode) (string, ImpactLevel) {
	switch {
	case before == after:
		return "changed", ImpactCompatible
	case before == schema.AuthModeOff && (after == schema.AuthModeRequired || after == schema.AuthModeOptional || after == schema.AuthModeAnonymous):
		// All validating modes reject credentials that off forwarded unchanged.
		return "tightened", ImpactBreaking
	case before == schema.AuthModeOptional && (after == schema.AuthModeRequired || after == schema.AuthModeAnonymous):
		return "tightened", ImpactBreaking
	case (before == schema.AuthModeRequired && after == schema.AuthModeAnonymous) || (before == schema.AuthModeAnonymous && after == schema.AuthModeRequired):
		return "changed", ImpactBreaking
	case after == schema.AuthModeOptional && (before == schema.AuthModeRequired || before == schema.AuthModeAnonymous):
		return "relaxed", ImpactDangerous
	default:
		return "changed", ImpactDangerous
	}
}

func (c *_Diff) compareRequirement(owner, prefix string, baseline, candidate *schema.PermissionRequire, baselinePos, candidatePos schema.Position) {
	if requirementKey(c.baseline, baseline) == requirementKey(c.candidate, candidate) {
		return
	}
	impact := c.requirementImpact(baseline, nil, candidate, nil)
	code := prefix + ".require.changed"
	message := "permission requirement changed"
	if baseline == nil {
		code, message = prefix+".require.added", "permission requirement was added"
	} else if candidate == nil {
		code, message = prefix+".require.removed", "permission requirement was removed"
	}
	c.add(impact, code, owner, message, baselinePos, candidatePos)
}

// Legacy spellings are compared by their runtime meaning without changing schema values.
func authComparisonMode(mode schema.AuthMode, declaration string) schema.AuthMode {
	switch declaration {
	case "service":
		return (&schema.Service{AuthMode: mode}).NormalizedAuth()
	case "method":
		return (&schema.Method{AuthMode: mode}).NormalizedAuth()
	default:
		if mode == "" || mode == schema.AuthModeUnset {
			// A raw legacy web declaration has no unconditional mode until analyzed.
			return schema.AuthModeUnset
		}
		return (&schema.Web{AuthMode: mode}).NormalizedAuth()
	}
}
