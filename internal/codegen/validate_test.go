package codegen

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestValidateDomainRejectsMalformedNestedSchemas(t *testing.T) {
	tests := []struct {
		name     string
		spec     schema.DomainSpec
		expected string
	}{
		{name: "api permission callback", spec: schema.DomainSpec{Actors: []*schema.Actor{{Name: "Client", Permission: new(schema.ActorPermission{Service: &schema.Service{Name: "Permission", Api: true}, Method: new(schema.Method{})})}}}, expected: "cannot be used as a framework callback"},
		{name: "api resource callback", spec: schema.DomainSpec{Resources: []*schema.Resource{{Name: "Document", CheckService: &schema.Service{Name: "Check", Api: true}}}}, expected: "cannot be used as a framework callback"},
		{name: "api missing actor", spec: schema.DomainSpec{Services: []*schema.Service{{Name: "OrderApiService", Api: true}}}, expected: "at least one for Actor"},
		{name: "ext and pub", spec: schema.DomainSpec{Services: []*schema.Service{{Name: "StorageService", Ext: true, Pub: true}}}, expected: "ext, api and pub are mutually exclusive"},
		{name: "ext and api", spec: schema.DomainSpec{Services: []*schema.Service{{Name: "StorageApiService", Ext: true, Api: true}}}, expected: "ext, api and pub are mutually exclusive"},
		{name: "api and pub", spec: schema.DomainSpec{Services: []*schema.Service{{Name: "OrderService", Api: true, Pub: true}}}, expected: "cannot combine api and pub"},
		{name: "nil import", spec: schema.DomainSpec{Imports: []*schema.Import{nil}}, expected: "nil import"},
		{name: "missing imported domain", spec: schema.DomainSpec{Imports: []*schema.Import{{Name: "shared"}}}, expected: "has no domain schema"},
		{name: "malformed imported domain", spec: schema.DomainSpec{Imports: []*schema.Import{{Name: "shared", Domain: schema.NewDomainFromSpec(schema.DomainSpec{Webs: []*schema.Web{nil}})}}}, expected: "import shared"},
		{name: "incomplete actor auth", spec: schema.DomainSpec{Actors: []*schema.Actor{{Name: "Client", Auth: new(schema.ActorAuth{})}}}, expected: "incomplete auth support"},
		{name: "incomplete actor permission", spec: schema.DomainSpec{Actors: []*schema.Actor{{Name: "Client", Permission: new(schema.ActorPermission{})}}}, expected: "incomplete permission support"},
		{name: "malformed actor permission service", spec: schema.DomainSpec{Actors: []*schema.Actor{{Name: "Client", Permission: new(schema.ActorPermission{Service: &schema.Service{Name: "Permission", Methods: []*schema.Method{nil}}, Method: new(schema.Method{})})}}}, expected: "nil method"},
		{name: "nil resource action", spec: schema.DomainSpec{Resources: []*schema.Resource{{Name: "Document", Actions: []*schema.ResourceAction{nil}}}}, expected: "nil action"},
		{name: "nil resource check", spec: schema.DomainSpec{Resources: []*schema.Resource{{Name: "Document", Checks: []*schema.ResourceCheck{nil}}}}, expected: "nil check"},
		{name: "resource check without method", spec: schema.DomainSpec{Resources: []*schema.Resource{{Name: "Document", Checks: []*schema.ResourceCheck{{Name: "owner"}}}}}, expected: "check owner is nil"},
		{name: "nil web", spec: schema.DomainSpec{Webs: []*schema.Web{nil}}, expected: "nil web"},
		{name: "nil web audience", spec: schema.DomainSpec{Webs: []*schema.Web{{Name: "Portal", Audiences: []*schema.ActorAudience{nil}}}}, expected: "nil audience"},
		{name: "nil service audience", spec: schema.DomainSpec{Services: []*schema.Service{{Name: "Documents", Audiences: []*schema.ActorAudience{nil}}}}, expected: "nil audience"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			domain := schema.NewDomainFromSpec(test.spec)
			err := ValidateDomain(domain)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, err)
			}
		})
	}
}

func TestValidateDomainRejectsCyclicStructuralTypes(t *testing.T) {
	kind := new(schema.Type{Kind: schema.TypeKindList})
	kind.List = new(schema.ListType{Value: kind})
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Data: []*schema.Data{{Name: "Node", Kind: schema.DataKindData, Members: []*schema.DataMember{{Name: "value", Type: kind}}}}})
	if err := ValidateDomain(domain); err == nil || !strings.Contains(err.Error(), "cyclic type") {
		t.Fatalf("expected cyclic type error, got %v", err)
	}
}

func TestValidateDomainRejectsMalformedGenerics(t *testing.T) {
	box := new(schema.Data{Name: "Box", Kind: schema.DataKindData, TypeParameters: []*schema.TypeParameter{{Name: "T"}}})
	for _, kind := range []*schema.Type{
		{Kind: schema.TypeKindData, Data: box},
		{Kind: schema.TypeKindData, Data: box, TypeArguments: []*schema.Type{nil}},
		{Kind: schema.TypeKindData, Data: box, TypeArguments: []*schema.Type{{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}, {Kind: schema.TypeKindScalar, Scalar: schema.ScalarBinary}}},
	} {
		domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Data: []*schema.Data{{Name: "Value", Kind: schema.DataKindData, Members: []*schema.DataMember{{Name: "box", Type: kind}}}}})
		if err := ValidateDomain(domain); err == nil {
			t.Fatal("expected invalid generic schema to be rejected")
		}
	}
}

func TestValidateDomainRejectsCyclicPermissionExpressions(t *testing.T) {
	expression := new(schema.PermissionExpression{Mode: schema.PermissionRequireModeAll})
	expression.Children = []*schema.PermissionExpression{expression}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Services: []*schema.Service{{Name: "Service", Require: new(schema.PermissionRequire{Expression: expression})}}})
	if err := ValidateDomain(domain); err == nil || !strings.Contains(err.Error(), "cyclic permission") {
		t.Fatalf("expected cyclic permission error, got %v", err)
	}
}

func TestValidateDomainAllowsRecursiveDataAndSharedTypes(t *testing.T) {
	data := new(schema.Data{Name: "Node", Kind: schema.DataKindData})
	reference := new(schema.Type{Kind: schema.TypeKindData, Data: data, Nullable: true})
	kind := new(schema.Type{Kind: schema.TypeKindMap, Map: new(schema.MapType{Key: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}), Value: reference})})
	data.Members = []*schema.DataMember{{Name: "next", Type: reference}, {Name: "children", Type: kind}, {Name: "otherChildren", Type: kind}}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Data: []*schema.Data{data}})
	if err := ValidateDomain(domain); err != nil {
		t.Fatal(err)
	}
}

func TestValidateExtensionEventModifier(t *testing.T) {
	for _, data := range []*schema.Data{
		{Name: "Invalid", Kind: schema.DataKindData, Ext: true},
		{Name: "InvalidConfig", Kind: schema.DataKindConfig, Ext: true},
		{Name: "InvalidEvent", Kind: schema.DataKindEvent, Pub: true, Ext: true},
	} {
		if err := validateData(data); err == nil {
			t.Fatalf("accepted invalid extension: %+v", data)
		}
	}
	if err := validateData(&schema.Data{Name: "AuditRecordedEvent", Kind: schema.DataKindEvent, Ext: true}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDomainRejectsMalformedReferencedData(t *testing.T) {
	referenced := new(schema.Data{Name: "Nested", Kind: schema.DataKindData, Members: []*schema.DataMember{nil}})
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Data: []*schema.Data{{Name: "Value", Kind: schema.DataKindData, Members: []*schema.DataMember{{Name: "nested", Type: new(schema.Type{Kind: schema.TypeKindData, Data: referenced})}}}}})
	if err := ValidateDomain(domain); err == nil || !strings.Contains(err.Error(), "nil member") {
		t.Fatalf("expected referenced data error, got %v", err)
	}
}

func TestPrepareValidatesMethodAuthByOwner(t *testing.T) {
	for _, owner := range []string{"method", "service", "web"} {
		for _, mode := range []schema.AuthMode{schema.AuthModeInherit, schema.AuthModeOff, "invalid"} {
			t.Run(owner+"/"+string(mode), func(t *testing.T) {
				spec := schema.DomainSpec{Name: "demo"}
				switch owner {
				case "method":
					spec.Services = []*schema.Service{{Name: "ReadService", Methods: []*schema.Method{{Name: "read", Auth: mode}}}}
				case "service":
					spec.Services = []*schema.Service{{Name: "ReadService", Auth: mode}}
				case "web":
					spec.Webs = []*schema.Web{{Name: "PortalWeb", Auth: mode}}
				}
				_, err := Prepare(schema.NewDomainFromSpec(spec), Selection{})
				valid := owner == "method" && mode == schema.AuthModeInherit || owner == "web" && mode == schema.AuthModeOff
				if (err == nil) != valid {
					t.Fatalf("valid=%v, got %v", valid, err)
				}
			})
		}
	}
}

func TestPrepareRequiresCanonicalCallbackMethods(t *testing.T) {
	for _, owner := range []string{"actor auth", "actor permission", "resource", "resource action"} {
		for _, mutation := range []string{"none", "missing", "copy", "duplicate", "unnamed", "no service"} {
			t.Run(owner+"/"+mutation, func(t *testing.T) {
				method := new(schema.Method{Name: "call"})
				service := new(schema.Service{Name: "CallbackService", Methods: []*schema.Method{method}})
				wantError := ""
				switch mutation {
				case "missing":
					service.Methods = []*schema.Method{{Name: "other"}}
					wantError = "not found in service"
				case "copy":
					copy := *method
					service.Methods = []*schema.Method{&copy}
					wantError = "must reference the method node"
				case "duplicate":
					service.Methods = append(service.Methods, new(schema.Method{Name: method.Name}))
					wantError = "duplicate method"
				case "unnamed":
					method.Name = ""
					wantError = "no method name"
				case "no service":
					service = nil
					wantError = "no service"
					if strings.HasPrefix(owner, "actor") {
						wantError = "incomplete"
					}
				}
				spec := schema.DomainSpec{Name: "demo"}
				switch owner {
				case "actor auth":
					spec.Actors = []*schema.Actor{{Name: "ClientActor", Auth: new(schema.ActorAuth{
						Credential: new(schema.Data{Kind: schema.DataKindData}), Info: new(schema.Data{Kind: schema.DataKindData}),
						Service: service, Method: method,
					})}}
				case "actor permission":
					spec.Actors = []*schema.Actor{{Name: "ClientActor", Permission: new(schema.ActorPermission{Service: service, Method: method})}}
				default:
					resource := new(schema.Resource{Name: "Document", CheckService: service})
					checks := []*schema.ResourceCheck{{Name: "byId", Method: method}}
					if owner == "resource" {
						resource.Checks = checks
					} else {
						resource.Actions = []*schema.ResourceAction{{Name: "read", Checks: checks}}
					}
					spec.Resources = []*schema.Resource{resource}
				}
				_, err := Prepare(schema.NewDomainFromSpec(spec), Selection{})
				if wantError == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), wantError) {
					t.Fatalf("expected %q, got %v", wantError, err)
				}
			})
		}
	}
}
