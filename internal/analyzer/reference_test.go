package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

func TestParseTypeAndFixRef(t *testing.T) {
	page := &schema.Data{
		Name: "Page",
		Kind: schema.DataKindData,
		TypeParameters: []*schema.TypeParameter{
			{Name: "TItem"},
		},
	}
	user := &schema.Data{Name: "User", Kind: schema.DataKindData}

	tp := parseTypeTest(t, refGrammarType("Page", refGrammarType("User")))
	fixTypeRefTest(t, tp, &_RefContext{
		dataList: map[string]*schema.Data{
			"Page": page,
			"User": user,
		},
		typeParameters: map[string]*schema.TypeParameter{
			"TItem": page.TypeParameters[0],
		},
	})

	if tp.Kind != schema.TypeKindData {
		t.Fatalf("unexpected type kind: %v", tp.Kind)
	}
	if tp.Data != page {
		t.Fatalf("unexpected data: %+v", tp.Data)
	}
	if len(tp.TypeArguments) != 1 || tp.TypeArguments[0].Data != user {
		t.Fatalf("unexpected type arguments: %+v", tp.TypeArguments)
	}
	if tp.Name() != "PageOfUser" {
		t.Fatalf("unexpected type name: %s", tp.Name())
	}
}

func TestTypeRefData(t *testing.T) {
	user := &schema.Data{Name: "User", Kind: schema.DataKindData}
	page := &schema.Type{
		Kind: schema.TypeKindData,
		Data: &schema.Data{Name: "Page"},
		TypeArguments: []*schema.Type{
			{Kind: schema.TypeKindData, Data: user},
		},
	}
	refs := referencedData(page)

	if refs[page.Data] != refKindDirect {
		t.Fatalf("unexpected direct ref kind: %v", refs[page.Data])
	}
	if refs[user] != refKindDirect {
		t.Fatalf("unexpected nested ref kind: %v", refs[user])
	}
}

func TestParseTypeMapAndNullable(t *testing.T) {
	typ := parseTypeTest(t, nullableType(mapType(plainType(grammar.String), refGrammarType("User"))))
	if typ.Kind != schema.TypeKindMap {
		t.Fatalf("unexpected type kind: %v", typ.Kind)
	}
	if !typ.Nullable {
		t.Fatal("expected nullable map type")
	}
	if typ.Map.Key.Kind != schema.TypeKindScalar || typ.Map.Key.Scalar != schema.ScalarString {
		t.Fatalf("unexpected map key: %+v", typ.Map.Key)
	}
	if typ.Map.Value.Kind != schema.TypeKindUnresolvedReference || typ.Map.Value.SkelName != "User" {
		t.Fatalf("unexpected map value: %+v", typ.Map.Value)
	}
}

func TestFixRefReturnsErrorWhenDefinitionMissing(t *testing.T) {
	typ := parseTypeTest(t, refGrammarType("User"))

	expectFixTypeRefDiagnostic(t, "definition of User not found", typ, &_RefContext{})
}

func TestFixRefReturnsErrorWhenGenericTypeArgsMismatch(t *testing.T) {
	page := &schema.Data{
		Name: "Page",
		Kind: schema.DataKindData,
		TypeParameters: []*schema.TypeParameter{
			{Name: "TItem"},
		},
	}

	typ := parseTypeTest(t, refGrammarType("Page", refGrammarType("User"), refGrammarType("Profile")))

	expectFixTypeRefDiagnostic(t, "mismatched type arguments", typ, &_RefContext{
		dataList: map[string]*schema.Data{
			"Page":    page,
			"User":    {Name: "User", Kind: schema.DataKindData},
			"Profile": {Name: "Profile", Kind: schema.DataKindData},
		},
	})
}

func TestFixRefReturnsErrorWhenGenericTypeArgsMissing(t *testing.T) {
	page := &schema.Data{
		Name: "Page",
		Kind: schema.DataKindData,
		TypeParameters: []*schema.TypeParameter{
			{Name: "TItem"},
		},
	}

	typ := parseTypeTest(t, refGrammarType("Page"))

	expectFixTypeRefDiagnostic(t, "need type argument", typ, &_RefContext{
		dataList: map[string]*schema.Data{
			"Page": page,
		},
	})
}

func TestFixRefReturnsErrorWhenMapKeyIsNullable(t *testing.T) {
	typ := parseTypeTest(t, mapType(nullableType(plainType(grammar.String)), plainType(grammar.Int)))

	expectFixTypeRefDiagnostic(t, "incorrect key type, must not be nullable", typ, &_RefContext{})
}

func TestFixRefAllowsUUIDMapKey(t *testing.T) {
	typ := parseTypeTest(t, mapType(plainType(grammar.UUID), plainType(grammar.Int)))

	fixTypeRefTest(t, typ, &_RefContext{})
	if typ.Map.Key.Kind != schema.TypeKindScalar || typ.Map.Key.Scalar != schema.ScalarUUID {
		t.Fatalf("unexpected map key: %+v", typ.Map.Key)
	}
}

func TestFixRefReturnsErrorWhenMapKeyTypeIsUnsupported(t *testing.T) {
	typ := parseTypeTest(t, mapType(plainType(grammar.Float), plainType(grammar.Int)))

	expectFixTypeRefDiagnostic(t, "int/string/uuid or Enum expected", typ, &_RefContext{})
}
