package schema

import (
	"testing"

	"go.yorun.ai/skelc/internal/model"
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
				diff.compareAuth("owner", prefix, AuthModeUnset, AuthModeNoAuth, model.Position{}, model.Position{})
				diff.compareAuth("owner", prefix, AuthModeOptional, AuthModeAuth, model.Position{}, model.Position{})
				diff.compareAuth("owner", prefix, AuthModeAuth, AuthModeOptional, model.Position{}, model.Position{})
			}
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.auth.changed":   ImpactDangerous,
			"service.auth.tightened": ImpactDangerous,
			"service.auth.relaxed":   ImpactDangerous,
			"method.auth.changed":    ImpactDangerous,
			"method.auth.tightened":  ImpactDangerous,
			"method.auth.relaxed":    ImpactDangerous,
		})
	})

	t.Run("permission requirements", func(t *testing.T) {
		read := &Requirement{Mode: RequirementModeCode, Code: "read"}
		write := &Requirement{Mode: RequirementModeCode, Code: "write"}
		changes := diffChanges(func(diff *_Diff) {
			for _, prefix := range []string{"service", "method"} {
				diff.compareRequirement("owner", prefix, nil, read, model.Position{}, model.Position{})
				diff.compareRequirement("owner", prefix, read, nil, model.Position{}, model.Position{})
				diff.compareRequirement("owner", prefix, read, write, model.Position{}, model.Position{})
			}
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"service.require.added":   ImpactDangerous,
			"service.require.removed": ImpactDangerous,
			"service.require.changed": ImpactDangerous,
			"method.require.added":    ImpactDangerous,
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
			{AuthModeRequired, AuthModeGuest, "changed", ImpactDangerous},
			{AuthModeGuest, AuthModeRequired, "changed", ImpactDangerous},
			{AuthModeOptional, AuthModeRequired, "tightened", ImpactDangerous},
			{AuthModeOptional, AuthModeGuest, "tightened", ImpactDangerous},
			{AuthModeRequired, AuthModeOptional, "relaxed", ImpactDangerous},
			{AuthModeGuest, AuthModeOptional, "relaxed", ImpactDangerous},
			{AuthModeUnset, AuthModeRequired, "changed", ImpactDangerous},
			{AuthModeRequired, AuthModeUnset, "changed", ImpactDangerous},
			{AuthModeAuth, AuthModeRequired, "changed", ImpactCompatible},
			{AuthModeRequired, AuthModeAuth, "changed", ImpactCompatible},
		} {
			changes := diffChanges(func(diff *_Diff) {
				diff.compareAuth("owner", prefix, test.before, test.after, model.Position{}, model.Position{})
			})
			if len(changes) != 1 || changes[0].Code != prefix+".auth."+test.change || changes[0].Impact != test.impact {
				t.Fatalf("%s %s -> %s: %+v", prefix, test.before, test.after, changes)
			}
		}
		equivalent := AuthModeOptional
		if prefix == "web" {
			equivalent = AuthModeOff
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareAuth("owner", prefix, AuthModeNoAuth, equivalent, model.Position{}, model.Position{})
		})
		if len(changes) != 1 || changes[0].Impact != ImpactCompatible {
			t.Fatalf("%s legacy noauth: %+v", prefix, changes)
		}
	}
}
