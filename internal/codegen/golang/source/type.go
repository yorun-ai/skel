package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

type _Types struct{ bindings common.TypeBindings }

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
		Plain:        common.ChooseString(p.Nullable, "*"+plain, plain),
		Imports:      cloneImports(valueType.Imports),
		DefaultValue: common.ChooseString(p.Nullable, "nil", plain+"{}"),
	}
}

func (r _Types) castMapType(p *model.Type) *Type {
	keyType := r.castType(p.Map.Key)
	valueType := r.castType(p.Map.Value)
	plain := fmt.Sprintf("map[%s]%s", keyType.Plain, valueType.Plain)
	return &Type{
		Plain:        common.ChooseString(p.Nullable, "*"+plain, plain),
		Imports:      collectTypeImports(keyType, valueType),
		DefaultValue: common.ChooseString(p.Nullable, "nil", plain+"{}"),
	}
}

func (r _Types) castEnumType(p *model.Type) *Type {
	p = r.bindings.Type(p)
	enumName := transEnumName(p.Enum)
	unspecifiedItemName := transUnspecifiedItemName(p.Enum)
	imports := []*Import(nil)
	if p.ExternalImportPath != "" {
		enumName = p.ExternalAlias + "." + enumName
		unspecifiedItemName = p.ExternalAlias + "." + unspecifiedItemName
		imports = []*Import{{Path: p.ExternalImportPath, Alias: goImportAlias(p)}}
	}
	return &Type{
		Plain:        common.ChooseString(p.Nullable, fmt.Sprintf("*%s", enumName), enumName),
		Imports:      imports,
		DefaultValue: common.ChooseString(p.Nullable, "nil", unspecifiedItemName),
	}
}

func (r _Types) castDataType(p *model.Type) *Type {
	p = r.bindings.Type(p)
	structName := transDataName(p.Data)
	imports := []*Import(nil)
	if p.ExternalImportPath != "" {
		structName = p.ExternalAlias + "." + structName
		imports = []*Import{{Path: p.ExternalImportPath, Alias: goImportAlias(p)}}
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
		Plain:        common.ChooseString(p.Nullable, fmt.Sprintf("*%s", structName), structName),
		Imports:      imports,
		DefaultValue: common.ChooseString(p.Nullable, "nil", fmt.Sprintf("%s{}", structName)),
	}
}

func goImportAlias(p *model.Type) string {
	defaultAlias := p.ExternalDomain[strings.LastIndex(p.ExternalDomain, ".")+1:]
	generatedAlias := p.ExternalDomain != "" && p.ExternalAlias != defaultAlias+"pub" && p.ExternalAlias != defaultAlias+"api"
	if p.ExternalAliasExplicit || generatedAlias {
		return p.ExternalAlias
	}
	return ""
}

func (r _Types) castTypeParameter(p *model.Type) *Type {
	return &Type{
		Plain:        common.ChooseString(p.Nullable, "*"+p.TypeParameter.Name, p.TypeParameter.Name),
		DefaultValue: common.ChooseString(p.Nullable, "nil", ""),
	}
}
