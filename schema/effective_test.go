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

func TestComposeRequirementsOrdersWithinGroupsWithoutChangingDeclarations(t *testing.T) {
	check := func(name string) *schema.PermissionExpression {
		return new(schema.PermissionExpression{
			Mode: schema.PermissionRequireModeCheck,
			Check: new(schema.PermissionCheckInvocation{
				CheckName: name,
			}),
		})
	}
	first, last, nested := check("first"), check("last"), check("nested")
	unresolved := new(schema.PermissionExpression{
		Check: new(schema.PermissionCheckInvocation{
			ResourceSkelName: "external.File",
			ActionName:       "read",
			CheckName:        "owner",
		}),
	})
	group := new(schema.PermissionExpression{
		Mode: schema.PermissionRequireModeAny,
		Children: []*schema.PermissionExpression{
			nested, unresolved,
			{Mode: schema.PermissionRequireModeCode, Code: "nested-code"},
		},
	})
	declared := new(schema.PermissionExpression{
		Mode: schema.PermissionRequireModeAll,
		Children: []*schema.PermissionExpression{
			first, group,
			{Mode: schema.PermissionRequireModeCode, Code: "first-code"},
			last,
			{Mode: schema.PermissionRequireModeCode, Code: "last-code"},
		},
	})
	value, err := schema.ComposeRequirements(
		new(schema.PermissionRequire{Expression: declared}),
		new(schema.PermissionRequire{Expression: new(schema.PermissionExpression{
			Mode: schema.PermissionRequireModeCode, Code: "method-code",
		})}),
	)
	if err != nil {
		t.Fatal(err)
	}
	root := value.Expression
	if root.Mode != schema.PermissionRequireModeAll || len(root.Children) != 2 || root.Children[0].Code != "method-code" {
		t.Fatalf("conjoined requirements were not ordered: %+v", root)
	}
	ordered := root.Children[1]
	if ordered.Mode != schema.PermissionRequireModeAll || len(ordered.Children) != 5 ||
		ordered.Children[0].Code != "first-code" || ordered.Children[1].Code != "last-code" ||
		ordered.Children[2].Mode != schema.PermissionRequireModeAny ||
		ordered.Children[3].Check.CheckName != "first" || ordered.Children[4].Check.CheckName != "last" {
		t.Fatalf("grouping or stable evaluation order lost: %+v", ordered)
	}
	children := ordered.Children[2].Children
	if len(children) != 3 || children[0].Code != "nested-code" ||
		children[1].Mode != schema.PermissionRequireModeAll || children[2].Check.CheckName != "nested" {
		t.Fatalf("nested group was not ordered after expansion: %+v", children)
	}
	if expanded := children[1].Children; len(expanded) != 2 || expanded[0].Code != "external.File:read" || expanded[1].Check.CheckName != "owner" {
		t.Fatalf("unresolved source term lost its grouping: %+v", expanded)
	}
	if declared.Children[0] != first || declared.Children[1] != group || declared.Children[3] != last ||
		group.Children[0] != nested || group.Children[1] != unresolved || unresolved.Mode != "" {
		t.Fatal("composition changed the declared order or unresolved source term")
	}
}
