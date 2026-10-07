package diff

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func _testServiceRules(t *testing.T, coverage *_RuleCoverage) {
	t.Helper()
	t.Run("api boundary", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareService("owner", &schema.Service{}, &schema.Service{Api: true})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{"service.api.changed": ImpactBreaking})
	})

	t.Run("arguments", func(t *testing.T) {
		prefixes := []string{"resource.check.argument", "method.argument", "task.trigger.argument"}
		for _, prefix := range prefixes {
			t.Run(prefix, func(t *testing.T) {
				changes := diffChanges(func(diff *_Diff) {
					diff.compareArguments("owner", prefix,
						[]*schema.Argument{{Name: "same", Type: scalarType("string"), Example: "old"}},
						[]*schema.Argument{{Name: "same", Type: scalarType("int"), Sensitive: true, Example: "new"}})
					diff.compareArguments("owner", prefix,
						[]*schema.Argument{{Name: "old", Type: scalarType("string")}},
						[]*schema.Argument{{Name: "new", Type: scalarType("string")}})
					diff.compareArguments("owner", prefix,
						[]*schema.Argument{{Name: "first", Type: scalarType("string")}, {Name: "second", Type: scalarType("string")}},
						[]*schema.Argument{{Name: "second", Type: scalarType("string")}, {Name: "first", Type: scalarType("string")}})
				})
				coverage.assert(t, changes, map[string]ImpactLevel{
					prefix + ".removed":           ImpactBreaking,
					prefix + ".type.changed":      ImpactBreaking,
					prefix + ".sensitive.changed": ImpactDangerous,
					prefix + ".example.changed":   ImpactCompatible,
					prefix + ".added":             ImpactBreaking,
					prefix + ".order.changed":     ImpactBreaking,
				})
			})
		}
	})

	t.Run("audiences", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			baseline := []*schema.ActorAudience{{Actor: "Old"}}
			candidate := []*schema.ActorAudience{{Actor: "New"}}
			diff.compareAudiences("owner", "service.audience", baseline, candidate)
			diff.compareAudiences("owner", "web.audience", baseline, candidate)
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.audience.removed": ImpactBreaking,
			"service.audience.added":   ImpactCompatible,
			"web.audience.removed":     ImpactBreaking,
			"web.audience.added":       ImpactCompatible,
		})
	})

	t.Run("authentication", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			for _, prefix := range []string{"service", "method"} {
				diff.compareAuth("owner", prefix, schema.AuthModeRequired, schema.AuthModeAnonymous, schema.Position{}, schema.Position{})
				diff.compareAuth("owner", prefix, schema.AuthModeOptional, schema.AuthModeRequired, schema.Position{}, schema.Position{})
				diff.compareAuth("owner", prefix, schema.AuthModeRequired, schema.AuthModeOptional, schema.Position{}, schema.Position{})
			}
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.auth.changed":   ImpactBreaking,
			"service.auth.tightened": ImpactBreaking,
			"service.auth.relaxed":   ImpactDangerous,
			"method.auth.changed":    ImpactBreaking,
			"method.auth.tightened":  ImpactBreaking,
			"method.auth.relaxed":    ImpactDangerous,
		})
	})

	t.Run("permission requirements", func(t *testing.T) {
		read := &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "read"}}
		write := &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "write"}}
		changes := diffChanges(func(diff *_Diff) {
			for _, prefix := range []string{"service", "method"} {
				diff.compareRequirement("owner", prefix, nil, read, schema.Position{}, schema.Position{})
				diff.compareRequirement("owner", prefix, read, nil, schema.Position{}, schema.Position{})
				diff.compareRequirement("owner", prefix, read, write, schema.Position{}, schema.Position{})
			}
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.require.added":   ImpactBreaking,
			"service.require.removed": ImpactDangerous,
			"service.require.changed": ImpactDangerous,
			"method.require.added":    ImpactBreaking,
			"method.require.removed":  ImpactDangerous,
			"method.require.changed":  ImpactDangerous,
		})
	})

	t.Run("service methods", func(t *testing.T) {
		method := func(name string) *schema.Method {
			return &schema.Method{Name: name, SkelName: name, AuthMode: schema.AuthModeUnset, Arguments: []*schema.Argument{}}
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareService("Users",
				&schema.Service{Audiences: []*schema.ActorAudience{}, AuthMode: schema.AuthModeUnset, Methods: []*schema.Method{method("old")}},
				&schema.Service{Audiences: []*schema.ActorAudience{}, AuthMode: schema.AuthModeUnset, Methods: []*schema.Method{method("new")}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.method.removed": ImpactBreaking,
			"service.method.added":   ImpactCompatible,
		})
	})

	t.Run("method body", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareMethod("Users.get",
				&schema.Method{AuthMode: schema.AuthModeUnset, Arguments: []*schema.Argument{}, ResultType: scalarType("string")},
				&schema.Method{AuthMode: schema.AuthModeUnset, Arguments: []*schema.Argument{}, ResultType: scalarType("int"), ArgumentsSensitive: true, Example: "new"})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"method.result.changed":        ImpactBreaking,
			"method.sensitive.changed":     ImpactDangerous,
			"method.documentation.changed": ImpactCompatible,
		})
	})
}

func TestExtServiceCompatibility(t *testing.T) {
	for _, ext := range []bool{false, true} {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareService("owner", &schema.Service{Ext: ext}, &schema.Service{Ext: !ext})
		})
		expected := ImpactBreaking
		coverage := &_RuleCoverage{covered: map[string]ImpactLevel{}}
		coverage.assert(t, changes, map[string]ImpactLevel{"service.ext.changed": expected})
	}
}

func TestAuthModeTransitionClassification(t *testing.T) {
	for _, prefix := range []string{"service", "method", "web"} {
		for _, test := range []struct {
			before, after schema.AuthMode
			change        string
			impact        ImpactLevel
		}{
			{schema.AuthModeRequired, schema.AuthModeAnonymous, "changed", ImpactBreaking},
			{schema.AuthModeAnonymous, schema.AuthModeRequired, "changed", ImpactBreaking},
			{schema.AuthModeOptional, schema.AuthModeRequired, "tightened", ImpactBreaking},
			{schema.AuthModeOptional, schema.AuthModeAnonymous, "tightened", ImpactBreaking},
			{schema.AuthModeRequired, schema.AuthModeOptional, "relaxed", ImpactDangerous},
			{schema.AuthModeAnonymous, schema.AuthModeOptional, "relaxed", ImpactDangerous},
			{schema.AuthModeUnset, schema.AuthModeRequired, "changed", ImpactDangerous},
			{schema.AuthModeRequired, schema.AuthModeUnset, "changed", ImpactDangerous},
		} {
			changes := diffChanges(func(diff *_Diff) {
				diff.compareAuth("owner", prefix, test.before, test.after, schema.Position{}, schema.Position{})
			})
			impact := test.impact
			if prefix == "service" && (test.before == schema.AuthModeUnset || test.after == schema.AuthModeUnset) {
				impact = ImpactCompatible
			}
			if len(changes) != 1 || changes[0].Code != prefix+".auth."+test.change || changes[0].Impact != impact {
				t.Fatalf("%s %s -> %s: %+v", prefix, test.before, test.after, changes)
			}
		}
	}
}

func TestAuthModeDefaultMigrationClassification(t *testing.T) {
	for _, legacy := range []schema.AuthMode{"", schema.AuthModeUnset} {
		for _, test := range []struct {
			declaration string
			canonical   schema.AuthMode
			impact      ImpactLevel
		}{
			{"method", schema.AuthModeInherit, ImpactCompatible},
			{"service", schema.AuthModeRequired, ImpactCompatible},
			{"web", schema.AuthModeRequired, ImpactDangerous},
			{"web", schema.AuthModeOff, ImpactDangerous},
			{"web", schema.AuthModeOptional, ImpactDangerous},
			{"method", schema.AuthModeRequired, ImpactDangerous},
		} {
			changes := diffChanges(func(diff *_Diff) {
				diff.compareAuth("owner", test.declaration, legacy, test.canonical, schema.Position{}, schema.Position{})
			})
			if len(changes) != 1 || changes[0].Impact != test.impact {
				t.Fatalf("%s %q -> %s: %+v", test.declaration, legacy, test.canonical, changes)
			}
		}
	}
}

func TestDiffEffectiveAuthentication(t *testing.T) {
	for _, test := range []struct {
		name                                                   string
		beforeService, afterService, beforeMethod, afterMethod schema.AuthMode
		breaking                                               bool
	}{
		{"explicit equals inherited", schema.AuthModeRequired, schema.AuthModeRequired, schema.AuthModeInherit, schema.AuthModeRequired, false},
		{"legacy equals inherited", schema.AuthModeRequired, schema.AuthModeRequired, schema.AuthModeUnset, schema.AuthModeInherit, false},
		{"inherited tightening", schema.AuthModeOptional, schema.AuthModeRequired, schema.AuthModeInherit, schema.AuthModeInherit, true},
		{"explicit method unaffected", schema.AuthModeOptional, schema.AuthModeRequired, schema.AuthModeOptional, schema.AuthModeOptional, false},
		{"preserve old default explicitly", schema.AuthModeOptional, schema.AuthModeRequired, schema.AuthModeInherit, schema.AuthModeOptional, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			makeDomain := func(serviceAuth, methodAuth schema.AuthMode) *schema.Domain {
				service := serviceDeclaration("Service", "read")
				service.Service.AuthMode = serviceAuth
				service.Service.Methods[0].AuthMode = methodAuth
				return newTestDomain(service)
			}
			report, err := Compare(makeDomain(test.beforeService, test.beforeMethod), makeDomain(test.afterService, test.afterMethod))
			if err != nil {
				t.Fatal(err)
			}
			if report.Compatible == test.breaking {
				t.Fatalf("unexpected report: %+v", report)
			}
			if !test.breaking && report.Summary.Dangerous != 0 {
				t.Fatalf("equivalent auth marked dangerous: %+v", report)
			}
		})
	}
}

func TestAddedExtensionMethodRemainsCompatible(t *testing.T) {
	before, after := serviceDeclaration("Extension", "old"), serviceDeclaration("Extension", "old", "new")
	before.Service.Ext, after.Service.Ext = true, true
	report, err := Compare(newTestDomain(before), newTestDomain(after))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 1 || report.Changes[0].Impact != ImpactCompatible {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffDuplicatePermissionRequirement(t *testing.T) {
	before, after := serviceDeclaration("Service", "read"), serviceDeclaration("Service", "read")
	before.Service.Require, after.Service.Require = &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "read"}}, &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "read"}}
	after.Service.Methods[0].Require = &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "read"}}
	report, err := Compare(newTestDomain(before), newTestDomain(after))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 1 || report.Changes[0].Impact != ImpactCompatible {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffIgnoresCachedEffectivePolicies(t *testing.T) {
	before, after := serviceDeclaration("Service", "read"), serviceDeclaration("Service", "read")
	baseline, candidate := newTestDomain(before), newTestDomain(after)
	after.Service.Methods[0].EffectiveAuthMode = schema.AuthModeAnonymous
	after.Service.Methods[0].EffectiveRequire = new(schema.PermissionRequire{Expression: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "extra"})})
	report, err := Compare(baseline, candidate)
	if err != nil || len(report.Changes) != 0 {
		t.Fatalf("derived values affected diff: %+v, %v", report, err)
	}
	before.Service.AuthMode = schema.AuthModeOptional
	after.Service.AuthMode = schema.AuthModeRequired
	report, err = Compare(baseline, candidate)
	if err != nil || report.Compatible {
		t.Fatalf("stale derived values hid declared tightening: %+v, %v", report, err)
	}
}

func TestDiffEffectivePermissionConjunctions(t *testing.T) {
	p := &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "read"}}
	q := &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "write"}}
	r := &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "admin"}}
	all := func(children ...*schema.PermissionRequire) *schema.PermissionRequire {
		return &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeAll, Children: requirementExpressions(children)}}
	}
	any := func(children ...*schema.PermissionRequire) *schema.PermissionRequire {
		return &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: schema.PermissionRequireModeAny, Children: requirementExpressions(children)}}
	}
	for _, test := range []struct {
		name                                                   string
		beforeService, beforeMethod, afterService, afterMethod *schema.PermissionRequire
		impact                                                 ImpactLevel
	}{
		{"service tightened", p, nil, all(p, q), nil, ImpactBreaking},
		{"method tightened", nil, p, nil, all(p, q), ImpactBreaking},
		{"nested conjunction tightened", all(p, q), nil, all(q, all(r, p, p)), nil, ImpactBreaking},
		{"relaxed", all(p, q), nil, p, nil, ImpactDangerous},
		{"reordered and duplicated", all(p, q), nil, all(q, p, p), nil, ImpactCompatible},
		{"replacement", p, nil, q, nil, ImpactDangerous},
		{"disjunction remains opaque", p, nil, all(p, any(p, q)), nil, ImpactDangerous},
		{"disjunction replaced", any(p, q), nil, p, nil, ImpactDangerous},
		{"already enforced by method", p, q, all(p, q), q, ImpactCompatible},
		{"already enforced by service", q, p, q, all(p, q), ImpactCompatible},
		{"moved between scopes", p, q, all(p, q), nil, ImpactCompatible},
		{"additional requirement across scopes", p, q, all(p, r), q, ImpactBreaking},
	} {
		t.Run(test.name, func(t *testing.T) {
			makeDomain := func(servicePolicy, methodPolicy *schema.PermissionRequire) *schema.Domain {
				service := serviceDeclaration("Service", "read")
				service.Service.Require = servicePolicy
				service.Service.Methods[0].Require = methodPolicy
				return newTestDomain(service)
			}
			report, err := Compare(makeDomain(test.beforeService, test.beforeMethod), makeDomain(test.afterService, test.afterMethod))
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Changes) == 0 || report.Compatible != (test.impact != ImpactBreaking) {
				t.Fatalf("unexpected report: %+v", report)
			}
			for _, change := range report.Changes {
				if change.Impact != test.impact {
					t.Fatalf("unexpected change: %+v", change)
				}
			}
		})
	}
}
