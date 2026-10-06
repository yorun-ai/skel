package projection

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/schema"
)

func projectRequirement(value *model.PermissionRequire) *schema.Requirement {
	if value == nil {
		return nil
	}
	return projectRequirementExpr(value.Expr)
}

func projectRequirementExpr(value *model.PermissionExpr) *schema.Requirement {
	if value == nil {
		return nil
	}
	mode := schema.RequirementMode(value.Mode)
	if mode == "" && value.Check != nil {
		mode = schema.RequirementModeReference
	}
	result := &schema.Requirement{Mode: mode, Code: value.Code}
	if value.Check != nil {
		arguments := make([]*schema.RequirementCheckArgument, 0, len(value.Check.Arguments))
		for _, argument := range value.Check.Arguments {
			arguments = append(arguments, &schema.RequirementCheckArgument{Name: argument.Name, JSONPath: argument.JsonPath, Type: projectType(argument.Type)})
		}
		result.Check = &schema.RequirementCheck{
			Resource: value.Check.ResourceSkelName, Action: value.Check.ActionName,
			Check: value.Check.CheckName, Arguments: arguments,
		}
	}
	for _, child := range value.Children {
		result.Children = append(result.Children, projectRequirementExpr(child))
	}
	return result
}

func projectType(value *model.Type) *schema.Type {
	if value == nil {
		return nil
	}
	result := &schema.Type{Nullable: value.Nullable}
	switch value.Kind {
	case model.TypeKindUnresolvedReference:
		result.Kind = schema.TypeKindImportedReference
		result.Name = value.SkelName
		if value.ExternalAlias != "" {
			result.Name = value.ExternalAlias + "." + value.SkelName
		}
	case model.TypeKindScalar:
		result.Kind = schema.TypeKindScalar
		result.Name = strings.ToLower(value.Scalar.Name())
	case model.TypeKindList:
		result.Kind = schema.TypeKindList
		result.Element = projectType(value.List.Value)
	case model.TypeKindMap:
		result.Kind = schema.TypeKindMap
		result.Key = projectType(value.Map.Key)
		result.Value = projectType(value.Map.Value)
	case model.TypeKindEnum:
		result.Kind = schema.TypeKindEnum
		result.Name = value.SkelName
	case model.TypeKindData:
		result.Kind = schema.TypeKindData
		if value.Data != nil {
			switch value.Data.Kind {
			case model.DataKindConfig:
				result.Kind = schema.TypeKindConfig
			case model.DataKindEvent:
				result.Kind = schema.TypeKindEvent
			}
		}
		result.Name = value.SkelName
	case model.TypeKindTypeParameter:
		result.Kind = schema.TypeKindTypeParameter
		if value.TypeParameter != nil {
			result.Name = value.TypeParameter.Name
		}
	default:
		result.Kind = schema.TypeKind(fmt.Sprintf("unknown:%d", value.Kind))
	}
	for _, argument := range value.TypeArguments {
		result.Arguments = append(result.Arguments, projectType(argument))
	}
	return result
}

func metadata(description string, deprecated bool, reason string) schema.Metadata {
	return schema.Metadata{Description: description, Deprecated: deprecated, DeprecatedReason: reason}
}

func normalizedAuth(value model.AuthMode) schema.AuthMode {
	switch value {
	case "", model.AuthModeUnset:
		return schema.AuthModeInherit
	case model.AuthModeAuth:
		return schema.AuthModeRequired
	case model.AuthModeNoAuth:
		return schema.AuthModeOptional
	default:
		return schema.AuthMode(value)
	}
}
