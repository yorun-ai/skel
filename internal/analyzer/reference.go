package analyzer

import (
	"slices"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/nameutil"
)

type _RefKind int

const (
	refKindDirect _RefKind = iota
	refKindNullable
	refKindList
	refKindMap
)

func (rk _RefKind) isHard() bool {
	return rk == refKindDirect
}

type _Refs map[*model.Data]_RefKind

func (r _Refs) put(rk _RefKind, rs *model.Data) {
	if prevKind, exists := r[rs]; exists && prevKind <= rk {
		return
	}
	r[rs] = rk
}

func (r _Refs) merge(other _Refs) {
	for refData, refKind := range other {
		r.put(refKind, refData)
	}
}

func (r _Refs) override(refKind _RefKind, other _Refs) {
	for refData := range other {
		r.put(refKind, refData)
	}
}

func referencedData(t *model.Type) _Refs {
	refs := _Refs{}
	switch t.Kind {
	case model.TypeKindData:
		rk := refKindDirect
		if t.Nullable {
			rk = refKindNullable
		}
		refs.put(rk, t.Data)
		for _, arg := range t.TypeArguments {
			refs.override(rk, referencedData(arg))
		}
	case model.TypeKindList:
		refs.override(refKindList, referencedData(t.List.Value))
	case model.TypeKindMap:
		refs.override(refKindMap, referencedData(t.Map.Value))
	}
	return refs
}

func checkTypeCanBeMapKey(reporter *_DiagnosticReporter, t *model.Type) bool {
	valid := reporter.checkNot(t.Nullable, "%s incorrect key type, must not be nullable", t.Pos)

	canBeMapKey := false
	if slices.Contains(mapKeyTypes, t.Kind) {
		if t.Kind != model.TypeKindScalar {
			canBeMapKey = true
		} else if slices.Contains(mapKeyScalarTypes, t.Scalar) {
			canBeMapKey = true
		}
	}
	valid = reporter.check(canBeMapKey, "%s incorrect key type, int/string/uuid or Enum expected", t.Pos) && valid
	return valid
}

func parseType(reporter *_DiagnosticReporter, s *grammar.Type) (*model.Type, bool) {
	if reporter.cancelled() {
		return nil, false
	}
	if s == nil {
		return nil, true
	}
	valid := true

	t := &model.Type{
		Pos:           position(s.Pos),
		Kind:          typeKindNone,
		Scalar:        scalarNone,
		List:          nil,
		Map:           nil,
		Data:          nil,
		TypeParameter: nil,
		Nullable:      false,
	}

	switch {
	case s.Plain != nil:
		t.Kind = model.TypeKindScalar
		switch *s.Plain {
		case grammar.Int:
			t.Scalar = model.ScalarInt
		case grammar.Float:
			t.Scalar = model.ScalarFloat
		case grammar.Boolean:
			t.Scalar = model.ScalarBoolean
		case grammar.String:
			t.Scalar = model.ScalarString
		case grammar.Decimal:
			t.Scalar = model.ScalarDecimal
		case grammar.Binary:
			t.Scalar = model.ScalarBinary
		case grammar.Timestamp:
			t.Scalar = model.ScalarTimestamp
		case grammar.Duration:
			t.Scalar = model.ScalarDuration
		case grammar.LocalDate:
			t.Scalar = model.ScalarLocalDate
		case grammar.LocalTime:
			t.Scalar = model.ScalarLocalTime
		case grammar.LocalDateTime:
			t.Scalar = model.ScalarLocalDateTime
		case grammar.UUID:
			t.Scalar = model.ScalarUUID
		case grammar.JSON:
			t.Scalar = model.ScalarJSON
		default:
			reporter.reportReferencef("%s unknown PlainType %s", s.Pos, *s.Plain)
			valid = false
		}

	case s.List != nil:
		valueType, valueValid := parseType(reporter, s.List.Value)
		valid = valueValid && valid
		t.Kind = model.TypeKindList
		t.List = &model.ListType{
			Value: valueType,
		}

	case s.Map != nil:
		keyType, keyValid := parseType(reporter, s.Map.Key)
		valueType, valueValid := parseType(reporter, s.Map.Value)
		valid = keyValid && valueValid && valid
		t.Kind = model.TypeKindMap
		t.Map = &model.MapType{
			Key:   keyType,
			Value: valueType,
		}

	case s.Reference != nil:
		refName, refQualifier, refPos, referenceValid := parseReferenceName(reporter, s.Reference.Name)
		valid = referenceValid && valid
		valid = checkCase(reporter, "Enum/Data/TypeParameter", caseTypeCamel, &grammar.Identifier{Value: refName, Pos: refPos}) && valid
		t.Kind = model.TypeKindUnresolvedReference
		typeArgs := make([]*model.Type, 0, len(s.Reference.TypeArguments))
		for _, typeArg := range s.Reference.TypeArguments {
			if reporter.cancelled() {
				break
			}
			parsedTypeArg, argumentValid := parseType(reporter, typeArg)
			valid = argumentValid && valid
			typeArgs = append(typeArgs, parsedTypeArg)
		}
		t.ReferencePos = position(refPos)
		t.SkelName = refName
		t.ExternalAlias = refQualifier
		t.TypeArguments = typeArgs

	default:
		reporter.reportReferencef("%s unknown Type %+v", s.Pos, s)
		valid = false
	}

	t.Nullable = s.Nullable

	return t, valid
}

func parseReferenceName(reporter *_DiagnosticReporter, name *grammar.QualifiedName) (string, string, lexer.Position, bool) {
	if !reporter.check(name != nil && len(name.Parts) > 0, "missing reference type") {
		return "", "", lexer.Position{}, false
	}
	last := name.Parts[len(name.Parts)-1]
	qualifier, _, _ := nameutil.SplitQualified(name.String())
	return last.Value, qualifier, last.Pos, true
}
