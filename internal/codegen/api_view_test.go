package codegen

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func containsData(data []*schema.Data, want *schema.Data) bool {
	for _, item := range data {
		if item == want {
			return true
		}
	}
	return false
}

func TestBuildApiViewCollectsClientDataAndDependencies(t *testing.T) {
	dependency := &schema.Data{Name: "Secret"}
	status := &schema.Enum{Name: "Status"}
	payload := &schema.Data{Name: "Payload", Members: []*schema.DataMember{
		{Name: "secret", Type: &schema.Type{Kind: schema.TypeKindData, Data: dependency}},
		{Name: "status", Type: &schema.Type{Kind: schema.TypeKindEnum, Enum: status}},
	}}
	unused := &schema.Data{Name: "Unused"}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name:  "demo.user",
		Data:  []*schema.Data{payload, dependency, unused},
		Enums: []*schema.Enum{status, {Name: "UnusedStatus"}},
		Services: []*schema.Service{{
			Name:      "UserService",
			SkelName:  "demo.user.UserService",
			Api:       true,
			Audiences: []*schema.ActorAudience{{Actor: "UserActor"}},
			Methods: []*schema.Method{{
				Name:       "get",
				SkelName:   "get",
				ResultType: &schema.Type{Kind: schema.TypeKindData, Data: payload},
			}},
		}},
	})

	view, err := BuildApiView(domain, ApiFilter{})
	if err != nil {
		t.Fatal(err)
	}

	if len(view.Services) != 1 || view.Services[0].SkelName != "demo.user.UserService" {
		t.Fatalf("expected the client service, got %+v", view.Services)
	}
	if !containsData(view.Data, payload) || !containsData(view.Data, dependency) {
		t.Fatalf("client data and its dependencies must be present: %+v", view.Data)
	}
	if containsData(view.Data, unused) {
		t.Fatalf("unreferenced data must be excluded: %+v", view.Data)
	}
	if len(view.Enums) != 1 || view.Enums[0] != status {
		t.Fatalf("expected only the referenced enum: %+v", view.Enums)
	}
}

func TestBuildApiViewKeepsPublicTypesAndSkipsExternalDependencies(t *testing.T) {
	external := &schema.Data{Name: "Remote", Domain: "identity.user"}
	externalType := &schema.Type{
		Kind:           schema.TypeKindData,
		Data:           external,
		ExternalDomain: "identity.user",
	}
	pubData := &schema.Data{Pub: true, Name: "PublicPayload"}
	local := &schema.Data{Name: "LocalPayload", Members: []*schema.DataMember{{Name: "remote", Type: externalType}}}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{pubData, local},
	})

	view, err := BuildApiView(domain, ApiFilter{})
	if err != nil {
		t.Fatal(err)
	}

	if !containsData(view.Data, pubData) {
		t.Fatalf("public data must stay in the API view: %+v", view.Data)
	}
	if containsData(view.Data, external) {
		t.Fatalf("external data must not enter the API view: %+v", view.Data)
	}
}

func TestApiTypeRootsCollectsDeclaredTypes(t *testing.T) {
	payload := &schema.Data{Name: "Payload"}
	memberType := &schema.Type{Kind: schema.TypeKindData, Data: payload}
	data := &schema.Data{Name: "Envelope", Members: []*schema.DataMember{{Name: "payload", Type: memberType}}}
	resultType := &schema.Type{Kind: schema.TypeKindData, Data: payload}
	declaredArgument := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}
	runtimeArgument := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}
	service := &schema.Service{
		Name:     "UserService",
		SkelName: "demo.user.UserService",
		Methods: []*schema.Method{{
			Name:       "get",
			SkelName:   "get",
			ResultType: resultType,
			Arguments: []*schema.Argument{
				{Name: "id", Source: schema.ArgumentSourceDeclared, Type: declaredArgument},
				{Name: "code", Source: schema.ArgumentSourcePermissionCode, Type: runtimeArgument},
			},
		}},
	}

	roots := ApiTypeRoots([]*schema.Data{data}, []*schema.Service{service})

	if len(roots) != 3 {
		t.Fatalf("expected member, result and declared argument types: %+v", roots)
	}
	if roots[0] != memberType || roots[1] != resultType || roots[2] != declaredArgument {
		t.Fatalf("unexpected API type roots: %+v", roots)
	}
}
