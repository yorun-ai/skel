package descriptor_test

import (
	"encoding/json"
	"testing"

	"go.yorun.ai/skel/descriptor"
)

func TestResourceChecksReferenceServiceAfterJSONRoundTrip(t *testing.T) {
	resource := new(descriptor.Resource{
		Checks: []*descriptor.ResourceCheck{{Name: "byId", MethodName: "checkById"}},
		Actions: []*descriptor.ResourceAction{{Name: "read", Checks: []*descriptor.ResourceCheck{
			{Name: "owner", MethodName: "checkReadOwner"},
		}}},
		CheckService: new(descriptor.Service{Methods: []*descriptor.Method{
			{Name: "other"},
			{Name: "checkById", Arguments: []*descriptor.Member{{Name: "id"}}},
			{Name: "checkReadOwner", Arguments: []*descriptor.Member{{Name: "ownerId"}}},
		}}),
	})
	encoded, err := json.Marshal(resource)
	if err != nil {
		t.Fatal(err)
	}
	var fields struct {
		Checks  []map[string]json.RawMessage `json:"checks"`
		Actions []struct {
			Checks []map[string]json.RawMessage `json:"checks"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, check := range []map[string]json.RawMessage{fields.Checks[0], fields.Actions[0].Checks[0]} {
		if check["methodName"] == nil || check["method"] != nil || check["arguments"] != nil {
			t.Fatalf("check must serialize only a method reference: %s", encoded)
		}
	}
	var decoded descriptor.Resource
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, current := range []*descriptor.Resource{resource, &decoded} {
		for index, check := range []*descriptor.ResourceCheck{current.Checks[0], current.Actions[0].Checks[0]} {
			method := current.CheckMethod(check)
			if method == nil || method != current.CheckService.Methods[index+1] {
				t.Fatal("check did not resolve to the service's method object")
			}
			current.CheckService.Methods[index+1].Arguments = []*descriptor.Member{{Name: "updated"}}
			if method.Arguments[0].Name != "updated" {
				t.Fatal("check retained stale arguments")
			}
		}
	}
}

func TestResourceCheckMissingReferences(t *testing.T) {
	var absent *descriptor.Resource
	if absent.CheckMethod(new(descriptor.ResourceCheck{MethodName: "check"})) != nil {
		t.Fatal("absent resource has a method")
	}
	for _, service := range []*descriptor.Service{nil, {}, {Methods: []*descriptor.Method{nil, {Name: "other"}, {}}}} {
		resource := new(descriptor.Resource{CheckService: service})
		for _, check := range []*descriptor.ResourceCheck{nil, {}, {MethodName: "missing"}} {
			if resource.CheckMethod(check) != nil {
				t.Fatal("missing check reference resolved")
			}
		}
	}
}
