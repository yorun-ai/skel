package typescript

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

type _Types struct{ bindings common.TypeBindings }

type _Type struct {
	Plain string
}

type _TypeImport struct {
	Alias string
	Path  string
}

func (r _Types) castType(p *model.Type) *_Type {
	if p == nil {
		return nil
	}

	switch p.Kind {
	case model.TypeKindScalar:
		return castScalarType(p)
	case model.TypeKindList:
		return r.castListType(p)
	case model.TypeKindMap:
		return r.castMapType(p)
	case model.TypeKindEnum:
		return r.castEnumType(p)
	case model.TypeKindData:
		return r.castDataType(p)
	case model.TypeKindTypeParameter:
		return r.castTypeParameter(p)
	}

	return nil
}

func castScalarType(p *model.Type) *_Type {
	switch p.Scalar {
	case model.ScalarInt:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "number | null", "number"),
		}
	case model.ScalarFloat:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "number | null", "number"),
		}
	case model.ScalarBoolean:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "boolean | null", "boolean"),
		}
	case model.ScalarString:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarDecimal:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarBinary:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "Uint8Array | null", "Uint8Array"),
		}
	case model.ScalarTimestamp:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarDuration:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarLocalDate:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarLocalTime:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarLocalDateTime:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarUUID:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	case model.ScalarJSON:
		return &_Type{
			Plain: common.ChooseString(p.Nullable, "string | null", "string"),
		}
	}
	return nil
}

func (r _Types) castListType(p *model.Type) *_Type {
	valueType := r.castType(p.List.Value)
	arrayType := fmt.Sprintf("Array<%s>", valueType.Plain)
	return &_Type{
		Plain: common.ChooseString(p.Nullable, fmt.Sprintf("%s | null", arrayType), arrayType),
	}
}

func (r _Types) castMapType(p *model.Type) *_Type {
	keyType := r.castType(p.Map.Key)
	valueType := r.castType(p.Map.Value)
	mapType := fmt.Sprintf("Record<%s, %s>", keyType.Plain, valueType.Plain)
	return &_Type{
		Plain: common.ChooseString(p.Nullable, fmt.Sprintf("%s | null", mapType), mapType),
	}
}

func (r _Types) castEnumType(p *model.Type) *_Type {
	p = r.bindings.Type(p)
	enumName := transEnumName(p.Enum)
	if p.ExternalAlias != "" {
		enumName = fmt.Sprintf("%s.%s", p.ExternalAlias, enumName)
	}
	return &_Type{
		Plain: common.ChooseString(p.Nullable, enumName+" | null", enumName),
	}
}

func (r _Types) castDataType(p *model.Type) *_Type {
	p = r.bindings.Type(p)
	dataName := transDataName(p.Data)
	if p.ExternalAlias != "" {
		dataName = fmt.Sprintf("%s.%s", p.ExternalAlias, dataName)
	}
	if len(p.TypeArguments) > 0 {
		typeArgNames := make([]string, 0, len(p.TypeArguments))
		for _, typeArg := range p.TypeArguments {
			castedTypeArg := r.castType(typeArg)
			typeArgNames = append(typeArgNames, castedTypeArg.Plain)
		}
		dataName = fmt.Sprintf("%s<%s>", dataName, strings.Join(typeArgNames, ", "))
	}
	return &_Type{
		Plain: common.ChooseString(p.Nullable, dataName+" | null", dataName),
	}
}

func (r _Types) castTypeParameter(p *model.Type) *_Type {
	return &_Type{
		Plain: common.ChooseString(p.Nullable, p.TypeParameter.Name+" | null", p.TypeParameter.Name),
	}
}
