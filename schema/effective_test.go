package schema_test

import (
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestComputeEffectivePolicyAuthentication(t *testing.T) {
	for _, test := range []struct{ service, method, want schema.AuthMode }{
		{"", "", schema.AuthModeRequired},
		{schema.AuthModeUnset, schema.AuthModeUnset, schema.AuthModeRequired},
		{schema.AuthModeOptional, schema.AuthModeInherit, schema.AuthModeOptional},
		{schema.AuthModeAnonymous, schema.AuthModeUnset, schema.AuthModeAnonymous},
		{schema.AuthModeRequired, schema.AuthModeOptional, schema.AuthModeOptional},
		{schema.AuthModeOptional, schema.AuthModeRequired, schema.AuthModeRequired},
		{schema.AuthModeNoAuth, schema.AuthModeInherit, schema.AuthModeOptional},
		{schema.AuthModeOptional, schema.AuthModeAuth, schema.AuthModeRequired},
		{schema.AuthModeRequired, schema.AuthModeNoAuth, schema.AuthModeOptional},
	} {
		service, method := new(schema.Service{AuthMode: test.service}), new(schema.Method{AuthMode: test.method})
		value, err := schema.ComputeEffectivePolicy(service, method)
		if err != nil || value.AuthMode != test.want || value.Require != nil {
			t.Fatalf("%q/%q: %+v, %v", test.service, test.method, value, err)
		}
		if service.AuthMode != test.service || method.AuthMode != test.method || method.EffectiveAuthMode != "" {
			t.Fatal("pure computation modified declarations")
		}
	}
}

func TestEffectivePolicyCompositionAndRefresh(t *testing.T) {
	service := new(schema.Service{Name: "ReadService", AuthMode: schema.AuthModeRequired, Require: new(schema.PermissionRequire{
		Expression: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeAny, Children: []*schema.PermissionExpression{
			{Mode: schema.PermissionRequireModeCode, Code: "demo.File:read"},
			{Mode: schema.PermissionRequireModeCode, Code: "demo.File:admin"},
		}}),
	})})
	method := new(schema.Method{Name: "read", AuthMode: schema.AuthModeInherit, Require: new(schema.PermissionRequire{
		Expression: new(schema.PermissionExpression{Check: new(schema.PermissionCheckInvocation{
			ResourceSkelName: "external.File", ActionName: "read", CheckName: "owner",
			Arguments: []*schema.PermissionCheckArgument{{JsonPath: "id"}},
		})}),
	})})
	service.Methods = []*schema.Method{method}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Services: []*schema.Service{service}})
	if err := schema.ValidateEffectivePolicy(domain); err == nil {
		t.Fatal("missing effective policy accepted")
	}
	if err := schema.PopulateEffectivePolicies(domain); err != nil {
		t.Fatal(err)
	}
	root := method.EffectiveRequire.Expression
	if root.Mode != schema.PermissionRequireModeAll || len(root.Children) != 2 || root.Children[0].Mode != schema.PermissionRequireModeAny || root.Children[1].Mode != schema.PermissionRequireModeAll {
		t.Fatalf("permission grouping lost: %+v", root)
	}
	check := root.Children[1].Children[1].Check
	if root.Children[1].Children[0].Code != "external.File:read" || check.ServiceSkelName != "" || check.Arguments[0].Type != nil {
		t.Fatal("unresolved check binding changed during composition")
	}
	if err := schema.ValidateEffectivePolicy(domain); err != nil {
		t.Fatal(err)
	}
	check.Arguments[0].JsonPath = "tampered"
	if method.Require.Expression.Check.Arguments[0].JsonPath != "id" {
		t.Fatal("effective and declared bindings alias each other")
	}
	if err := schema.ValidateEffectivePolicy(domain); err == nil || !strings.Contains(err.Error(), "effectiveRequire") {
		t.Fatalf("tampered argument accepted: %v", err)
	}
	service.AuthMode = schema.AuthModeOptional
	if err := schema.PopulateEffectivePolicies(domain); err != nil {
		t.Fatal(err)
	}
	if method.EffectiveAuthMode != schema.AuthModeOptional {
		t.Fatal("effective authentication was not refreshed")
	}
	if err := schema.ValidateEffectivePolicy(domain); err != nil {
		t.Fatal(err)
	}
	// Validation is read-only, including when it rejects stale declarations.
	old := method.EffectiveRequire
	service.Require = nil
	if err := schema.ValidateEffectivePolicy(domain); err == nil || method.EffectiveRequire != old {
		t.Fatal("validation repaired stale policy")
	}
}

func TestEffectivePolicyRejectsInvalidInputsAtomically(t *testing.T) {
	cycle := new(schema.PermissionExpression{Mode: schema.PermissionRequireModeAll})
	cycle.Children = []*schema.PermissionExpression{cycle}
	for _, require := range []*schema.PermissionRequire{
		{}, {Expression: cycle}, {Expression: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeAny, Children: []*schema.PermissionExpression{nil}})},
	} {
		first := new(schema.Method{Name: "first", EffectiveAuthMode: schema.AuthModeAnonymous})
		second := new(schema.Method{Name: "second", Require: require})
		domain := schema.NewDomainFromSpec(schema.DomainSpec{Services: []*schema.Service{{Name: "Service", Methods: []*schema.Method{first, second}}}})
		if err := schema.PopulateEffectivePolicies(domain); err == nil || first.EffectiveAuthMode != schema.AuthModeAnonymous || second.EffectiveAuthMode != "" {
			t.Fatalf("failed derivation changed effective fields: %v", err)
		}
	}
	for _, mode := range []schema.AuthMode{schema.AuthModeOff, "unknown"} {
		if _, err := schema.ComputeEffectivePolicy(new(schema.Service{}), new(schema.Method{AuthMode: mode})); err == nil {
			t.Fatalf("invalid method auth %q accepted", mode)
		}
	}
	shared := new(schema.Method{Name: "shared"})
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Services: []*schema.Service{
		{Name: "One", AuthMode: schema.AuthModeRequired, Methods: []*schema.Method{shared}},
		{Name: "Two", AuthMode: schema.AuthModeOptional, Methods: []*schema.Method{shared}},
	}})
	if err := schema.PopulateEffectivePolicies(domain); err == nil || shared.EffectiveAuthMode != "" {
		t.Fatalf("conflicting shared method policies accepted: %v", err)
	}
}

func TestEffectivePolicyDoesNotReuseCachedValues(t *testing.T) {
	service := new(schema.Service{AuthMode: schema.AuthModeOptional})
	method := new(schema.Method{EffectiveAuthMode: schema.AuthModeRequired})
	value, err := schema.ComputeEffectivePolicy(service, method)
	if err != nil || !reflect.DeepEqual(value, schema.EffectivePolicy{AuthMode: schema.AuthModeOptional}) {
		t.Fatalf("cached value affected computation: %+v, %v", value, err)
	}
}
