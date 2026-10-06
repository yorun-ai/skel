package codegen

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestInstantiateMembersCopiesArgumentsAndPreservesRecursiveIdentity(t *testing.T) {
	parameter := new(schema.TypeParameter{Name: "TItem"})
	parameterRef := new(schema.Type{Kind: schema.TypeKindTypeParameter, TypeParameter: parameter})
	data := new(schema.Data{Name: "Node", Kind: schema.DataKindData, TypeParameters: []*schema.TypeParameter{parameter}})
	data.Members = []*schema.DataMember{
		{Name: "value", Type: parameterRef},
		{Name: "next", Type: new(schema.Type{Kind: schema.TypeKindData, Data: data, Nullable: true, TypeArguments: []*schema.Type{parameterRef}})},
	}
	arg := new(schema.Type{Kind: schema.TypeKindList, List: new(schema.ListType{Element: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString})})})
	input := new(schema.Type{Kind: schema.TypeKindData, Data: data, TypeArguments: []*schema.Type{arg}})
	members, err := InstantiateMembers(input)
	if err != nil {
		t.Fatal(err)
	}
	if members[1].Type.Data != data || members[1].Type.TypeArguments[0].Kind != schema.TypeKindList || !members[1].Type.Nullable {
		t.Fatal("recursive instantiation lost semantic identity")
	}
	members[0].Type.List.Element.Scalar = schema.ScalarInt
	if arg.List.Element.Scalar != schema.ScalarString || data.Members[0].Type.Kind != schema.TypeKindTypeParameter {
		t.Fatal("returned expressions alias input")
	}
	// Open recursive Node<TItem> must preserve TItem from the caller's scope.
	open, err := InstantiateMembers(data.Members[1].Type)
	if err != nil {
		t.Fatal(err)
	}
	if open[0].Type.TypeParameter != parameter {
		t.Fatal("open type argument lost")
	}
	input.TypeArguments = nil
	if _, err := InstantiateMembers(input); err == nil {
		t.Fatal("wrong generic arity accepted")
	}
}
