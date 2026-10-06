package schema

import (
	"fmt"
	"reflect"
	"slices"

	"go.yorun.ai/skel/internal/model"
)

func (c *_Diff) compareService(owner string, baseline, candidate *ServiceSchema) {
	if baseline.Ext != candidate.Ext {
		c.add(ImpactBreaking, "service.ext.changed", owner, "service contract direction changed", model.Position{}, model.Position{})
	}
	if baseline.Api != candidate.Api {
		c.add(ImpactBreaking, "service.api.changed", owner, "service invocation boundary changed", model.Position{}, model.Position{})
	}
	baselineByName := methodsByName(baseline.Methods)
	candidateByName := methodsByName(candidate.Methods)
	c.compareAudiences(owner, "service.audience", baseline.Audiences, candidate.Audiences)
	authChangeStart := len(c.report.Changes)
	c.compareAuth(owner, "service", baseline.Auth, candidate.Auth, model.Position{}, model.Position{})
	// A service default only affects methods that inherit it. Explicit method
	// policies can preserve every existing interaction despite a default change.
	if len(c.report.Changes) > authChangeStart && len(baseline.Methods) > 0 {
		impact := ImpactCompatible
		for _, method := range baseline.Methods {
			if other := candidateByName[method.Name]; other != nil {
				_, current := authImpact(effectiveMethodAuth(method.Auth, baseline.Auth), effectiveMethodAuth(other.Auth, candidate.Auth))
				if impactOrder(current) < impactOrder(impact) {
					impact = current
				}
			}
		}
		c.report.Changes[authChangeStart].Impact = impact
	}
	requirementStart := len(c.report.Changes)
	c.compareRequirement(owner, "service", baseline.Require, candidate.Require, model.Position{}, model.Position{})
	if len(c.report.Changes) > requirementStart && len(baseline.Methods) > 0 {
		impact := ImpactCompatible
		for _, method := range baseline.Methods {
			if other := candidateByName[method.Name]; other != nil {
				current := requirementImpact(baseline.Require, method.Require, candidate.Require, other.Require)
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
			c.add(ImpactBreaking, "service.method.removed", symbol, fmt.Sprintf("service method %s was removed", method.Name), method.Pos, model.Position{})
			continue
		}
		c.compareMethod(symbol, method, other, baseline, candidate)
	}
	for _, method := range candidate.Methods {
		if baselineByName[method.Name] == nil {
			c.add(ImpactCompatible, "service.method.added", owner+"."+method.Name,
				fmt.Sprintf("service method %s was added", method.Name), model.Position{}, method.Pos)
		}
	}
}

func (c *_Diff) compareMethod(owner string, baseline, candidate *Method, services ...*ServiceSchema) {
	before, after := baseline.Auth, candidate.Auth
	if len(services) == 2 {
		before = effectiveMethodAuth(before, services[0].Auth)
		after = effectiveMethodAuth(after, services[1].Auth)
	}
	if baseline.Auth != candidate.Auth || before != after {
		code, impact := authImpact(authComparisonMode(before, "method"), authComparisonMode(after, "method"))
		c.add(impact, "method.auth."+code, owner, fmt.Sprintf("effective authentication changed from %s to %s", before, after), baseline.Pos, candidate.Pos)
	}
	requirementStart := len(c.report.Changes)
	c.compareRequirement(owner, "method", baseline.Require, candidate.Require, baseline.Pos, candidate.Pos)
	if len(services) == 2 && len(c.report.Changes) > requirementStart {
		c.report.Changes[requirementStart].Impact = requirementImpact(services[0].Require, baseline.Require, services[1].Require, candidate.Require)
	}
	c.compareArguments(owner, "method.argument", baseline.Arguments, candidate.Arguments)
	if !reflect.DeepEqual(baseline.Result, candidate.Result) {
		c.add(typeChangeImpact(baseline.Result, candidate.Result, usageOutput), "method.result.changed", owner,
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
			c.add(typeChangeImpact(argument.Type, other.Type, usageInput), prefix+".type.changed", symbol,
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
	code, impact := authImpact(authComparisonMode(baseline, prefix), authComparisonMode(candidate, prefix))
	c.add(impact, prefix+".auth."+code, owner, fmt.Sprintf("authentication changed from %s to %s", baseline, candidate), baselinePos, candidatePos)
}

func effectiveMethodAuth(mode, service AuthMode) AuthMode {
	mode = authComparisonMode(mode, "method")
	if mode == AuthModeInherit {
		return authComparisonMode(service, "service")
	}
	return mode
}

func authImpact(before, after AuthMode) (string, ImpactLevel) {
	switch {
	case before == after:
		return "changed", ImpactCompatible
	case before == AuthModeOff && (after == AuthModeRequired || after == AuthModeOptional || after == AuthModeAnonymous):
		// All validating modes reject credentials that off forwarded unchanged.
		return "tightened", ImpactBreaking
	case before == AuthModeOptional && (after == AuthModeRequired || after == AuthModeAnonymous):
		return "tightened", ImpactBreaking
	case (before == AuthModeRequired && after == AuthModeAnonymous) || (before == AuthModeAnonymous && after == AuthModeRequired):
		return "changed", ImpactBreaking
	case after == AuthModeOptional && (before == AuthModeRequired || before == AuthModeAnonymous):
		return "relaxed", ImpactDangerous
	default:
		return "changed", ImpactDangerous
	}
}

func (c *_Diff) compareRequirement(owner, prefix string, baseline, candidate *Requirement, baselinePos, candidatePos model.Position) {
	if reflect.DeepEqual(baseline, candidate) {
		return
	}
	impact := requirementImpact(baseline, nil, candidate, nil)
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

// Compare effective service and method requirements as sets of conjuncts.
// Disjunctions remain opaque: recognizing arbitrary logical implications is
// outside this comparison, so non-equivalent expressions stay dangerous.
func requirementImpact(beforeService, beforeMethod, afterService, afterMethod *Requirement) ImpactLevel {
	before := requirementConjuncts(beforeService, beforeMethod)
	after := requirementConjuncts(afterService, afterMethod)
	containsBefore := true
	for _, left := range before {
		if !slices.ContainsFunc(after, func(right *Requirement) bool { return reflect.DeepEqual(left, right) }) {
			containsBefore = false
			break
		}
	}
	if containsBefore && len(before) == len(after) {
		return ImpactCompatible
	}
	if len(before) == 0 {
		return ImpactBreaking
	}
	isDisjunction := func(value *Requirement) bool { return value.Mode == RequirementModeAny }
	if containsBefore && !slices.ContainsFunc(before, isDisjunction) && !slices.ContainsFunc(after, isDisjunction) {
		return ImpactBreaking
	}
	return ImpactDangerous
}

// Flatten conjunctions and deduplicate terms, including across policy scopes.
func requirementConjuncts(service, method *Requirement) []*Requirement {
	var result []*Requirement
	var add func(*Requirement)
	add = func(value *Requirement) {
		if value == nil {
			return
		}
		if value.Mode == RequirementModeAll {
			for _, child := range value.Children {
				add(child)
			}
			return
		}
		if !slices.ContainsFunc(result, func(term *Requirement) bool { return reflect.DeepEqual(term, value) }) {
			result = append(result, value)
		}
	}
	add(service)
	add(method)
	return result
}
