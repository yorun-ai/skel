package typescript

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
)

func TestCastTypeMapsBinaryToUint8Array(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind:   model.TypeKindScalar,
		Scalar: model.ScalarBinary,
	})
	if got.Plain != "Uint8Array" {
		t.Fatalf("unexpected binary type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsUUIDToString(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind:   model.TypeKindScalar,
		Scalar: model.ScalarUUID,
	})
	if got.Plain != "string" {
		t.Fatalf("unexpected uuid type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsJSONToString(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind:   model.TypeKindScalar,
		Scalar: model.ScalarJSON,
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
	kind := &model.Type{
		Kind: model.TypeKindData,
		Data: &model.Data{
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
	kind := &model.Type{
		Kind: model.TypeKindEnum,
		Enum: &model.Enum{
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
	box := &model.Data{Name: "Box", TypeParameters: []*model.TypeParameter{parameter}}
	for _, test := range []struct {
		name string
		kind *model.Type
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
