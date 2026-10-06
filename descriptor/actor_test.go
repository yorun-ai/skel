package descriptor_test

import (
	"encoding/json"
	"testing"

	"go.yorun.ai/skel/descriptor"
)

func TestActorMethodsReferenceServiceAfterJSONRoundTrip(t *testing.T) {
	actor := new(descriptor.Actor{
		Auth: new(descriptor.ActorAuth{
			MethodName: "authenticate",
			Service: new(descriptor.Service{Methods: []*descriptor.Method{
				{Name: "other"}, {Name: "authenticate", Hash: "auth-hash"},
			}}),
		}),
		Permission: new(descriptor.ActorPermission{
			MethodName: "check",
			Service: new(descriptor.Service{Methods: []*descriptor.Method{
				{Name: "other"}, {Name: "check", Hash: "permission-hash"},
			}}),
		}),
	})
	encoded, err := json.Marshal(actor)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, capability := range []string{"auth", "permission"} {
		var fieldsOfCapability map[string]json.RawMessage
		if err := json.Unmarshal(fields[capability], &fieldsOfCapability); err != nil {
			t.Fatal(err)
		}
		if fieldsOfCapability["method"] != nil || fieldsOfCapability["methodName"] == nil {
			t.Fatalf("%s must serialize a method reference, got %s", capability, fields[capability])
		}
	}
	var decoded descriptor.Actor
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, current := range []*descriptor.Actor{actor, &decoded} {
		if current.Auth.Method() != current.Auth.Service.Methods[1] || current.Permission.Method() != current.Permission.Service.Methods[1] {
			t.Fatal("accessor did not return the service's method object")
		}
		current.Auth.Service.Methods[1].Hash = "updated"
		current.Permission.Service.Methods[1].Hash = "updated"
		if current.Auth.Method().Hash != "updated" || current.Permission.Method().Hash != "updated" {
			t.Fatal("accessor retained stale method metadata")
		}
	}
}

func TestActorMethodMissingReferences(t *testing.T) {
	var auth *descriptor.ActorAuth
	var permission *descriptor.ActorPermission
	if auth.Method() != nil || permission.Method() != nil {
		t.Fatal("absent capability has a method")
	}
	for _, service := range []*descriptor.Service{nil, {}, {Methods: []*descriptor.Method{nil, {Name: "other"}, {}}}} {
		for _, name := range []string{"", "missing"} {
			auth = new(descriptor.ActorAuth{Service: service, MethodName: name})
			permission = new(descriptor.ActorPermission{Service: service, MethodName: name})
			if auth.Method() != nil || permission.Method() != nil {
				t.Fatal("missing method reference resolved")
			}
		}
	}
}
