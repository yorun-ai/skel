package schema

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDependenciesCoverAllDeclarationReferences(t *testing.T) {
	// Each traversal path uses a distinct foreign name so another path cannot
	// hide a missing dependency.
	value := func(name string) *Type { return &Type{Kind: TypeKindData, Name: "foreign." + name} }
	members := func(name string) *DataSchema { return &DataSchema{Members: []*Member{{Type: value(name)}}} }
	arguments := func(name string) []*Argument { return []*Argument{{Type: value(name)}} }
	check := func(name string) *Requirement {
		return &Requirement{Check: &RequirementCheck{Resource: "foreign." + name, Arguments: []*RequirementCheckArgument{{Type: value(name + "Argument")}}}}
	}
	doc := &Document{Domain: "demo", Declarations: []*Declaration{
		{Kind: DeclarationTypeActor, SkelName: "demo.Caller", Actor: &ActorSchema{AuthCredential: members("Credential"), AuthInfo: members("AuthInfo")}},
		{Kind: DeclarationTypeConfig, SkelName: "demo.Settings", Data: members("ConfigValue")},
		{Kind: DeclarationTypeEvent, SkelName: "demo.Changed", Data: members("EventValue")},
		{Kind: DeclarationTypeResource, SkelName: "demo.Record", Resource: &ResourceSchema{
			Checks:  []*ResourceCheck{{Arguments: arguments("ResourceArgument")}},
			Actions: []*ResourceAction{{Checks: []*ResourceCheck{{Arguments: arguments("ActionArgument")}}}},
		}},
		{Kind: DeclarationTypeService, SkelName: "demo.ReadService", Service: &ServiceSchema{
			Audiences: []*Audience{{Actor: "foreign.ServiceCaller"}},
			Require:   &Requirement{Mode: RequirementModeAll, Children: []*Requirement{check("ServiceCheck"), {Mode: RequirementModeCode, Code: "foreign.CodeResource:read"}}},
			Methods:   []*Method{{Arguments: arguments("MethodArgument"), Result: value("MethodResult"), Require: check("MethodCheck")}},
		}},
		{Kind: DeclarationTypeTask, SkelName: "demo.Job", Task: &TaskSchema{Triggers: []*Trigger{{Arguments: arguments("TaskArgument")}}}},
		{Kind: DeclarationTypeWeb, SkelName: "demo.Web", Web: &WebSchema{Audiences: []*Audience{{Actor: "foreign.WebCaller"}}}},
		{Kind: DeclarationTypeData, SkelName: "demo.Value", Data: &DataSchema{Members: []*Member{
			{Type: &Type{Kind: TypeKindData, Name: "demo.Value"}},
			{Type: &Type{Kind: TypeKindEvent, Name: "foreign.Changed"}},
			{Type: &Type{Kind: TypeKindConfig, Name: "foreign.Settings"}},
			{Type: &Type{Kind: TypeKindMap, Key: &Type{Kind: TypeKindEnum, Name: "foreign.MapKey"}, Value: &Type{Kind: TypeKindList, Element: &Type{Kind: TypeKindData, Name: "foreign.Container", Arguments: []*Type{{Kind: TypeKindEnum, Name: "foreign.Status"}}}}}},
			{Type: value("Container")}, // Repeated references must be deduplicated.
		}}},
		{Kind: DeclarationTypeEnum, SkelName: "demo.Status"},
	}}
	report := Dependencies(doc)
	want := []Dependency{
		{Domain: "foreign", Name: "ActionArgument", Kind: "data"},
		{Domain: "foreign", Name: "AuthInfo", Kind: "data"},
		{Domain: "foreign", Name: "Changed", Kind: "event"},
		{Domain: "foreign", Name: "CodeResource", Kind: "resource"},
		{Domain: "foreign", Name: "ConfigValue", Kind: "data"},
		{Domain: "foreign", Name: "Container", Kind: "data"},
		{Domain: "foreign", Name: "Credential", Kind: "data"},
		{Domain: "foreign", Name: "EventValue", Kind: "data"},
		{Domain: "foreign", Name: "MapKey", Kind: "enum"},
		{Domain: "foreign", Name: "MethodArgument", Kind: "data"},
		{Domain: "foreign", Name: "MethodCheck", Kind: "resource"},
		{Domain: "foreign", Name: "MethodCheckArgument", Kind: "data"},
		{Domain: "foreign", Name: "MethodResult", Kind: "data"},
		{Domain: "foreign", Name: "ResourceArgument", Kind: "data"},
		{Domain: "foreign", Name: "ServiceCaller", Kind: "actor"},
		{Domain: "foreign", Name: "ServiceCheck", Kind: "resource"},
		{Domain: "foreign", Name: "ServiceCheckArgument", Kind: "data"},
		{Domain: "foreign", Name: "Settings", Kind: "config"},
		{Domain: "foreign", Name: "Status", Kind: "enum"},
		{Domain: "foreign", Name: "TaskArgument", Kind: "data"},
		{Domain: "foreign", Name: "WebCaller", Kind: "actor"},
	}
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
