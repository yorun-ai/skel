package codegen

import (
	"errors"
	"reflect"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestWalkTypeVisitsStructuralChildrenInOrder(t *testing.T) {
	key := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}
	value := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarInt}
	root := &schema.Type{Kind: schema.TypeKindMap, Map: &schema.MapType{Key: key, Value: &schema.Type{
		Kind: schema.TypeKindList, List: &schema.ListType{Value: value},
	}}}
	kinds := []schema.TypeKind{}
	if err := WalkType(root, func(type_ *schema.Type) error {
		kinds = append(kinds, type_.Kind)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	expected := []schema.TypeKind{schema.TypeKindMap, schema.TypeKindScalar, schema.TypeKindList, schema.TypeKindScalar}
	if !reflect.DeepEqual(kinds, expected) {
		t.Fatalf("unexpected walk order: %v", kinds)
	}
}

func TestWalkTypeGraphTerminatesOnRecursiveData(t *testing.T) {
	data := &schema.Data{Name: "Node", Kind: schema.DataKindData}
	reference := &schema.Type{Kind: schema.TypeKindData, Data: data}
	data.Members = []*schema.DataMember{{Name: "next", Type: reference}}
	visits := 0
	if err := WalkTypeGraph(reference, func(*schema.Type) error {
		visits++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if visits != 1 {
		t.Fatalf("expected recursive type to be visited once, got %d", visits)
	}
}

func TestWalkTypePropagatesVisitorError(t *testing.T) {
	expected := errors.New("stop")
	err := WalkType(&schema.Type{Kind: schema.TypeKindScalar}, func(*schema.Type) error { return expected })
	if !errors.Is(err, expected) {
		t.Fatalf("expected visitor error, got %v", err)
	}
}

func TestVisitTypesSharesTraversalStateAcrossRoots(t *testing.T) {
	shared := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}
	roots := []*schema.Type{
		{Kind: schema.TypeKindList, List: &schema.ListType{Value: shared}},
		{Kind: schema.TypeKindMap, Map: &schema.MapType{Key: shared, Value: shared}},
	}
	visits := map[*schema.Type]int{}
	VisitTypes(roots, func(kind *schema.Type) {
		visits[kind]++
	})
	if visits[shared] != 1 {
		t.Fatalf("expected shared type to be visited once, got %d", visits[shared])
	}
}

func TestVisitTypeGraphsSharesReferencedDataAcrossRoots(t *testing.T) {
	data := &schema.Data{Name: "Shared", Kind: schema.DataKindData}
	reference := &schema.Type{Kind: schema.TypeKindData, Data: data}
	data.Members = []*schema.DataMember{{Name: "value", Type: &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}}}
	visits := map[*schema.Type]int{}
	VisitTypeGraphs([]*schema.Type{reference, reference}, func(kind *schema.Type) {
		visits[kind]++
	})
	if visits[reference] != 1 || len(visits) != 2 {
		t.Fatalf("expected one shared graph traversal, got %+v", visits)
	}
}
