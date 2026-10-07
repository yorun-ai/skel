package schema

import (
	"fmt"
	"testing"
)

func ExampleType_ContainsBinaryType() {
	attachment := &Data{
		Name: "Attachment",
		Members: []*DataMember{
			{Name: "content", Type: &Type{Kind: TypeKindScalar, Scalar: ScalarBinary}},
		},
	}
	attachmentType := &Type{Kind: TypeKindData, Data: attachment}

	fmt.Println(attachmentType.Name())
	fmt.Println(attachmentType.ContainsBinaryType())

	// Output:
	// Attachment
	// true
}

func TestTypeContainsBinaryType(t *testing.T) {
	binaryData := &Data{
		Name: "Attachment",
		Members: []*DataMember{
			{Name: "content", Type: &Type{Kind: TypeKindScalar, Scalar: ScalarBinary}},
		},
	}
	valueType := &Type{Kind: TypeKindData, Data: binaryData}

	if !valueType.ContainsBinaryType() {
		t.Fatal("expected data containing binary member to report binary type")
	}
	if valueType.Name() != "Attachment" {
		t.Fatalf("unexpected type name: %q", valueType.Name())
	}
}

func TestContainsBinaryTypeGenericBindings(t *testing.T) {
	parameter := &TypeParameter{Name: "T"}
	box := &Data{Name: "Box", TypeParameters: []*TypeParameter{parameter}}
	box.Members = []*DataMember{{Name: "value", Type: &Type{Kind: TypeKindTypeParameter, TypeParameter: parameter}}}
	binary := &Type{Kind: TypeKindScalar, Scalar: ScalarBinary}
	text := &Type{Kind: TypeKindScalar, Scalar: ScalarString}
	instance := func(arg *Type) *Type { return &Type{Kind: TypeKindData, Data: box, TypeArguments: []*Type{arg}} }
	for _, test := range []struct {
		name string
		kind *Type
		want bool
	}{
		{"binary", instance(binary), true},
		{"string", instance(text), false},
		{"nested binary", instance(instance(binary)), true},
		{"nested string", instance(instance(text)), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.kind.ContainsBinaryType(); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
	box.Members = append([]*DataMember{{Name: "next", Type: instance(&Type{Kind: TypeKindTypeParameter, TypeParameter: parameter})}}, box.Members...)
	if !instance(binary).ContainsBinaryType() || instance(text).ContainsBinaryType() {
		t.Fatal("incorrect recursive generic classification")
	}
	box.Members = nil
	if instance(binary).ContainsBinaryType() {
		t.Fatal("unused generic parameter is not on the wire")
	}
}
