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
		body = "list<" + typeKey(domain, value.List.Value) + ">"
	case schema.TypeKindMap:
		body = "map<" + typeKey(domain, value.Map.Key) + "," + typeKey(domain, value.Map.Value) + ">"
	case schema.TypeKindTypeParameter:
		body = "parameter:" + strconv.Quote(value.TypeParameter.Name)
	default:
		kind := fmt.Sprint(value.Kind)
		if value.Kind == schema.TypeKindData && value.Data != nil {
			kind += ":" + string(value.Data.Kind)
		}
		body = kind + ":" + strconv.Quote(typeName(domain, value))
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
		result = "list<" + typeDisplay(domain, value.List.Value) + ">"
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
	code := value.Code
	if resource, action, ok := strings.Cut(code, ":"); ok {
		code = referenceName(domain, resource) + ":" + action
	}
	result := string(value.Mode) + ":" + strconv.Quote(code)
	if check := value.Check; check != nil {
		result += "[" + strconv.Quote(referenceName(domain, check.ResourceSkelName)) + "," + strconv.Quote(check.ActionName) + "," + strconv.Quote(check.CheckName)
		for _, arg := range check.Arguments {
			result += "," + strconv.Quote(arg.Name) + ":" + strconv.Quote(arg.JsonPath) + ":" + typeKey(domain, arg.Type)
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
	before := requirementConjuncts(c.baseline, beforeService, beforeMethod)
	after := requirementConjuncts(c.candidate, afterService, afterMethod)
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
