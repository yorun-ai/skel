package analyzer

import "go.yorun.ai/skelc/internal/model"

const typeKindNone model.TypeKind = 0

const scalarNone model.Scalar = 0

var (
	mapKeyTypes       = []model.TypeKind{model.TypeKindScalar, model.TypeKindEnum}
	mapKeyScalarTypes = []model.Scalar{model.ScalarInt, model.ScalarString, model.ScalarUUID}
)

type _RefContext struct {
	enums                  map[string]*model.Enum
	dataList               map[string]*model.Data
	typeParameters         map[string]*model.TypeParameter
	imports                map[string]*_DomainImport
	invalidData            map[*model.Data]bool
	unavailable            map[string]bool
	allowUnresolvedImports bool
}

func fixTypeRef(reporter *_DiagnosticReporter, t *model.Type, refCtx *_RefContext) bool {
	// 1. fix Enum/Data/TypeParameter type references
	// 2. check map key type (int/string/uuid/Enum)

	if t == nil {
		return true
	}

	switch t.Kind {
	case model.TypeKindUnresolvedReference:
		refName := t.SkelName
		refQualifier := t.ExternalAlias
		if refQualifier != "" {
			import_ := refCtx.imports[refQualifier]
			if import_ == nil && refCtx.allowUnresolvedImports {
				valid := true
				for _, typeArg := range t.TypeArguments {
					valid = fixTypeRef(reporter, typeArg, refCtx) && valid
				}
				return valid
			}
			if !reporter.checkReference(import_ != nil, "%s import alias %s not found", t.Pos, refQualifier) {
				return false
			}
			enum, enumOK := import_.Domain.enumsMap[refName]
			dataType, dataOK := import_.Domain.dataMap[refName]
			if !reporter.checkReference(enumOK || dataOK, "%s definition of %s.%s not found", t.Pos, refQualifier, refName) {
				return false
			}
			if enumOK {
				if !reporter.check(enum.Pub, "%s imported enum %s.%s is not public", t.Pos, import_.Model.Alias, refName) {
					return false
				}
				t.Kind = model.TypeKindEnum
				t.Enum = enum
				t.SkelName = enum.SkelName
				t.ExternalDomain = import_.Domain.name
				t.ExternalAlias = import_.Model.Alias
				t.ExternalAliasExplicit = import_.Model.ExplicitAlias
				return true
			}
			if !checkDataValueType(reporter, t, dataType) {
				return false
			}
			if !reporter.check(dataType.Pub, "%s imported data %s.%s is not public", t.Pos, import_.Model.Alias, refName) {
				return false
			}
			t.Kind = model.TypeKindData
			t.Data = dataType
			t.SkelName = dataType.SkelName
			t.ExternalAlias = import_.Model.Alias
			t.ExternalDomain = import_.Domain.name
			t.ExternalAliasExplicit = import_.Model.ExplicitAlias
			valid := true
			for _, typeArg := range t.TypeArguments {
				valid = fixTypeRef(reporter, typeArg, refCtx) && valid
			}
			return checkTypeArguments(reporter, t, refName) && valid
		}
		if refCtx.unavailable[refName] {
			return false
		}
		enum, enumOK := refCtx.enums[refName]
		dataType, dataOK := refCtx.dataList[refName]
		param, paramOK := refCtx.typeParameters[refName]
		if dataOK && refCtx.invalidData[dataType] {
			return false
		}
		if !reporter.checkReference(enumOK || dataOK || paramOK, "%s definition of %s not found", t.Pos, refName) {
			return false
		}
		if enumOK {
			t.Kind = model.TypeKindEnum
			t.Enum = enum
			t.SkelName = enum.SkelName
			return true
		}
		if dataOK {
			if !checkDataValueType(reporter, t, dataType) {
				return false
			}
			t.Kind = model.TypeKindData
			t.Data = dataType
			t.SkelName = dataType.SkelName
			valid := true
			for _, typeArg := range t.TypeArguments {
				valid = fixTypeRef(reporter, typeArg, refCtx) && valid
			}
			return checkTypeArguments(reporter, t, refName) && valid
		}
		t.Kind = model.TypeKindTypeParameter
		t.TypeParameter = param
		return true

	case model.TypeKindList:
		return fixTypeRef(reporter, t.List.Value, refCtx)

	case model.TypeKindMap:
		keyValid := fixTypeRef(reporter, t.Map.Key, refCtx)
		if keyValid && !(refCtx.allowUnresolvedImports && t.Map.Key.Kind == model.TypeKindUnresolvedReference) {
			keyValid = checkTypeCanBeMapKey(reporter, t.Map.Key)
		}
		return fixTypeRef(reporter, t.Map.Value, refCtx) && keyValid
	}
	return true
}

func checkTypeArguments(reporter *_DiagnosticReporter, t *model.Type, refName string) bool {
	referencePos := t.ReferencePos
	if referencePos == (model.Position{}) {
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

func checkDataValueType(reporter *_DiagnosticReporter, kind *model.Type, data *model.Data) bool {
	return reporter.check(data.Kind == model.DataKindData,
		"%s %s %s cannot be used as a value type", kind.Pos, data.Kind, data.Name)
}
