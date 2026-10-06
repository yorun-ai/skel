package codegen

import (
	"testing"

	"go.yorun.ai/skel/internal/model"
)

func TestInstantiateMembersCopiesArgumentsAndPreservesRecursiveIdentity(t *testing.T) {
	parameter := new(model.TypeParameter{Name: "TItem"})
	parameterRef := new(model.Type{Kind: model.TypeKindTypeParameter, TypeParameter: parameter})
	data := new(model.Data{Name: "Node", Kind: model.DataKindData, TypeParameters: []*model.TypeParameter{parameter}})
	data.Members = []*model.DataMember{
		{Name: "value", Type: parameterRef},
		{Name: "next", Type: new(model.Type{Kind: model.TypeKindData, Data: data, Nullable: true, TypeArguments: []*model.Type{parameterRef}})},
	}
	arg := new(model.Type{Kind: model.TypeKindList, List: new(model.ListType{Value: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString})})})
	input := new(model.Type{Kind: model.TypeKindData, Data: data, TypeArguments: []*model.Type{arg}})
	members, err := InstantiateMembers(input)
	if err != nil {
		t.Fatal(err)
	}
	if members[1].Type.Data != data || members[1].Type.TypeArguments[0].Kind != model.TypeKindList || !members[1].Type.Nullable {
		t.Fatal("recursive instantiation lost semantic identity")
	}
	members[0].Type.List.Value.Scalar = model.ScalarInt
	if arg.List.Value.Scalar != model.ScalarString || data.Members[0].Type.Kind != model.TypeKindTypeParameter {
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
