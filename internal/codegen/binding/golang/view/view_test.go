package view

import (
	"slices"
	"testing"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/schema"
)

func TestViewSeparatesPubResources(t *testing.T) {
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo",
		Resources: []*schema.Resource{
			{Pub: true, Name: "PublicUser", Actions: []*schema.ResourceAction{{Name: "read"}}},
			{Name: "LocalUser", Actions: []*schema.ResourceAction{{Name: "read"}}},
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
	if !slices.Equal(regularView.Reexports.Resources, pubView.Resources) {
		t.Fatalf("regular facade lost public resources: %+v", regularView.Reexports.Resources)
	}
}

func TestPubViewRejectsServiceRequiringNonPubResource(t *testing.T) {
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo",
		Resources: []*schema.Resource{
			{Name: "LocalUser", Actions: []*schema.ResourceAction{{Name: "read"}}},
		},
		Services: []*schema.Service{
			{
				Pub:      true,
				Name:     "UserService",
				SkelName: "demo.UserService",
				Require: &schema.PermissionRequire{
					Expression: &schema.PermissionExpression{
						Mode: schema.PermissionRequireModeCode,
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
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo",
		Data: []*schema.Data{{Name: "Public", Pub: true}, {Name: "Local"}},
		Enums: []*schema.Enum{
			{Name: "PublicStatus", Pub: true},
			{Name: "LocalStatus"},
		},
		Resources: []*schema.Resource{
			{Pub: true, Name: "PublicUser"},
			{Name: "LocalUser"},
		},
		Services: []*schema.Service{
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
	declared := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}
	injected := &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo",
		Services: []*schema.Service{{Name: "ExampleApiService", Api: true, Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}, Methods: []*schema.Method{{
			Arguments: []*schema.Argument{
				{Name: "input", Type: declared, Source: schema.ArgumentSourceDeclared},
				{Name: "code", Type: injected, Source: schema.ArgumentSourcePermissionCode},
			},
		}}}},
	})
	for _, mode := range []Mode{ModeApi, ModeFull, ModeRegular} {
		t.Run(string(mode), func(t *testing.T) {
			view, err := Build(mode, domain, codegen.ApiFilter{})
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
