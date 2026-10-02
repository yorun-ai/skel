package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

type Type struct {
	Plain string
}

type TypeImport struct {
	Alias string
	Path  string
}

func castType(p *model.Type) *Type {
	if p == nil {
		return nil
	}

	switch p.Kind {
	case model.TypeKindScalar:
		return castScalarType(p)
	case model.TypeKindList:
		return castListType(p)
	case model.TypeKindMap:
		return castMapType(p)
	case model.TypeKindEnum:
		return castEnumType(p)
	case model.TypeKindData:
		return castDataType(p)
	case model.TypeKindTypeParameter:
		return castTypeParameter(p)
	}

	return nil
}

func castScalarType(p *model.Type) *Type {
	switch p.Scalar {
	case model.ScalarInt:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "number | null", "number"),
		}
	case model.ScalarFloat:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "number | null", "number"),
		}
	case model.ScalarBoolean:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "boolean | null", "boolean"),
		}
	case model.ScalarString:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarDecimal:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarBinary:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "Uint8Array | null", "Uint8Array"),
		}
	case model.ScalarTimestamp:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarDuration:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarLocalDate:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarLocalTime:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarLocalDateTime:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarUUID:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarJSON:
		return &Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	}
	return nil
}

func castListType(p *model.Type) *Type {
	valueType := castType(p.List.Value)
	arrayType := fmt.Sprintf("Array<%s>", valueType.Plain)
	return &Type{
		Plain: common.ChooseString(p.Nullable, fmt.Sprintf("%s | null", arrayType), arrayType),
	}
}

func castMapType(p *model.Type) *Type {
	keyType := castType(p.Map.Key)
	valueType := castType(p.Map.Value)
	mapType := fmt.Sprintf("Record<%s, %s>", keyType.Plain, valueType.Plain)
	return &Type{
		Plain: common.ChooseString(p.Nullable, fmt.Sprintf("%s | null", mapType), mapType),
	}
}

func castEnumType(p *model.Type) *Type {
	enumName := transEnumName(p.Enum)
	if p.ExternalAlias != "" {
		enumName = fmt.Sprintf("%s.%s", p.ExternalAlias, enumName)
	}
	return &Type{
		Plain: common.ChooseString(p.Nullable, enumName+" | null", enumName),
	}
}

func castDataType(p *model.Type) *Type {
	dataName := transDataName(p.Data)
	if p.ExternalAlias != "" {
		dataName = fmt.Sprintf("%s.%s", p.ExternalAlias, dataName)
	}
	if len(p.TypeArguments) > 0 {
		typeArgNames := make([]string, 0, len(p.TypeArguments))
		for _, typeArg := range p.TypeArguments {
			castedTypeArg := castType(typeArg)
			typeArgNames = append(typeArgNames, castedTypeArg.Plain)
		}
		dataName = fmt.Sprintf("%s<%s>", dataName, strings.Join(typeArgNames, ", "))
	}
	return &Type{
		Plain: common.ChooseString(p.Nullable, dataName+" | null", dataName),
	}
}

func castTypeParameter(p *model.Type) *Type {
	return &Type{
		Plain: common.ChooseString(p.Nullable, p.TypeParameter.Name+" | null", p.TypeParameter.Name),
	}
}
