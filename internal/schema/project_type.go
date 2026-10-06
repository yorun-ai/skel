package schema

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/model"
)

func projectRequirement(value *model.PermissionRequire) *Requirement {
	if value == nil {
		return nil
	}
	return projectRequirementExpr(value.Expr)
}

func projectRequirementExpr(value *model.PermissionExpr) *Requirement {
	if value == nil {
		return nil
	}
	mode := RequirementMode(value.Mode)
	if mode == "" && value.Check != nil {
		mode = RequirementModeReference
	}
	result := &Requirement{Mode: mode, Code: value.Code}
	if value.Check != nil {
		arguments := make([]*RequirementCheckArgument, 0, len(value.Check.Arguments))
		for _, argument := range value.Check.Arguments {
			arguments = append(arguments, &RequirementCheckArgument{Name: argument.Name, JSONPath: argument.JsonPath, Type: projectType(argument.Type)})
		}
		result.Check = &RequirementCheck{
			Resource: value.Check.ResourceSkelName, Action: value.Check.ActionName,
			Check: value.Check.CheckName, Arguments: arguments,
		}
	}
	for _, child := range value.Children {
		result.Children = append(result.Children, projectRequirementExpr(child))
	}
	return result
}

func projectType(value *model.Type) *Type {
	if value == nil {
		return nil
	}
	result := &Type{Nullable: value.Nullable}
	switch value.Kind {
	case model.TypeKindUnresolvedReference:
		result.Kind = TypeKindImportedReference
		result.Name = value.SkelName
		if value.ExternalAlias != "" {
			result.Name = value.ExternalAlias + "." + value.SkelName
		}
	case model.TypeKindScalar:
		result.Kind = TypeKindScalar
		result.Name = strings.ToLower(value.Scalar.Name())
	case model.TypeKindList:
		result.Kind = TypeKindList
		result.Element = projectType(value.List.Value)
	case model.TypeKindMap:
		result.Kind = TypeKindMap
		result.Key = projectType(value.Map.Key)
		result.Value = projectType(value.Map.Value)
	case model.TypeKindEnum:
		result.Kind = TypeKindEnum
		result.Name = value.SkelName
	case model.TypeKindData:
		result.Kind = TypeKindData
		if value.Data != nil {
			switch value.Data.Kind {
			case model.DataKindConfig:
				result.Kind = TypeKindConfig
			case model.DataKindEvent:
				result.Kind = TypeKindEvent
			}
		}
		result.Name = value.SkelName
	case model.TypeKindTypeParameter:
		result.Kind = TypeKindTypeParameter
		if value.TypeParameter != nil {
			result.Name = value.TypeParameter.Name
		}
	default:
		result.Kind = TypeKind(fmt.Sprintf("unknown:%d", value.Kind))
	}
	for _, argument := range value.TypeArguments {
		result.Arguments = append(result.Arguments, projectType(argument))
	}
	return result
}

func metadata(description string, deprecated bool, reason string) Metadata {
	return Metadata{Description: description, Deprecated: deprecated, DeprecatedReason: reason}
}

func normalizedAuth(value model.AuthMode) AuthMode {
	switch value {
	case "", model.AuthModeUnset:
		return AuthModeInherit
	case model.AuthModeAuth:
		return AuthModeRequired
	case model.AuthModeNoAuth:
		return AuthModeOptional
	default:
		return AuthMode(value)
	}
}
