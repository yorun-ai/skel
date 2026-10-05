package schema

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDependenciesCoverAllDeclarationReferences(t *testing.T) {
	value := &Type{Kind: TypeKindData, Name: "foreign.Value", Arguments: []*Type{{Kind: TypeKindEnum, Name: "foreign.Status"}}}
	arguments := []*Argument{{Type: &Type{Kind: TypeKindMap, Key: &Type{Kind: TypeKindScalar, Name: "string"}, Value: &Type{Kind: TypeKindList, Element: value}}}}
	members := &DataSchema{Members: []*Member{{Type: value}}}
	check := &Requirement{Check: &RequirementCheck{Resource: "foreign.Record", Arguments: []*RequirementCheckArgument{{Type: value}}}}
	audience := []*Audience{{Actor: "foreign.Caller"}}
	doc := &Document{Domain: "demo", Declarations: []*Declaration{
		{Kind: DeclarationTypeActor, SkelName: "demo.Caller", Actor: &ActorSchema{AuthCredential: members, AuthInfo: members}},
		{Kind: DeclarationTypeConfig, SkelName: "demo.Settings", Data: members},
		{Kind: DeclarationTypeEvent, SkelName: "demo.Changed", Data: members},
		{Kind: DeclarationTypeResource, SkelName: "demo.Record", Resource: &ResourceSchema{Checks: []*ResourceCheck{{Arguments: arguments}}, Actions: []*ResourceAction{{Checks: []*ResourceCheck{{Arguments: arguments}}}}}},
		{Kind: DeclarationTypeService, SkelName: "demo.ReadService", Service: &ServiceSchema{Audiences: audience, Require: &Requirement{Mode: RequirementModeAll, Children: []*Requirement{check, {Mode: RequirementModeCode, Code: "foreign.Record:read"}}}, Methods: []*Method{{Arguments: arguments, Result: value, Require: check}}}},
		{Kind: DeclarationTypeTask, SkelName: "demo.Job", Task: &TaskSchema{Triggers: []*Trigger{{Arguments: arguments}}}},
		{Kind: DeclarationTypeWeb, SkelName: "demo.Web", Web: &WebSchema{Audiences: audience}},
		{Kind: DeclarationTypeData, SkelName: "demo.Value", Data: &DataSchema{Members: []*Member{{Type: &Type{Kind: TypeKindData, Name: "demo.Value"}}, {Type: &Type{Kind: TypeKindEvent, Name: "foreign.Changed"}}, {Type: &Type{Kind: TypeKindConfig, Name: "foreign.Settings"}}}}},
		{Kind: DeclarationTypeEnum, SkelName: "demo.Status"},
	}}
	report := Dependencies(doc)
	want := []Dependency{{Domain: "foreign", Name: "Caller", Kind: "actor"}, {Domain: "foreign", Name: "Changed", Kind: "event"}, {Domain: "foreign", Name: "Record", Kind: "resource"}, {Domain: "foreign", Name: "Settings", Kind: "config"}, {Domain: "foreign", Name: "Status", Kind: "enum"}, {Domain: "foreign", Name: "Value", Kind: "data"}}
	if !reflect.DeepEqual(report.Dependencies, want) {
		t.Fatalf("dependencies=%+v", report.Dependencies)
	}
	for _, names := range [][]string{report.Actors, report.Configs, report.Events, report.Resources, report.Services, report.Tasks, report.Webs, report.Data, report.Enums} {
		if len(names) != 1 {
			t.Fatalf("missing declaration group: %+v", report)
		}
	}
	encoded, err := json.Marshal(report)
	if err != nil || strings.Contains(string(encoded), "null") {
		t.Fatalf("invalid result %s: %v", encoded, err)
	}
}
