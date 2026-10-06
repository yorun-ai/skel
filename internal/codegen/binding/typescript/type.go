package typescript

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/schema"
)

type _Types struct{ bindings binding.TypeBindings }

type _Type struct {
	Plain string
}

type _TypeImport struct {
	Alias string
	Path  string
}

func (r _Types) castType(p *schema.Type) *_Type {
	if p == nil {
		return nil
	}

	switch p.Kind {
	case schema.TypeKindScalar:
		return castScalarType(p)
	case schema.TypeKindList:
		return r.castListType(p)
	case schema.TypeKindMap:
		return r.castMapType(p)
	case schema.TypeKindEnum:
		return r.castEnumType(p)
	case schema.TypeKindData:
		return r.castDataType(p)
	case schema.TypeKindTypeParameter:
		return r.castTypeParameter(p)
	}

	return nil
}

func castScalarType(p *schema.Type) *_Type {
	switch p.Scalar {
	case schema.ScalarInt:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "number | null", "number"),
		}
	case schema.ScalarFloat:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "number | null", "number"),
		}
	case schema.ScalarBoolean:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "boolean | null", "boolean"),
		}
	case schema.ScalarString:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarDecimal:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarBinary:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "Uint8Array | null", "Uint8Array"),
		}
	case schema.ScalarTimestamp:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarDuration:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarLocalDate:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarLocalTime:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarLocalDateTime:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarUUID:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	case schema.ScalarJSON:
		return &_Type{
			Plain: binding.ChooseString(p.Nullable, "string | null", "string"),
		}
	}
	return nil
}

func (r _Types) castListType(p *schema.Type) *_Type {
	valueType := r.castType(p.List.Value)
	arrayType := fmt.Sprintf("Array<%s>", valueType.Plain)
	return &_Type{
		Plain: binding.ChooseString(p.Nullable, fmt.Sprintf("%s | null", arrayType), arrayType),
	}
}

func (r _Types) castMapType(p *schema.Type) *_Type {
	keyType := r.castType(p.Map.Key)
	valueType := r.castType(p.Map.Value)
	mapType := fmt.Sprintf("Record<%s, %s>", keyType.Plain, valueType.Plain)
	return &_Type{
		Plain: binding.ChooseString(p.Nullable, fmt.Sprintf("%s | null", mapType), mapType),
	}
}

func (r _Types) castEnumType(p *schema.Type) *_Type {
	importBinding := r.bindings[p]
	enumName := transEnumName(p.Enum)
	if importBinding != nil && importBinding.Alias != "" {
		enumName = fmt.Sprintf("%s.%s", importBinding.Alias, enumName)
	}
	return &_Type{
		Plain: binding.ChooseString(p.Nullable, enumName+" | null", enumName),
	}
}

func (r _Types) castDataType(p *schema.Type) *_Type {
	importBinding := r.bindings[p]
	dataName := transDataName(p.Data)
	if importBinding != nil && importBinding.Alias != "" {
		dataName = fmt.Sprintf("%s.%s", importBinding.Alias, dataName)
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
		Plain: binding.ChooseString(p.Nullable, dataName+" | null", dataName),
	}
}

func (r _Types) castTypeParameter(p *schema.Type) *_Type {
	return &_Type{
		Plain: binding.ChooseString(p.Nullable, p.TypeParameter.Name+" | null", p.TypeParameter.Name),
	}
}
