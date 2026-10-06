package schema

import "testing"

func TestDeclarationViewsPreservePubModifier(t *testing.T) {
	for _, modifier := range []string{"", "pub", "ext"} {
		t.Run(modifier, func(t *testing.T) {
			pub, ext := modifier == "pub", modifier == "ext"
			service := new(Service{Name: "ExampleService", Pub: pub, Ext: ext})
			event := new(Data{Name: "ExampleEvent", Kind: DataKindEvent, Pub: pub, Ext: ext})
			domain := NewDomainFromSpec(DomainSpec{Name: "demo", Services: []*Service{service}, Events: []*Data{event}})
			for _, declaration := range domain.Declarations() {
				if declaration.Pub != pub {
					t.Fatalf("%s: Pub = %v, want %v", declaration.Name, declaration.Pub, pub)
				}
			}
			if service.Pub != pub || service.Ext != ext || event.Pub != pub || event.Ext != ext {
				t.Fatal("declaration view changed source modifiers")
			}
		})
	}
}

func TestDeclarationViewsBorrowSemanticNodes(t *testing.T) {
	data := new(Data{Name: "Value", SkelName: "demo.Value", Kind: DataKindData, Description: "Documented.", Pos: Position{File: "data.skel", Line: 3, Column: 1}})
	domain := NewDomainFromSpec(DomainSpec{Name: "demo", Data: []*Data{data}, Resources: []*Resource{{Name: "Value", SkelName: "demo.Value"}}})
	entries := domain.Declarations()
	if len(entries) != 2 || entries[0].Data != data || entries[0].Pos != data.Pos || entries[0].Description != data.Description {
		t.Fatalf("lookup copied or lost semantic data: %+v", entries)
	}
	if domain.Find(DeclarationTypeData, "demo.Value").Data != data {
		t.Fatal("find did not return the original semantic node")
	}
	if domain.Find(DeclarationTypeResource, "demo.Value").Resource == nil {
		t.Fatal("declaration namespaces collided")
	}
	if domain.Find(DeclarationTypeEnum, "demo.Value") != nil {
		t.Fatal("missing kind matched an unrelated declaration")
	}
}
