package common

import (
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/model"
)

func TestValidateDomainRejectsMalformedNestedModels(t *testing.T) {
	tests := []struct {
		name     string
		spec     model.DomainSpec
		expected string
	}{
		{name: "api permission callback", spec: model.DomainSpec{Actors: []*model.Actor{{Name: "Client", PermService: &model.Service{Name: "Permission", Api: true}}}}, expected: "cannot be used as a framework callback"},
		{name: "api resource callback", spec: model.DomainSpec{Resources: []*model.Resource{{Name: "Document", CheckService: &model.Service{Name: "Check", Api: true}}}}, expected: "cannot be used as a framework callback"},
		{name: "api missing actor", spec: model.DomainSpec{Services: []*model.Service{{Name: "OrderApiService", Api: true}}}, expected: "at least one for Actor"},
		{name: "api and pub", spec: model.DomainSpec{Services: []*model.Service{{Name: "OrderService", Api: true, Pub: true}}}, expected: "cannot combine api and pub"},
		{name: "nil import", spec: model.DomainSpec{Imports: []*model.Import{nil}}, expected: "nil import"},
		{name: "missing imported domain", spec: model.DomainSpec{Imports: []*model.Import{{Name: "shared"}}}, expected: "has no domain model"},
		{name: "malformed imported domain", spec: model.DomainSpec{Imports: []*model.Import{{Name: "shared", Domain: model.NewDomainFromSpec(model.DomainSpec{Webs: []*model.Web{nil}})}}}, expected: "import shared"},
		{name: "incomplete actor auth", spec: model.DomainSpec{Actors: []*model.Actor{{Name: "Client", AuthEnabled: true}}}, expected: "incomplete auth support"},
		{name: "incomplete actor permission", spec: model.DomainSpec{Actors: []*model.Actor{{Name: "Client", PermEnabled: true}}}, expected: "incomplete permission support"},
		{name: "malformed optional permission service", spec: model.DomainSpec{Actors: []*model.Actor{{Name: "Client", PermService: &model.Service{Name: "Permission", Methods: []*model.Method{nil}}}}}, expected: "nil method"},
		{name: "nil resource action", spec: model.DomainSpec{Resources: []*model.Resource{{Name: "Document", Actions: []*model.ResourceAction{nil}}}}, expected: "nil action"},
		{name: "nil resource check", spec: model.DomainSpec{Resources: []*model.Resource{{Name: "Document", Checks: []*model.ResourceCheck{nil}}}}, expected: "nil check"},
		{name: "resource check without method", spec: model.DomainSpec{Resources: []*model.Resource{{Name: "Document", Checks: []*model.ResourceCheck{{Name: "owner"}}}}}, expected: "check owner is nil"},
		{name: "nil web", spec: model.DomainSpec{Webs: []*model.Web{nil}}, expected: "nil web"},
		{name: "nil web audience", spec: model.DomainSpec{Webs: []*model.Web{{Name: "Portal", Audiences: []*model.ActorAudience{nil}}}}, expected: "nil audience"},
		{name: "nil service audience", spec: model.DomainSpec{Services: []*model.Service{{Name: "Documents", Audiences: []*model.ActorAudience{nil}}}}, expected: "nil audience"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			domain := model.NewDomainFromSpec(test.spec)
			err := ValidateDomain(domain)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, err)
			}
		})
	}
}

func TestValidateDomainRejectsCyclicStructuralTypes(t *testing.T) {
	kind := new(model.Type{Kind: model.TypeKindList})
	kind.List = new(model.ListType{Value: kind})
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Data: []*model.Data{{Name: "Node", Kind: model.DataKindData, Members: []*model.DataMember{{Name: "value", Type: kind}}}}})
	if err := ValidateDomain(domain); err == nil || !strings.Contains(err.Error(), "cyclic type") {
		t.Fatalf("expected cyclic type error, got %v", err)
	}
}

func TestValidateDomainRejectsMalformedGenerics(t *testing.T) {
	box := new(model.Data{Name: "Box", Kind: model.DataKindData, TypeParameters: []*model.TypeParameter{{Name: "T"}}})
	for _, kind := range []*model.Type{
		{Kind: model.TypeKindData, Data: box},
		{Kind: model.TypeKindData, Data: box, TypeArguments: []*model.Type{nil}},
		{Kind: model.TypeKindData, Data: box, TypeArguments: []*model.Type{{Kind: model.TypeKindScalar, Scalar: model.ScalarString}, {Kind: model.TypeKindScalar, Scalar: model.ScalarBinary}}},
	} {
		domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Data: []*model.Data{{Name: "Value", Kind: model.DataKindData, Members: []*model.DataMember{{Name: "box", Type: kind}}}}})
		if err := ValidateDomain(domain); err == nil {
			t.Fatal("expected invalid generic model to be rejected")
		}
	}
}

func TestValidateDomainRejectsCyclicPermissionExpressions(t *testing.T) {
	expression := new(model.PermissionExpr{Mode: model.PermissionRequireModeAll})
	expression.Children = []*model.PermissionExpr{expression}
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Services: []*model.Service{{Name: "Service", Require: new(model.PermissionRequire{Expr: expression})}}})
	if err := ValidateDomain(domain); err == nil || !strings.Contains(err.Error(), "cyclic permission") {
		t.Fatalf("expected cyclic permission error, got %v", err)
	}
}

func TestValidateDomainAllowsRecursiveDataAndSharedTypes(t *testing.T) {
	data := new(model.Data{Name: "Node", Kind: model.DataKindData})
	reference := new(model.Type{Kind: model.TypeKindData, Data: data, Nullable: true})
	kind := new(model.Type{Kind: model.TypeKindMap, Map: new(model.MapType{Key: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString}), Value: reference})})
	data.Members = []*model.DataMember{{Name: "next", Type: reference}, {Name: "children", Type: kind}, {Name: "otherChildren", Type: kind}}
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Data: []*model.Data{data}})
	if err := ValidateDomain(domain); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDomainRejectsMalformedReferencedData(t *testing.T) {
	referenced := new(model.Data{Name: "Nested", Kind: model.DataKindData, Members: []*model.DataMember{nil}})
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Data: []*model.Data{{Name: "Value", Kind: model.DataKindData, Members: []*model.DataMember{{Name: "nested", Type: new(model.Type{Kind: model.TypeKindData, Data: referenced})}}}}})
	if err := ValidateDomain(domain); err == nil || !strings.Contains(err.Error(), "nil member") {
		t.Fatalf("expected referenced data error, got %v", err)
	}
}
