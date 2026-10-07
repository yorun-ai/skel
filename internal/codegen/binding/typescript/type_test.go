package typescript

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestCastTypeMapsBinaryToUint8Array(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{
		Kind:   schema.TypeKindScalar,
		Scalar: schema.ScalarBinary,
	})
	if got.Plain != "Uint8Array" {
		t.Fatalf("unexpected binary type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsUUIDToString(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{
		Kind:   schema.TypeKindScalar,
		Scalar: schema.ScalarUUID,
	})
	if got.Plain != "string" {
		t.Fatalf("unexpected uuid type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsJSONToString(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{
		Kind:   schema.TypeKindScalar,
		Scalar: schema.ScalarJSON,
	})
	if got.Plain != "string" {
		t.Fatalf("unexpected json type mapping: %s", got.Plain)
	}
}

func TestCastMapTypeMapsUUIDKeyToString(t *testing.T) {
	got := (_Types{}).castType(codegentest.MapType(codegentest.UUIDType(), codegentest.StringType()))
	if got.Plain != "Record<string, string>" {
		t.Fatalf("unexpected uuid-keyed map type: %s", got.Plain)
	}
}

func TestCastTypeQualifiesExternalDataWithAlias(t *testing.T) {
	kind := &schema.Type{
		Kind: schema.TypeKindData,
		Data: &schema.Data{
			Name: "UserSummary",
		},
		ExternalAlias: "userpub",
	}
	got := (_Types{bindings: binding.TypeBindings{kind: {Alias: "userpub", Path: "@example/user"}}}).castType(kind)
	if got.Plain != "userpub.UserSummary" {
		t.Fatalf("unexpected external type: %s", got.Plain)
	}
}

func TestCastTypeQualifiesExternalEnumWithAlias(t *testing.T) {
	kind := &schema.Type{
		Kind: schema.TypeKindEnum,
		Enum: &schema.Enum{
			Name: "UserStatus",
		},
		ExternalAlias: "userpub",
	}
	got := (_Types{bindings: binding.TypeBindings{kind: {Alias: "userpub", Path: "@example/user"}}}).castType(kind)
	if got.Plain != "userpub.UserStatus" {
		t.Fatalf("unexpected external type: %s", got.Plain)
	}
}

func TestCastNullableTypeParameterInCollectionsAndArguments(t *testing.T) {
	parameter := codegentest.TypeParam("TValue")
	plain := codegentest.TypeParamType(parameter)
	nullable := codegentest.NullableType(codegentest.TypeParamType(parameter))
	box := &schema.Data{Name: "Box", TypeParameters: []*schema.TypeParameter{parameter}}
	for _, test := range []struct {
		name string
		kind *schema.Type
		want string
	}{
		{name: "parameter", kind: plain, want: "TValue"},
		{name: "nullable parameter", kind: nullable, want: "TValue | null"},
		{name: "list elements", kind: codegentest.ListType(nullable), want: "Array<TValue | null>"},
		{name: "map values", kind: codegentest.MapType(codegentest.StringType(), nullable), want: "Record<string, TValue | null>"},
		{name: "nested generic", kind: codegentest.DataType(box, nullable), want: "Box<TValue | null>"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := (_Types{}).castType(test.kind)
			if got.Plain != test.want {
				t.Fatalf("type: got %q want %q", got.Plain, test.want)
			}
		})
	}
}
