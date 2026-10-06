package schema

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDependenciesCoverAllDeclarationReferences(t *testing.T) {
	value := func(name string) *Type {
		return new(Type{Kind: TypeKindData, SkelName: "foreign." + name, Data: new(Data{Kind: DataKindData})})
	}
	members := func(name string) *Data { return new(Data{Members: []*DataMember{{Type: value(name)}}}) }
	arguments := func(name string) []*Argument { return []*Argument{{Type: value(name)}} }
	check := func(name string) *PermissionExpression {
		return new(PermissionExpression{Check: new(PermissionCheckInvocation{ResourceSkelName: "foreign." + name, Arguments: []*PermissionCheckArgument{{Type: value(name + "Argument")}}})})
	}
	container := value("Container")
	container.TypeArguments = []*Type{{Kind: TypeKindEnum, SkelName: "foreign.Status"}}
	event, config := value("Changed"), value("Settings")
	event.Data.Kind, config.Data.Kind = DataKindEvent, DataKindConfig
	doc := NewDomainFromSpec(DomainSpec{Name: "demo",
		Actors:  []*Actor{{SkelName: "demo.Caller", Auth: new(ActorAuth{Credential: members("Credential"), Info: members("AuthInfo")})}},
		Configs: []*Data{members("ConfigValue")}, Events: []*Data{members("EventValue")},
		Resources: []*Resource{{SkelName: "demo.Record", Checks: []*ResourceCheck{{Method: new(Method{Arguments: arguments("ResourceArgument")})}}, Actions: []*ResourceAction{{Checks: []*ResourceCheck{{Method: new(Method{Arguments: arguments("ActionArgument")})}}}}}},
		Services: []*Service{{SkelName: "demo.ReadService", Audiences: []*ActorAudience{{Actor: "foreign.ServiceCaller"}},
			Require: new(PermissionRequire{Expression: new(PermissionExpression{Mode: PermissionRequireModeAll, Children: []*PermissionExpression{check("ServiceCheck"), {Mode: PermissionRequireModeCode, Code: "foreign.CodeResource:read"}}})}),
			Methods: []*Method{{Arguments: arguments("MethodArgument"), ResultType: value("MethodResult"), Require: new(PermissionRequire{Expression: check("MethodCheck")})}},
		}},
		Tasks: []*Task{{SkelName: "demo.Job", Triggers: []*TaskTrigger{{Arguments: arguments("TaskArgument")}}}},
		Webs:  []*Web{{SkelName: "demo.Web", Audiences: []*ActorAudience{{Actor: "foreign.WebCaller"}}}},
		Data: []*Data{{SkelName: "demo.Value", Members: []*DataMember{
			{Type: new(Type{Kind: TypeKindData, SkelName: "demo.Value", Data: new(Data{Kind: DataKindData})})},
			{Type: event}, {Type: config},
			{Type: new(Type{Kind: TypeKindMap, Map: new(MapType{Key: new(Type{Kind: TypeKindEnum, SkelName: "foreign.MapKey"}), Value: new(Type{Kind: TypeKindList, List: new(ListType{Value: container})})})})},
			{Type: value("Container")},
		}}},
		Enums: []*Enum{{SkelName: "demo.Status"}},
	})
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
