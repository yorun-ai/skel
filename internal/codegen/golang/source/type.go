package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

type Type struct {
	Plain          string
	Imports        []*Import
	DefaultValue   string
	DefaultImports []*Import
}

const skelImport = "go.yorun.ai/vine/core/skel"

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

func castListType(p *model.Type) *Type {
	valueType := castType(p.List.Value)
	plain := fmt.Sprintf("[]%s", valueType.Plain)
	return &Type{
		Plain:        common.ChooseString(p.Nullable, "*"+plain, plain),
		Imports:      cloneImports(valueType.Imports),
		DefaultValue: common.ChooseString(p.Nullable, "nil", plain+"{}"),
	}
}

func castMapType(p *model.Type) *Type {
	keyType := castType(p.Map.Key)
	valueType := castType(p.Map.Value)
	plain := fmt.Sprintf("map[%s]%s", keyType.Plain, valueType.Plain)
	return &Type{
		Plain:        common.ChooseString(p.Nullable, "*"+plain, plain),
		Imports:      collectTypeImports(keyType, valueType),
		DefaultValue: common.ChooseString(p.Nullable, "nil", plain+"{}"),
	}
}

func castEnumType(p *model.Type) *Type {
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

func castDataType(p *model.Type) *Type {
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
			castedTypeArg := castType(typeArg)
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

func castTypeParameter(p *model.Type) *Type {
	return &Type{
		Plain:        common.ChooseString(p.Nullable, "*"+p.TypeParameter.Name, p.TypeParameter.Name),
		DefaultValue: common.ChooseString(p.Nullable, "nil", ""),
	}
}
