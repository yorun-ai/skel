package analyzer

import (
	"go.yorun.ai/skel/internal/symbol"
	"go.yorun.ai/skel/schema"
)

const typeKindNone schema.TypeKind = 0

const scalarNone schema.Scalar = 0

var (
	mapKeyTypes       = []schema.TypeKind{schema.TypeKindScalar, schema.TypeKindEnum}
	mapKeyScalarTypes = []schema.Scalar{schema.ScalarInt, schema.ScalarString, schema.ScalarUUID}
)

type _RefContext struct {
	enums                  map[string]*schema.Enum
	dataList               map[string]*schema.Data
	typeParameters         map[string]*schema.TypeParameter
	imports                map[string]*_DomainImport
	invalidData            map[*schema.Data]bool
	unavailable            map[string]bool
	allowUnresolvedImports bool
}

func fixTypeRef(reporter *_DiagnosticReporter, t *schema.Type, refCtx *_RefContext) bool {
	// 1. fix Enum/Data/TypeParameter type references
	// 2. check map key type (int/string/uuid/Enum)

	if reporter.cancelled() {
		return false
	}
	if t == nil {
		return true
	}

	switch t.Kind {
	case schema.TypeKindUnresolvedReference:
		resolved := refCtx.bindType(t)
		refName := t.SkelName
		refQualifier := t.ExternalAlias
		if refQualifier != "" {
			import_ := refCtx.imports[refQualifier]
			if import_ == nil && refCtx.allowUnresolvedImports {
				valid := true
				for _, typeArg := range t.TypeArguments {
					if reporter.cancelled() {
						break
					}
					valid = fixTypeRef(reporter, typeArg, refCtx) && valid
				}
				return valid
			}
			if !reporter.checkReference(import_ != nil, "%s import alias %s not found", t.Pos, refQualifier) {
				return false
			}
			enum := import_.Domain.enumsMap[refName]
			dataType := import_.Domain.dataMap[refName]
			if !reporter.checkReference(resolved.Status == symbol.Resolved, "%s definition of %s.%s not found", t.Pos, refQualifier, refName) {
				return false
			}
			if resolved.Kind == symbol.Enum {
				if !reporter.check(enum.Pub, "%s imported enum %s.%s is not public", t.Pos, import_.Schema.Alias, refName) {
					return false
				}
				t.Kind = schema.TypeKindEnum
				t.Enum = enum
				t.SkelName = enum.SkelName
				t.ExternalDomain = import_.Domain.name
				t.ExternalAlias = import_.Schema.Alias
				t.ExternalAliasExplicit = import_.Schema.ExplicitAlias
				return true
			}
			if !checkDataValueType(reporter, t, dataType) {
				return false
			}
			if !reporter.check(dataType.Pub, "%s imported data %s.%s is not public", t.Pos, import_.Schema.Alias, refName) {
				return false
			}
			t.Kind = schema.TypeKindData
			t.Data = dataType
			t.SkelName = dataType.SkelName
			t.ExternalAlias = import_.Schema.Alias
			t.ExternalDomain = import_.Domain.name
			t.ExternalAliasExplicit = import_.Schema.ExplicitAlias
			valid := true
			for _, typeArg := range t.TypeArguments {
				if reporter.cancelled() {
					break
				}
				valid = fixTypeRef(reporter, typeArg, refCtx) && valid
			}
			return checkTypeArguments(reporter, t, refName) && valid
		}
		if refCtx.unavailable[refName] {
			return false
		}
		enum := refCtx.enums[refName]
		dataType := refCtx.dataList[refName]
		param := refCtx.typeParameters[refName]
		if dataType != nil && refCtx.invalidData[dataType] {
			return false
		}
		if !reporter.checkReference(resolved.Status == symbol.Resolved, "%s definition of %s not found", t.Pos, refName) {
			return false
		}
		if resolved.Kind == symbol.Enum {
			t.Kind = schema.TypeKindEnum
			t.Enum = enum
			t.SkelName = enum.SkelName
			return true
		}
		if resolved.Kind == symbol.Data {
			if !checkDataValueType(reporter, t, dataType) {
				return false
			}
			t.Kind = schema.TypeKindData
			t.Data = dataType
			t.SkelName = dataType.SkelName
			valid := true
			for _, typeArg := range t.TypeArguments {
				if reporter.cancelled() {
					break
				}
				valid = fixTypeRef(reporter, typeArg, refCtx) && valid
			}
			return checkTypeArguments(reporter, t, refName) && valid
		}
		t.Kind = schema.TypeKindTypeParameter
		t.TypeParameter = param
		return true

	case schema.TypeKindList:
		return fixTypeRef(reporter, t.List.Value, refCtx)

	case schema.TypeKindMap:
		keyValid := fixTypeRef(reporter, t.Map.Key, refCtx)
		if keyValid && !(refCtx.allowUnresolvedImports && t.Map.Key.Kind == schema.TypeKindUnresolvedReference) {
			keyValid = checkTypeCanBeMapKey(reporter, t.Map.Key)
		}
		return fixTypeRef(reporter, t.Map.Value, refCtx) && keyValid
	}
	return true
}

func checkTypeArguments(reporter *_DiagnosticReporter, t *schema.Type, refName string) bool {
	referencePos := t.ReferencePos
	if referencePos == (schema.Position{}) {
		referencePos = t.Pos
	}
	if len(t.Data.TypeParameters) == 0 {
		return reporter.check(len(t.TypeArguments) == 0,
			"%s data %s do not support type argument(s)", referencePos, refName)
	}
	valid := reporter.check(len(t.TypeArguments) > 0,
		"%s generic data %s need type argument(s)", referencePos, refName)
	valid = reporter.check(len(t.Data.TypeParameters) == len(t.TypeArguments),
		"%s generic data %s have mismatched type arguments(s), found=%d, expected=%d",
		referencePos, refName, len(t.TypeArguments), len(t.Data.TypeParameters)) && valid
	return valid
}

func checkDataValueType(reporter *_DiagnosticReporter, kind *schema.Type, data *schema.Data) bool {
	return reporter.check(data.Kind == schema.DataKindData,
		"%s %s %s cannot be used as a value type", kind.Pos, data.Kind, data.Name)
}
