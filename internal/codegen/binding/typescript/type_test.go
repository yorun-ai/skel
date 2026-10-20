package typescript

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestCastTypeMapsJSONToString(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{
		Kind:   schema.TypeKindScalar,
		Scalar: schema.ScalarJSON,
	})
	if got.Plain != "string" {
		t.Fatalf("unexpected json type mapping: %s", got.Plain)
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
