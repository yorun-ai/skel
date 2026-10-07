package descriptor_test

import (
	"encoding/json"
	"testing"

	"go.yorun.ai/skel/descriptor"
)

func TestActorAudienceMatching(t *testing.T) {
	audiences := []*descriptor.ActorAudience{
		{
			SkelName: "demo.User",
			Via:      descriptor.ActorViaClient,
		},
		{
			SkelName: "demo.Admin",
		},
	}
	service := new(descriptor.Service{
		Audiences: audiences,
	})
	web := new(descriptor.Web{
		Audiences: audiences,
	})
	for _, test := range []struct {
		actor string
		via   descriptor.ActorViaKind
		want  bool
	}{
		{
			actor: "demo.User",
			via:   descriptor.ActorViaClient,
			want:  true,
		},
		{
			actor: "demo.User",
			via:   descriptor.ActorViaAgent,
		},
		{
			actor: "demo.Admin",
			via:   descriptor.ActorViaOpenAPI,
			want:  true,
		},
		{
			actor: "other.User",
			via:   descriptor.ActorViaClient,
		},
	} {
		if got := service.HasAudience(test.actor, test.via); got != test.want {
			t.Errorf("Service.HasAudience(%q, %q) = %v, want %v", test.actor, test.via, got, test.want)
		}
		if got := web.HasAudience(test.actor, test.via); got != test.want {
			t.Errorf("Web.HasAudience(%q, %q) = %v, want %v", test.actor, test.via, got, test.want)
		}
	}
	var emptyService descriptor.Service
	var emptyWeb descriptor.Web
	if emptyService.HasAudience("demo.User", descriptor.ActorViaClient) || emptyWeb.HasAudience("demo.User", descriptor.ActorViaClient) {
		t.Fatal("empty audiences must not grant access")
	}
}

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
