package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/model"
)

type _Types struct{ bindings binding.TypeBindings }

type Type struct {
	Plain          string
	Imports        []*Import
	DefaultValue   string
	DefaultImports []*Import
}

const skelImport = "go.yorun.ai/vine/core/skel"

func (r _Types) castType(p *model.Type) *Type {
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

func (r _Types) castListType(p *model.Type) *Type {
	valueType := r.castType(p.List.Value)
	plain := fmt.Sprintf("[]%s", valueType.Plain)
	return &Type{
		Plain:        binding.ChooseString(p.Nullable, "*"+plain, plain),
		Imports:      cloneImports(valueType.Imports),
		DefaultValue: binding.ChooseString(p.Nullable, "nil", plain+"{}"),
	}
}

func (r _Types) castMapType(p *model.Type) *Type {
	keyType := r.castType(p.Map.Key)
	valueType := r.castType(p.Map.Value)
	plain := fmt.Sprintf("map[%s]%s", keyType.Plain, valueType.Plain)
	return &Type{
		Plain:        binding.ChooseString(p.Nullable, "*"+plain, plain),
		Imports:      collectTypeImports(keyType, valueType),
		DefaultValue: binding.ChooseString(p.Nullable, "nil", plain+"{}"),
	}
}

func (r _Types) castEnumType(p *model.Type) *Type {
	importBinding := r.bindings[p]
	enumName := transEnumName(p.Enum)
	unspecifiedItemName := transUnspecifiedItemName(p.Enum)
	imports := []*Import(nil)
	if importBinding != nil && importBinding.Path != "" {
		enumName = importBinding.Alias + "." + enumName
		unspecifiedItemName = importBinding.Alias + "." + unspecifiedItemName
		imports = []*Import{{Path: importBinding.Path, Alias: goImportAlias(importBinding)}}
	}
	return &Type{
		Plain:        binding.ChooseString(p.Nullable, fmt.Sprintf("*%s", enumName), enumName),
		Imports:      imports,
		DefaultValue: binding.ChooseString(p.Nullable, "nil", unspecifiedItemName),
	}
}

func (r _Types) castDataType(p *model.Type) *Type {
	importBinding := r.bindings[p]
	structName := transDataName(p.Data)
	imports := []*Import(nil)
	if importBinding != nil && importBinding.Path != "" {
		structName = importBinding.Alias + "." + structName
		imports = []*Import{{Path: importBinding.Path, Alias: goImportAlias(importBinding)}}
	}
	if len(p.TypeArguments) > 0 {
		typeArgNames := make([]string, 0, len(p.TypeArguments))
		typeArgTypes := make([]*Type, 0, len(p.TypeArguments))
		for _, typeArg := range p.TypeArguments {
			castedTypeArg := r.castType(typeArg)
			typeArgNames = append(typeArgNames, castedTypeArg.Plain)
			typeArgTypes = append(typeArgTypes, castedTypeArg)
		}
		imports = collectTypeImports(append(typeArgTypes, &Type{Imports: imports})...)
		structName = fmt.Sprintf("%s[%s]", structName, strings.Join(typeArgNames, ", "))
	}
	return &Type{
		Plain:        binding.ChooseString(p.Nullable, fmt.Sprintf("*%s", structName), structName),
		Imports:      imports,
		DefaultValue: binding.ChooseString(p.Nullable, "nil", fmt.Sprintf("%s{}", structName)),
	}
}

func goImportAlias(importBinding *binding.ImportBinding) string {
	defaultAlias := importBinding.Domain[strings.LastIndex(importBinding.Domain, ".")+1:]
	generatedAlias := importBinding.Domain != "" && importBinding.Alias != defaultAlias+"pub" && importBinding.Alias != defaultAlias+"api"
	if importBinding.Explicit || generatedAlias {
		return importBinding.Alias
	}
	return ""
}

func (r _Types) castTypeParameter(p *model.Type) *Type {
	return &Type{
		Plain:        binding.ChooseString(p.Nullable, "*"+p.TypeParameter.Name, p.TypeParameter.Name),
		DefaultValue: binding.ChooseString(p.Nullable, "nil", ""),
	}
}
