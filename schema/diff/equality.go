package diff

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"go.yorun.ai/skel/schema"
)

type _Metadata struct {
	Description      string
	Deprecated       bool
	DeprecatedReason string
}

func metadata(description string, deprecated bool, reason string) _Metadata {
	return _Metadata{description, deprecated, reason}
}

func referenceName(domain *schema.Domain, name string) string {
	if domain == nil {
		return name
	}
	return domain.ReferenceName(name)
}
func typeName(domain *schema.Domain, value *schema.Type) string {
	if domain != nil {
		return domain.TypeReferenceName(value)
	}
	return value.SkelName
}

// typeKey examines only the type expression, never the target declaration graph.
func typeKey(domain *schema.Domain, value *schema.Type) string {
	if value == nil {
		return "void"
	}
	var body string
	switch value.Kind {
	case schema.TypeKindScalar:
		body = "scalar:" + strings.ToLower(value.Scalar.Name())
	case schema.TypeKindList:
		body = "list<" + typeKey(domain, value.List.Element) + ">"
	case schema.TypeKindMap:
		body = "map<" + typeKey(domain, value.Map.Key) + "," + typeKey(domain, value.Map.Value) + ">"
	case schema.TypeKindTypeParameter:
		body = "parameter:" + strconv.Quote(value.TypeParameter.Name)
	case schema.TypeKindUnresolvedReference, schema.TypeKindData, schema.TypeKindEnum:
		// Resolution adds declaration metadata, not a new reference identity.
		// Declaration-kind changes are compared at the declaration itself.
		body = "named:" + strconv.Quote(typeName(domain, value))
	default:
		body = fmt.Sprint(value.Kind) + ":" + strconv.Quote(typeName(domain, value))
	}
	for _, argument := range value.TypeArguments {
		body += "<" + typeKey(domain, argument) + ">"
	}
	if value.Nullable {
		body += "?"
	}
	return body
}
func (c *_Diff) sameType(before, after *schema.Type) bool {
	return typeKey(c.baseline, before) == typeKey(c.candidate, after)
}

func typeDisplay(domain *schema.Domain, value *schema.Type) string {
	if value == nil {
		return "void"
	}
	var result string
	switch value.Kind {
	case schema.TypeKindScalar:
		result = strings.ToLower(value.Scalar.Name())
	case schema.TypeKindList:
		result = "list<" + typeDisplay(domain, value.List.Element) + ">"
	case schema.TypeKindMap:
		result = "map<" + typeDisplay(domain, value.Map.Key) + ", " + typeDisplay(domain, value.Map.Value) + ">"
	case schema.TypeKindTypeParameter:
		result = value.TypeParameter.Name
	default:
		result = typeName(domain, value)
	}
	if len(value.TypeArguments) != 0 {
		args := make([]string, 0, len(value.TypeArguments))
		for _, arg := range value.TypeArguments {
			args = append(args, typeDisplay(domain, arg))
		}
		result += "<" + strings.Join(args, ", ") + ">"
	}
	if value.Nullable {
		result += "?"
	}
	return result
}

func requirementKey(domain *schema.Domain, value *schema.PermissionRequire) string {
	if value == nil {
		return ""
	}
	return expressionKey(domain, value.Expression)
}
func expressionKey(domain *schema.Domain, value *schema.PermissionExpression) string {
	if value == nil {
		return ""
	}
	value = schema.ExpandPermissionExpression(value)
	code := value.Code
	if resource, action, ok := strings.Cut(code, ":"); ok {
		code = referenceName(domain, resource) + ":" + action
	}
	result := string(value.Mode) + ":" + strconv.Quote(code)
	if check := value.Check; check != nil {
		result += "[" + strconv.Quote(referenceName(domain, check.ResourceSkelName)) + "," + strconv.Quote(check.ActionName) + "," + strconv.Quote(check.CheckName)
		for _, arg := range check.Arguments {
			// Target parameter names and types are supplied by resolution. The
			// source contract specifies ordered argument paths at the call site.
			result += "," + strconv.Quote(arg.JsonPath)
		}
		result += "]"
	}
	for _, child := range value.Children {
		result += "(" + expressionKey(domain, child) + ")"
	}
	return result
}

// Service and method policies form a conjunction. Disjunctions remain opaque.
func (c *_Diff) requirementImpact(beforeService, beforeMethod, afterService, afterMethod *schema.PermissionRequire) ImpactLevel {
	beforePolicy, err := schema.ComposeRequirements(beforeService, beforeMethod)
	if err != nil {
		c.err = fmt.Errorf("baseline effective requirement: %w", err)
		return ImpactDangerous
	}
	afterPolicy, err := schema.ComposeRequirements(afterService, afterMethod)
	if err != nil {
		c.err = fmt.Errorf("candidate effective requirement: %w", err)
		return ImpactDangerous
	}
	before := requirementConjuncts(c.baseline, beforePolicy)
	after := requirementConjuncts(c.candidate, afterPolicy)
	containsBefore := true
	for _, term := range before {
		if !slices.Contains(after, term) {
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
	isDisjunction := func(term string) bool { return strings.HasPrefix(term, string(schema.PermissionRequireModeAny)+":") }
	if containsBefore && !slices.ContainsFunc(before, isDisjunction) && !slices.ContainsFunc(after, isDisjunction) {
		return ImpactBreaking
	}
	return ImpactDangerous
}
func requirementConjuncts(domain *schema.Domain, policies ...*schema.PermissionRequire) []string {
	var result []string
	var add func(*schema.PermissionExpression)
	add = func(value *schema.PermissionExpression) {
		if value == nil {
			return
		}
		value = schema.ExpandPermissionExpression(value)
		if value.Mode == schema.PermissionRequireModeAll {
			for _, child := range value.Children {
				add(child)
			}
			return
		}
		term := expressionKey(domain, value)
		if !slices.Contains(result, term) {
			result = append(result, term)
		}
	}
	for _, policy := range policies {
		if policy != nil {
			add(policy.Expression)
		}
	}
	return result
}
