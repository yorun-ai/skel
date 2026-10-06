package codegen

import (
	"testing"

	"go.yorun.ai/skel/internal/model"
)

func containsData(data []*model.Data, want *model.Data) bool {
	for _, item := range data {
		if item == want {
			return true
		}
	}
	return false
}

func TestBuildApiViewCollectsClientDataAndDependencies(t *testing.T) {
	dependency := &model.Data{Name: "Secret"}
	status := &model.Enum{Name: "Status"}
	payload := &model.Data{Name: "Payload", Members: []*model.DataMember{
		{Name: "secret", Type: &model.Type{Kind: model.TypeKindData, Data: dependency}},
		{Name: "status", Type: &model.Type{Kind: model.TypeKindEnum, Enum: status}},
	}}
	unused := &model.Data{Name: "Unused"}
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name:  "demo.user",
		Data:  []*model.Data{payload, dependency, unused},
		Enums: []*model.Enum{status, {Name: "UnusedStatus"}},
		Services: []*model.Service{{
			Name:      "UserService",
			SkelName:  "demo.user.UserService",
			Api:       true,
			Audiences: []*model.ActorAudience{{Actor: "UserActor"}},
			Methods: []*model.Method{{
				Name:       "get",
				SkelName:   "get",
				ResultType: &model.Type{Kind: model.TypeKindData, Data: payload},
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
	external := &model.Data{Name: "Remote", Domain: "identity.user"}
	externalType := &model.Type{
		Kind:           model.TypeKindData,
		Data:           external,
		ExternalDomain: "identity.user",
	}
	pubData := &model.Data{Pub: true, Name: "PublicPayload"}
	local := &model.Data{Name: "LocalPayload", Members: []*model.DataMember{{Name: "remote", Type: externalType}}}
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{pubData, local},
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
	payload := &model.Data{Name: "Payload"}
	memberType := &model.Type{Kind: model.TypeKindData, Data: payload}
	data := &model.Data{Name: "Envelope", Members: []*model.DataMember{{Name: "payload", Type: memberType}}}
	resultType := &model.Type{Kind: model.TypeKindData, Data: payload}
	declaredArgument := &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString}
	runtimeArgument := &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString}
	service := &model.Service{
		Name:     "UserService",
		SkelName: "demo.user.UserService",
		Methods: []*model.Method{{
			Name:       "get",
			SkelName:   "get",
			ResultType: resultType,
			Arguments: []*model.Argument{
				{Name: "id", Source: model.ArgumentSourceDeclared, Type: declaredArgument},
				{Name: "code", Source: model.ArgumentSourcePermissionCode, Type: runtimeArgument},
			},
		}},
	}

	roots := ApiTypeRoots([]*model.Data{data}, []*model.Service{service})

	if len(roots) != 3 {
		t.Fatalf("expected member, result and declared argument types: %+v", roots)
	}
	if roots[0] != memberType || roots[1] != resultType || roots[2] != declaredArgument {
		t.Fatalf("unexpected API type roots: %+v", roots)
	}
}
