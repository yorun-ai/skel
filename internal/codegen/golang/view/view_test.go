package view

import (
	"slices"
	"testing"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
)

func TestViewSeparatesPubResources(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo",
		Resources: []*model.Resource{
			{Pub: true, Name: "PublicUser", Actions: []*model.ResourceAction{{Name: "read"}}},
			{Name: "LocalUser", Actions: []*model.ResourceAction{{Name: "read"}}},
		},
	})

	pubView, err := New(ModePub, domain)
	if err != nil {
		t.Fatal(err)
	}
	if len(pubView.Resources) != 1 || pubView.Resources[0].Name != "PublicUser" {
		t.Fatalf("unexpected pub resources: %+v", pubView.Resources)
	}

	regularView, err := New(ModeRegular, domain)
	if err != nil {
		t.Fatal(err)
	}
	if len(regularView.Resources) != 1 || regularView.Resources[0].Name != "LocalUser" {
		t.Fatalf("unexpected regular resources: %+v", regularView.Resources)
	}
}

func TestPubViewRejectsServiceRequiringNonPubResource(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo",
		Resources: []*model.Resource{
			{Name: "LocalUser", Actions: []*model.ResourceAction{{Name: "read"}}},
		},
		Services: []*model.Service{
			{
				Pub:      true,
				Name:     "UserService",
				SkelName: "demo.UserService",
				Require: &model.PermissionRequire{
					Expr: &model.PermissionExpr{
						Mode: model.PermissionRequireModeCode,
						Code: "demo.LocalUser:read",
					},
				},
			},
		},
	})

	if _, err := New(ModePub, domain); err == nil {
		t.Fatal("expected invalid public view error")
	}
}

func TestFullViewKeepsEveryDeclaration(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo",
		Data: []*model.Data{{Name: "Public", Pub: true}, {Name: "Local"}},
		Enums: []*model.Enum{
			{Name: "PublicStatus", Pub: true},
			{Name: "LocalStatus"},
		},
		Resources: []*model.Resource{
			{Pub: true, Name: "PublicUser"},
			{Name: "LocalUser"},
		},
		Services: []*model.Service{
			{Pub: true, Name: "PublicService", SkelName: "demo.PublicService"},
			{Name: "LocalService", SkelName: "demo.LocalService"},
		},
	})

	full := Full(domain)

	if len(full.Data) != 2 || len(full.Enums) != 2 || len(full.Resources) != 2 || len(full.Services) != 2 {
		t.Fatalf("unexpected full view: %+v", full)
	}
}

func TestViewTypeRootsRespectArgumentSource(t *testing.T) {
	declared := &model.Type{Kind: model.TypeKindScalar}
	injected := &model.Type{Kind: model.TypeKindScalar}
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo",
		Services: []*model.Service{{Name: "ExampleApiService", Api: true, Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}, Methods: []*model.Method{{
			Arguments: []*model.Argument{
				{Name: "input", Type: declared, Source: model.ArgumentSourceDeclared},
				{Name: "code", Type: injected, Source: model.ArgumentSourcePermissionCode},
			},
		}}}},
	})
	for _, mode := range []Mode{ModeApi, ModeFull, ModeRegular} {
		t.Run(string(mode), func(t *testing.T) {
			view, err := Build(mode, domain, common.ApiFilter{})
			if err != nil {
				t.Fatal(err)
			}
			roots := view.TypeRoots()
			if !slices.Contains(roots, declared) || slices.Contains(roots, injected) != (mode != ModeApi) {
				t.Fatalf("wrong argument roots for %s: %v", mode, roots)
			}
		})
	}
}
