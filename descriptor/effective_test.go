package descriptor_test

import (
	"encoding/json"
	"strings"
	"testing"

	"go.yorun.ai/skel/descriptor"
)

func TestValidateEffectivePolicyAfterJSONRoundTrip(t *testing.T) {
	service := new(descriptor.Service{Name: "Files", AuthMode: descriptor.AuthModeOptional,
		Require: new(descriptor.PermissionRequire{Expression: new(descriptor.PermissionExpression{Mode: descriptor.PermissionRequireModeCode, Code: "demo.File:read"})}),
		Methods: []*descriptor.Method{{Name: "read", AuthMode: descriptor.AuthModeInherit,
			Require: new(descriptor.PermissionRequire{Expression: new(descriptor.PermissionExpression{Mode: descriptor.PermissionRequireModeCheck,
				Check: new(descriptor.PermissionCheckInvocation{ResourceSkelName: "external.File", ActionName: "read", CheckName: "owner", ServiceSkelName: "external.FileCheckService", MethodSkelName: "checkOwner", CodeArgumentName: "code",
					Arguments: []*descriptor.PermissionCheckArgument{{Name: "id", JsonPath: "file.id", Type: new(descriptor.Type{Kind: descriptor.TypeKindScalar, Scalar: descriptor.ScalarString})}},
				}),
			})}),
		}},
	})
	method := service.Methods[0]
	value, err := descriptor.ComputeEffectivePolicy(service, method)
	if err != nil {
		t.Fatal(err)
	}
	method.EffectiveAuthMode, method.EffectiveRequire = value.AuthMode, value.Require
	domain := new(descriptor.Domain{Services: []*descriptor.Service{service}})
	if err := descriptor.ValidateEffectivePolicy(domain); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(domain)
	if err != nil {
		t.Fatal(err)
	}
	decode := func() *descriptor.Domain {
		t.Helper()
		var result descriptor.Domain
		if err := json.Unmarshal(encoded, &result); err != nil {
			t.Fatal(err)
		}
		return &result
	}
	if err := descriptor.ValidateEffectivePolicy(decode()); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		change func(*descriptor.Service, *descriptor.Method)
		field  string
	}{
		{"missing auth", func(_ *descriptor.Service, m *descriptor.Method) { m.EffectiveAuthMode = "" }, "effectiveAuthMode"},
		{"inherited auth", func(_ *descriptor.Service, m *descriptor.Method) { m.EffectiveAuthMode = descriptor.AuthModeInherit }, "effectiveAuthMode"},
		{"stale auth", func(s *descriptor.Service, _ *descriptor.Method) { s.AuthMode = descriptor.AuthModeRequired }, "effectiveAuthMode"},
		{"missing require", func(_ *descriptor.Service, m *descriptor.Method) { m.EffectiveRequire = nil }, "effectiveRequire"},
		{"missing service requirement", func(_ *descriptor.Service, m *descriptor.Method) { m.EffectiveRequire = m.Require }, "effectiveRequire"},
		{"changed binding", func(_ *descriptor.Service, m *descriptor.Method) {
			m.EffectiveRequire.Expression.Children[1].Check.Arguments[0].JsonPath = "other"
		}, "effectiveRequire"},
		{"changed type", func(_ *descriptor.Service, m *descriptor.Method) {
			m.EffectiveRequire.Expression.Children[1].Check.Arguments[0].Type.Scalar = descriptor.ScalarInt
		}, "effectiveRequire"},
	} {
		t.Run(test.name, func(t *testing.T) {
			current := decode()
			service := current.Services[0]
			test.change(service, service.Methods[0])
			before, _ := json.Marshal(current)
			err := descriptor.ValidateEffectivePolicy(current)
			if err == nil || !strings.Contains(err.Error(), "services.Files.methods.read."+test.field) {
				t.Fatalf("invalid policy accepted: %v", err)
			}
			after, _ := json.Marshal(current)
			if string(before) != string(after) {
				t.Fatal("validation modified descriptor")
			}
		})
	}
	// Derived bindings have their own values, even before serialization.
	method.EffectiveRequire.Expression.Children[1].Check.Arguments[0].Type.Scalar = descriptor.ScalarInt
	if method.Require.Expression.Check.Arguments[0].Type.Scalar != descriptor.ScalarString {
		t.Fatal("effective type aliases declared type")
	}
}

func TestValidateEffectivePolicyIncludesCallbacks(t *testing.T) {
	for _, owner := range []string{"auth", "permission", "resource"} {
		t.Run(owner, func(t *testing.T) {
			method := new(descriptor.Method{Name: "call", AuthMode: descriptor.AuthModeInherit})
			service := new(descriptor.Service{Name: "Callback", AuthMode: descriptor.AuthModeRequired, Methods: []*descriptor.Method{method}})
			domain := new(descriptor.Domain{})
			switch owner {
			case "auth":
				domain.Actors = []*descriptor.Actor{{Auth: new(descriptor.ActorAuth{Service: service})}}
			case "permission":
				domain.Actors = []*descriptor.Actor{{Permission: new(descriptor.ActorPermission{Service: service})}}
			case "resource":
				domain.Resources = []*descriptor.Resource{{CheckService: service}}
			}
			if err := descriptor.ValidateEffectivePolicy(domain); err == nil {
				t.Fatal("unpopulated callback policy accepted")
			}
			method.EffectiveAuthMode = descriptor.AuthModeRequired
			if err := descriptor.ValidateEffectivePolicy(domain); err != nil {
				t.Fatal(err)
			}
		})
	}
}
