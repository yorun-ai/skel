package schema

import (
	"testing"
)

func _testServiceRules(t *testing.T, coverage *_RuleCoverage) {
	t.Helper()
	t.Run("api boundary", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareService("owner", &ServiceSchema{}, &ServiceSchema{Api: true})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{"service.api.changed": ImpactBreaking})
	})

	t.Run("arguments", func(t *testing.T) {
		prefixes := []string{"resource.check.argument", "method.argument", "task.trigger.argument"}
		for _, prefix := range prefixes {
			t.Run(prefix, func(t *testing.T) {
				changes := diffChanges(func(diff *_Diff) {
					diff.compareArguments("owner", prefix,
						[]*Argument{{Name: "same", Type: scalarType("string"), Example: "old"}},
						[]*Argument{{Name: "same", Type: scalarType("int"), Sensitive: true, Example: "new"}})
					diff.compareArguments("owner", prefix,
						[]*Argument{{Name: "old", Type: scalarType("string")}},
						[]*Argument{{Name: "new", Type: scalarType("string")}})
					diff.compareArguments("owner", prefix,
						[]*Argument{{Name: "first", Type: scalarType("string")}, {Name: "second", Type: scalarType("string")}},
						[]*Argument{{Name: "second", Type: scalarType("string")}, {Name: "first", Type: scalarType("string")}})
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
			baseline := []*Audience{{Actor: "Old"}}
			candidate := []*Audience{{Actor: "New"}}
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
				diff.compareAuth("owner", prefix, AuthModeRequired, AuthModeAnonymous, Position{}, Position{})
				diff.compareAuth("owner", prefix, AuthModeOptional, AuthModeAuth, Position{}, Position{})
				diff.compareAuth("owner", prefix, AuthModeAuth, AuthModeOptional, Position{}, Position{})
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
		read := &Requirement{Mode: RequirementModeCode, Code: "read"}
		write := &Requirement{Mode: RequirementModeCode, Code: "write"}
		changes := diffChanges(func(diff *_Diff) {
			for _, prefix := range []string{"service", "method"} {
				diff.compareRequirement("owner", prefix, nil, read, Position{}, Position{})
				diff.compareRequirement("owner", prefix, read, nil, Position{}, Position{})
				diff.compareRequirement("owner", prefix, read, write, Position{}, Position{})
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
		method := func(name string) *Method {
			return &Method{Name: name, SkelName: name, Auth: AuthModeUnset, Arguments: []*Argument{}}
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareService("Users",
				&ServiceSchema{Audiences: []*Audience{}, Auth: AuthModeUnset, Methods: []*Method{method("old")}},
				&ServiceSchema{Audiences: []*Audience{}, Auth: AuthModeUnset, Methods: []*Method{method("new")}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.method.removed": ImpactBreaking,
			"service.method.added":   ImpactCompatible,
		})
	})

	t.Run("method body", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareMethod("Users.get",
				&Method{Auth: AuthModeUnset, Arguments: []*Argument{}, Result: scalarType("string")},
				&Method{Auth: AuthModeUnset, Arguments: []*Argument{}, Result: scalarType("int"), ArgumentsSensitive: true, Example: "new"})
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
			diff.compareService("owner", &ServiceSchema{Ext: ext}, &ServiceSchema{Ext: !ext})
		})
		expected := ImpactBreaking
		coverage := &_RuleCoverage{covered: map[string]ImpactLevel{}}
		coverage.assert(t, changes, map[string]ImpactLevel{"service.ext.changed": expected})
	}
}

func TestAuthModeTransitionClassification(t *testing.T) {
	for _, prefix := range []string{"service", "method", "web"} {
		for _, test := range []struct {
			before, after AuthMode
			change        string
			impact        ImpactLevel
		}{
			{AuthModeRequired, AuthModeAnonymous, "changed", ImpactBreaking},
			{AuthModeAnonymous, AuthModeRequired, "changed", ImpactBreaking},
			{AuthModeOptional, AuthModeRequired, "tightened", ImpactBreaking},
			{AuthModeOptional, AuthModeAnonymous, "tightened", ImpactBreaking},
			{AuthModeRequired, AuthModeOptional, "relaxed", ImpactDangerous},
			{AuthModeAnonymous, AuthModeOptional, "relaxed", ImpactDangerous},
			{AuthModeUnset, AuthModeRequired, "changed", ImpactDangerous},
			{AuthModeRequired, AuthModeUnset, "changed", ImpactDangerous},
			{AuthModeAuth, AuthModeRequired, "changed", ImpactCompatible},
			{AuthModeRequired, AuthModeAuth, "changed", ImpactCompatible},
		} {
			changes := diffChanges(func(diff *_Diff) {
				diff.compareAuth("owner", prefix, test.before, test.after, Position{}, Position{})
			})
			impact := test.impact
			if prefix == "service" && (test.before == AuthModeUnset || test.after == AuthModeUnset) {
				impact = ImpactCompatible
			}
			if len(changes) != 1 || changes[0].Code != prefix+".auth."+test.change || changes[0].Impact != impact {
				t.Fatalf("%s %s -> %s: %+v", prefix, test.before, test.after, changes)
			}
		}
		equivalent := AuthModeOptional
		if prefix == "web" {
			equivalent = AuthModeOff
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareAuth("owner", prefix, AuthModeNoAuth, equivalent, Position{}, Position{})
		})
		if len(changes) != 1 || changes[0].Impact != ImpactCompatible {
			t.Fatalf("%s legacy noauth: %+v", prefix, changes)
		}
	}
}

func TestAuthModeDefaultMigrationClassification(t *testing.T) {
	for _, legacy := range []AuthMode{"", AuthModeUnset} {
		for _, test := range []struct {
			declaration string
			canonical   AuthMode
			impact      ImpactLevel
		}{
			{"method", AuthModeInherit, ImpactCompatible},
			{"service", AuthModeRequired, ImpactCompatible},
			{"web", AuthModeRequired, ImpactDangerous},
			{"web", AuthModeOff, ImpactDangerous},
			{"web", AuthModeOptional, ImpactDangerous},
			{"method", AuthModeRequired, ImpactDangerous},
		} {
			changes := diffChanges(func(diff *_Diff) {
				diff.compareAuth("owner", test.declaration, legacy, test.canonical, Position{}, Position{})
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
		beforeService, afterService, beforeMethod, afterMethod AuthMode
		breaking                                               bool
	}{
		{"explicit equals inherited", AuthModeRequired, AuthModeRequired, AuthModeInherit, AuthModeRequired, false},
		{"legacy equals inherited", AuthModeRequired, AuthModeRequired, AuthModeUnset, AuthModeInherit, false},
		{"inherited tightening", AuthModeOptional, AuthModeRequired, AuthModeInherit, AuthModeInherit, true},
		{"explicit method unaffected", AuthModeOptional, AuthModeRequired, AuthModeOptional, AuthModeOptional, false},
		{"preserve old default explicitly", AuthModeOptional, AuthModeRequired, AuthModeInherit, AuthModeOptional, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			makeDocument := func(serviceAuth, methodAuth AuthMode) *Document {
				service := serviceDeclaration("Service", "read")
				service.Service.Auth = serviceAuth
				service.Service.Methods[0].Auth = methodAuth
				return newTestDocument(service)
			}
			report, err := Diff(makeDocument(test.beforeService, test.beforeMethod), makeDocument(test.afterService, test.afterMethod))
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
	report, err := Diff(newTestDocument(before), newTestDocument(after))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 1 || report.Changes[0].Impact != ImpactCompatible {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffDuplicatePermissionRequirement(t *testing.T) {
	before, after := serviceDeclaration("Service", "read"), serviceDeclaration("Service", "read")
	before.Service.Require, after.Service.Require = &Requirement{Mode: RequirementModeCode, Code: "read"}, &Requirement{Mode: RequirementModeCode, Code: "read"}
	after.Service.Methods[0].Require = &Requirement{Mode: RequirementModeCode, Code: "read"}
	report, err := Diff(newTestDocument(before), newTestDocument(after))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 1 || report.Changes[0].Impact != ImpactCompatible {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffEffectivePermissionConjunctions(t *testing.T) {
	p := &Requirement{Mode: RequirementModeCode, Code: "read"}
	q := &Requirement{Mode: RequirementModeCode, Code: "write"}
	r := &Requirement{Mode: RequirementModeCode, Code: "admin"}
	all := func(children ...*Requirement) *Requirement {
		return &Requirement{Mode: RequirementModeAll, Children: children}
	}
	any := func(children ...*Requirement) *Requirement {
		return &Requirement{Mode: RequirementModeAny, Children: children}
	}
	for _, test := range []struct {
		name                                                   string
		beforeService, beforeMethod, afterService, afterMethod *Requirement
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
			makeDocument := func(servicePolicy, methodPolicy *Requirement) *Document {
				service := serviceDeclaration("Service", "read")
				service.Service.Require = servicePolicy
				service.Service.Methods[0].Require = methodPolicy
				return newTestDocument(service)
			}
			report, err := Diff(makeDocument(test.beforeService, test.beforeMethod), makeDocument(test.afterService, test.afterMethod))
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
