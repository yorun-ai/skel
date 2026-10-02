package schema

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeRejectsInvalidSchemaDocuments(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"unsupported format", `{"format":"other.schema","formatVersion":1,"domain":"demo","declarations":[]}`},
		{"unsupported format version", `{"format":"yorun.skel.schema","formatVersion":2,"domain":"demo","declarations":[]}`},
		{"missing declarations", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo"}`},
		{"unknown declaration type", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"User","type":"unknown","skelName":"demo.User","data":{"members":[]}}]}`},
		{"unknown config lifecycle", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"Runtime","type":"config","skelName":"demo.Runtime","data":{"lifecycle":"session","members":[]}}]}`},
		{"unknown auth mode", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"Users","type":"service","skelName":"demo.Users","service":{"audiences":[],"auth":"sometimes","methods":[]}}]}`},
		{"unknown type kind", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"User","type":"data","skelName":"demo.User","data":{"members":[{"name":"id","type":{"kind":"mystery"}}]}}]}`},
		{"unknown requirement mode", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"Users","type":"service","skelName":"demo.Users","service":{"audiences":[],"auth":"unset","require":{"mode":"maybe"},"methods":[]}}]}`},
		{"incomplete required collection", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"User","type":"data","skelName":"demo.User","data":{}}]}`},
		{"null member type", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"User","type":"data","skelName":"demo.User","data":{"members":[{"name":"id","type":null}]}}]}`},
		{"unrelated type fields", `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"pub":true,"name":"User","type":"data","skelName":"demo.User","data":{"members":[{"name":"id","type":{"kind":"scalar","name":"string","element":{"kind":"scalar","name":"string"}}}]}}]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Decode(strings.NewReader(test.json)); err == nil {
				t.Fatal("expected schema validation to fail")
			}
		})
	}
}

func TestValidateRejectsCyclicTypesAndRequirements(t *testing.T) {
	kind := new(Type{Kind: TypeKindList})
	kind.Element = kind
	requirement := new(Requirement{Mode: RequirementModeAll})
	requirement.Children = []*Requirement{requirement}
	documents := []*Document{
		newTestDocument(&Declaration{Name: "Value", SkelName: "demo.Value", Kind: DeclarationTypeData, Data: new(DataSchema{Members: []*Member{{Name: "value", Type: kind}}})}),
		newTestDocument(&Declaration{Name: "Service", SkelName: "demo.Service", Kind: DeclarationTypeService, Service: new(ServiceSchema{Auth: AuthModeUnset, Audiences: []*Audience{}, Methods: []*Method{}, Require: requirement})}),
	}
	for _, document := range documents {
		var output bytes.Buffer
		if err := Encode(&output, document); err == nil || output.Len() != 0 {
			t.Fatalf("expected cycle to be rejected before encoding, got %v", err)
		}
		if err := Validate(document); err == nil || !strings.Contains(err.Error(), "cyclic") {
			t.Fatalf("expected cycle error, got %v", err)
		}
	}
}

func TestDecodeRejectsNullRequirementChildren(t *testing.T) {
	input := `{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"name":"Service","type":"service","skelName":"demo.Service","service":{"audiences":[],"auth":"unset","methods":[],"require":{"mode":"all","children":[null]}}}]}`
	if _, err := Decode(strings.NewReader(input)); err == nil {
		t.Fatal("expected null requirement child error")
	}
}

func TestValidateAllowsSharedTypesAndRequirements(t *testing.T) {
	scalar := new(Type{Kind: TypeKindScalar, Name: "string"})
	requirement := new(Requirement{Mode: RequirementModeCode, Code: "demo.File:read"})
	document := newTestDocument(
		&Declaration{Name: "Value", SkelName: "demo.Value", Kind: DeclarationTypeData, Data: new(DataSchema{Members: []*Member{{Name: "value", Type: new(Type{Kind: TypeKindMap, Key: scalar, Value: scalar})}}})},
		&Declaration{Name: "Service", SkelName: "demo.Service", Kind: DeclarationTypeService, Service: new(ServiceSchema{Auth: AuthModeUnset, Audiences: []*Audience{}, Methods: []*Method{}, Require: new(Requirement{Mode: RequirementModeAll, Children: []*Requirement{requirement, requirement}})})},
	)
	if err := Validate(document); err != nil {
		t.Fatal(err)
	}
}
