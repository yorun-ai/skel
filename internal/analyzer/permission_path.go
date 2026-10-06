package analyzer

import (
	"reflect"
	"strings"

	"go.yorun.ai/skel/internal/model"
)

func resolveMethodArgumentJsonPath(reporter *_DiagnosticReporter, method *model.Method, path string) (*model.Type, bool) {
	parts, valid := parsePermissionCheckJsonPath(reporter, path)
	if !valid || len(parts) == 0 {
		return nil, false
	}
	for _, arg := range method.Arguments {
		if reporter.cancelled() {
			break
		}
		if arg.Name == parts[0].Name {
			return resolveJsonPathPartType(reporter, arg.Type, parts[0], parts[1:], path)
		}
	}
	reporter.reportReferencef(`require check argument path %s references undefined input argument "%s"`, path, parts[0].Name)
	return nil, false
}

type _PermissionCheckPathPart struct {
	Name     string
	Wildcard bool
}

func parsePermissionCheckJsonPath(reporter *_DiagnosticReporter, path string) ([]_PermissionCheckPathPart, bool) {
	rawParts := strings.Split(path, ".")
	parts := make([]_PermissionCheckPathPart, 0, len(rawParts))
	wildcardCount := 0
	valid := reporter.check(path != "", "empty require check argument path")
	for _, rawPart := range rawParts {
		if reporter.cancelled() {
			break
		}
		partValid := reporter.check(rawPart != "", "require check argument path %s contains empty field", path)
		part := _PermissionCheckPathPart{Name: rawPart}
		if strings.HasSuffix(rawPart, "[*]") {
			part.Name = strings.TrimSuffix(rawPart, "[*]")
			part.Wildcard = true
			wildcardCount++
		}
		partValid = reporter.check(part.Name != "", "require check argument path %s contains empty field", path) && partValid
		partValid = reporter.check(!strings.ContainsAny(part.Name, "[]"),
			"require check argument path %s only supports field cascade and one [*]", path) && partValid
		valid = partValid && valid
		parts = append(parts, part)
	}
	valid = reporter.check(wildcardCount <= 1, "require check argument path %s supports at most one [*]", path) && valid
	if len(parts) > 0 {
		valid = reporter.checkNot(parts[len(parts)-1].Wildcard,
			"require check argument path %s cannot end with [*], use the list field path directly", path) && valid
	}
	return parts, valid
}

func resolveJsonPathType(reporter *_DiagnosticReporter, type_ *model.Type, parts []_PermissionCheckPathPart, fullPath string) (*model.Type, bool) {
	if len(parts) == 0 {
		return type_, true
	}
	if !reporter.check(type_ != nil && type_.Kind == model.TypeKindData,
		"require check argument path %s cannot select member on %s", fullPath, typeName(type_)) {
		return nil, false
	}
	for _, member := range type_.Data.Members {
		if reporter.cancelled() {
			break
		}
		if member.Name == parts[0].Name {
			return resolveJsonPathPartType(reporter, member.Type, parts[0], parts[1:], fullPath)
		}
	}
	reporter.reportReferencef(`require check argument path %s references undefined data member "%s"`, fullPath, parts[0].Name)
	return nil, false
}

func resolveJsonPathPartType(reporter *_DiagnosticReporter, type_ *model.Type, part _PermissionCheckPathPart, remainingParts []_PermissionCheckPathPart, fullPath string) (*model.Type, bool) {
	if !part.Wildcard {
		return resolveJsonPathType(reporter, type_, remainingParts, fullPath)
	}
	if !reporter.check(type_ != nil && type_.Kind == model.TypeKindList,
		"require check argument path %s can only use [*] on list, got %s", fullPath, typeName(type_)) {
		return nil, false
	}
	valueType, valid := resolveJsonPathType(reporter, type_.List.Value, remainingParts, fullPath)
	if !valid {
		return nil, false
	}
	return &model.Type{Kind: model.TypeKindList, List: &model.ListType{Value: valueType}}, true
}

func typeName(type_ *model.Type) string {
	if type_ == nil {
		return "<unknown>"
	}
	return type_.Name()
}

func typeEqual(a *model.Type, b *model.Type) bool {
	if a == nil || b == nil || a.Kind != b.Kind || a.Nullable != b.Nullable || a.Scalar != b.Scalar || a.SkelName != b.SkelName {
		return false
	}
	switch a.Kind {
	case model.TypeKindList:
		return typeEqual(a.List.Value, b.List.Value)
	case model.TypeKindMap:
		return typeEqual(a.Map.Key, b.Map.Key) && typeEqual(a.Map.Value, b.Map.Value)
	default:
		return reflect.DeepEqual(typeArgumentNames(a), typeArgumentNames(b))
	}
}

func typeArgumentNames(type_ *model.Type) []string {
	values := make([]string, 0, len(type_.TypeArguments))
	for _, argument := range type_.TypeArguments {
		values = append(values, argument.Name())
	}
	return values
}
