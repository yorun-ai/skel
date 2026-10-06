package analyzer

import (
	"slices"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
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

type _Refs map[*schema.Data]_RefKind

func (r _Refs) put(rk _RefKind, rs *schema.Data) {
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

func referencedData(t *schema.Type) _Refs {
	refs := _Refs{}
	switch t.Kind {
	case schema.TypeKindData:
		rk := refKindDirect
		if t.Nullable {
			rk = refKindNullable
		}
		refs.put(rk, t.Data)
		for _, arg := range t.TypeArguments {
			refs.override(rk, referencedData(arg))
		}
	case schema.TypeKindList:
		refs.override(refKindList, referencedData(t.List.Element))
	case schema.TypeKindMap:
		refs.override(refKindMap, referencedData(t.Map.Value))
	}
	return refs
}

func checkTypeCanBeMapKey(reporter *_DiagnosticReporter, t *schema.Type) bool {
	valid := reporter.checkNot(t.Nullable, "%s incorrect key type, must not be nullable", t.Pos)

	canBeMapKey := false
	if slices.Contains(mapKeyTypes, t.Kind) {
		if t.Kind != schema.TypeKindScalar {
			canBeMapKey = true
		} else if slices.Contains(mapKeyScalarTypes, t.Scalar) {
			canBeMapKey = true
		}
	}
	valid = reporter.check(canBeMapKey, "%s incorrect key type, int/string/uuid or Enum expected", t.Pos) && valid
	return valid
}

func parseType(reporter *_DiagnosticReporter, s *grammar.Type) (*schema.Type, bool) {
	if reporter.cancelled() {
		return nil, false
	}
	if s == nil {
		return nil, true
	}
	valid := true

	t := &schema.Type{
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
		t.Kind = schema.TypeKindScalar
		switch *s.Plain {
		case grammar.Int:
			t.Scalar = schema.ScalarInt
		case grammar.Float:
			t.Scalar = schema.ScalarFloat
		case grammar.Boolean:
			t.Scalar = schema.ScalarBoolean
		case grammar.String:
			t.Scalar = schema.ScalarString
		case grammar.Decimal:
			t.Scalar = schema.ScalarDecimal
		case grammar.Binary:
			t.Scalar = schema.ScalarBinary
		case grammar.Timestamp:
			t.Scalar = schema.ScalarTimestamp
		case grammar.Duration:
			t.Scalar = schema.ScalarDuration
		case grammar.LocalDate:
			t.Scalar = schema.ScalarLocalDate
		case grammar.LocalTime:
			t.Scalar = schema.ScalarLocalTime
		case grammar.LocalDateTime:
			t.Scalar = schema.ScalarLocalDateTime
		case grammar.UUID:
			t.Scalar = schema.ScalarUUID
		case grammar.JSON:
			t.Scalar = schema.ScalarJSON
		default:
			reporter.reportReferencef("%s unknown PlainType %s", s.Pos, *s.Plain)
			valid = false
		}

	case s.List != nil:
		valueType, valueValid := parseType(reporter, s.List.Value)
		valid = valueValid && valid
		t.Kind = schema.TypeKindList
		t.List = &schema.ListType{
			Element: valueType,
		}

	case s.Map != nil:
		keyType, keyValid := parseType(reporter, s.Map.Key)
		valueType, valueValid := parseType(reporter, s.Map.Value)
		valid = keyValid && valueValid && valid
		t.Kind = schema.TypeKindMap
		t.Map = &schema.MapType{
			Key:   keyType,
			Value: valueType,
		}

	case s.Reference != nil:
		refName, refQualifier, refPos, referenceValid := parseReferenceName(reporter, s.Reference.Name)
		valid = referenceValid && valid
		valid = checkCase(reporter, "Enum/Data/TypeParameter", caseTypeCamel, &grammar.Identifier{Value: refName, Pos: refPos}) && valid
		t.Kind = schema.TypeKindUnresolvedReference
		typeArgs := make([]*schema.Type, 0, len(s.Reference.TypeArguments))
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
