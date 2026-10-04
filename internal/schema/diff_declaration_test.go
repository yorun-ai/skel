package schema

import (
	"testing"

	"go.yorun.ai/skelc/internal/model"
)

func TestEventDirectionChangesAreBreaking(t *testing.T) {
	for _, ext := range []bool{false, true} {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareData("demo.audit.AuditRecordedEvent", &DataSchema{Ext: ext}, &DataSchema{Ext: !ext})
		})
		if len(changes) != 1 || changes[0].Code != "event.ext.changed" || changes[0].Impact != ImpactBreaking {
			t.Fatalf("unexpected direction diff: %+v", changes)
		}
	}
}

func _testDeclarationRules(t *testing.T, coverage *_RuleCoverage) {
	t.Helper()
	t.Run("document and declaration", func(t *testing.T) {
		tests := []struct {
			name      string
			baseline  *Document
			candidate *Document
			code      string
			impact    ImpactLevel
		}{
			{"domain name replaces nested changes", namedTestDocument("baseline", dataDeclaration("State")), namedTestDocument("candidate", enumDeclaration("State")), "domain.name.changed", ImpactBreaking},
			{"domain description", describedTestDocument("old"), describedTestDocument("new"), "domain.description.changed", ImpactCompatible},
			{"declaration removed", newTestDocument(dataDeclaration("User")), newTestDocument(), "declaration.removed", ImpactBreaking},
			{"declaration added", newTestDocument(), newTestDocument(dataDeclaration("User")), "declaration.added", ImpactCompatible},
			{"declaration type", newTestDocument(dataDeclaration("State")), newTestDocument(enumDeclaration("State")), "declaration.type.changed", ImpactBreaking},
			{"visibility increased", visibilityDocument(false), visibilityDocument(true), "declaration.visibility.increased", ImpactCompatible},
			{"visibility reduced", visibilityDocument(true), visibilityDocument(false), "declaration.visibility.reduced", ImpactBreaking},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				report, err := Diff(test.baseline, test.candidate)
				if err != nil {
					t.Fatal(err)
				}
				coverage.assert(t, report.Changes, map[string]ImpactLevel{test.code: test.impact})
			})
		}
	})

	t.Run("metadata", func(t *testing.T) {
		prefixes := []string{
			"declaration", "enum.item", "data.member", "actor.auth-credential.member",
			"actor.auth-info.member", "resource.action", "resource.check",
			"resource.check.argument", "method", "method.argument", "task.trigger",
			"task.trigger.argument",
		}
		for _, prefix := range prefixes {
			t.Run(prefix, func(t *testing.T) {
				changes := diffChanges(func(diff *_Diff) {
					diff.compareMetadata(prefix, "symbol", Metadata{}, Metadata{
						Description: "new", Deprecated: true, DeprecatedReason: "reason",
					}, model.Position{}, model.Position{})
				})
				coverage.assert(t, changes, map[string]ImpactLevel{
					prefix + ".description.changed":       ImpactCompatible,
					prefix + ".deprecated.changed":        ImpactCompatible,
					prefix + ".deprecated-reason.changed": ImpactCompatible,
				})
			})
		}
	})

	t.Run("members", func(t *testing.T) {
		tests := []struct {
			prefix        string
			reorderImpact ImpactLevel
		}{
			{"data.member", ImpactDangerous},
			{"actor.auth-credential.member", ImpactDangerous},
			{"actor.auth-info.member", ImpactDangerous},
		}
		for _, test := range tests {
			t.Run(test.prefix, func(t *testing.T) {
				changes := diffChanges(func(diff *_Diff) {
					diff.compareMembers("owner", test.prefix,
						[]*Member{{Name: "same", Type: scalarType("string"), Example: "old"}},
						[]*Member{{Name: "same", Type: scalarType("int"), Sensitive: true, Example: "new"}},
						test.reorderImpact)
					diff.compareMembers("owner", test.prefix,
						[]*Member{{Name: "old", Type: scalarType("string")}},
						[]*Member{{Name: "new", Type: scalarType("string")}},
						test.reorderImpact)
					diff.compareMembers("owner", test.prefix,
						[]*Member{{Name: "first", Type: scalarType("string")}, {Name: "second", Type: scalarType("string")}},
						[]*Member{{Name: "second", Type: scalarType("string")}, {Name: "first", Type: scalarType("string")}},
						test.reorderImpact)
				})
				coverage.assert(t, changes, map[string]ImpactLevel{
					test.prefix + ".removed":           ImpactBreaking,
					test.prefix + ".type.changed":      ImpactBreaking,
					test.prefix + ".sensitive.changed": ImpactDangerous,
					test.prefix + ".example.changed":   ImpactCompatible,
					test.prefix + ".added":             ImpactBreaking,
					test.prefix + ".order.changed":     test.reorderImpact,
				})
			})
		}
	})

	t.Run("enum", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareEnum("State", &EnumSchema{Items: []*EnumItem{{Name: "OLD"}}}, &EnumSchema{Items: []*EnumItem{{Name: "NEW"}}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"enum.item.removed": ImpactBreaking,
			"enum.item.added":   ImpactDangerous,
		})
	})

	t.Run("data", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareData("Runtime",
				&DataSchema{Lifecycle: ConfigLifecycleEternal, TypeParameters: []string{"T"}, Members: []*Member{}},
				&DataSchema{Lifecycle: ConfigLifecycleInstant, Sensitive: true, TypeParameters: []string{"U"}, Members: []*Member{}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"config.lifecycle.changed":     ImpactDangerous,
			"data.sensitive.changed":       ImpactDangerous,
			"data.type-parameters.changed": ImpactBreaking,
		})
	})

	t.Run("actor", func(t *testing.T) {
		auth := func(enabled, permission bool, vias ...string) *ActorSchema {
			result := &ActorSchema{AuthEnabled: enabled, PermEnabled: permission, Vias: []*ActorVia{}}
			for _, via := range vias {
				result.Vias = append(result.Vias, &ActorVia{Name: via})
			}
			if enabled {
				result.AuthCredential = &DataSchema{Members: []*Member{}}
				result.AuthInfo = &DataSchema{Members: []*Member{}}
			}
			return result
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareActor("Caller", auth(false, false, "old"), auth(false, false, "new"))
			diff.compareActor("Caller", auth(false, false), auth(true, true))
			diff.compareActor("Caller", auth(true, true), auth(false, false))
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"actor.via.removed":        ImpactBreaking,
			"actor.via.added":          ImpactCompatible,
			"actor.auth.added":         ImpactDangerous,
			"actor.auth.removed":       ImpactDangerous,
			"actor.permission.added":   ImpactDangerous,
			"actor.permission.removed": ImpactDangerous,
		})
	})
}

func _testResourceRules(t *testing.T, coverage *_RuleCoverage) {
	t.Helper()
	t.Run("resource", func(t *testing.T) {
		check := func(name string) *ResourceCheck { return &ResourceCheck{Name: name, Arguments: []*Argument{}} }
		action := func(name, code string) *ResourceAction {
			return &ResourceAction{Name: name, PermissionCode: code, Checks: []*ResourceCheck{}}
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareResource("User",
				&ResourceSchema{Checks: []*ResourceCheck{check("old")}, Actions: []*ResourceAction{action("old", "old"), action("same", "old")}},
				&ResourceSchema{Checks: []*ResourceCheck{check("new")}, Actions: []*ResourceAction{action("same", "new"), action("new", "new")}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"resource.action.removed":      ImpactBreaking,
			"resource.action.code.changed": ImpactDangerous,
			"resource.action.added":        ImpactCompatible,
			"resource.check.removed":       ImpactBreaking,
			"resource.check.added":         ImpactCompatible,
		})
	})
}

func _testTaskRules(t *testing.T, coverage *_RuleCoverage) {
	t.Helper()
	t.Run("task triggers", func(t *testing.T) {
		trigger := func(name string) *Trigger {
			return &Trigger{Name: name, SkelName: name, Arguments: []*Argument{}}
		}
		baselineSame := trigger("same")
		candidateSame := trigger("same")
		candidateSame.ArgumentsSensitive = true
		candidateSame.InputDescription = "new"
		changes := diffChanges(func(diff *_Diff) {
			diff.compareTask("Jobs",
				&TaskSchema{Triggers: []*Trigger{trigger("old"), baselineSame}},
				&TaskSchema{Triggers: []*Trigger{candidateSame, trigger("new")}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"task.trigger.removed":               ImpactBreaking,
			"task.trigger.added":                 ImpactCompatible,
			"task.trigger.sensitive.changed":     ImpactDangerous,
			"task.trigger.documentation.changed": ImpactCompatible,
		})
	})
}

func namedTestDocument(domain string, declaration *Declaration) *Document {
	document := newTestDocument(declaration)
	document.Domain = domain
	document.Description = domain + " description"
	return document
}

func describedTestDocument(description string) *Document {
	document := newTestDocument()
	document.Description = description
	return document
}

func visibilityDocument(public bool) *Document {
	declaration := dataDeclaration("User")
	declaration.Pub = public
	return newTestDocument(declaration)
}

func TestWebMountDiff(t *testing.T) {
	for _, paths := range [][2]string{{"", "/"}, {"", "/portal"}, {"/portal", ""}, {"/portal", "/other"}} {
		baseline := &Declaration{Kind: DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &WebSchema{MountPath: paths[0]}}
		candidate := &Declaration{Kind: DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &WebSchema{MountPath: paths[1]}}
		changes := diffChanges(func(diff *_Diff) { diff.compareDeclaration(baseline, candidate) })
		if len(changes) != 1 || changes[0].Code != "web.mount-path.changed" || changes[0].Impact != ImpactBreaking || changes[0].Change != ChangeModified {
			t.Fatalf("mount diff %q: %+v", paths, changes)
		}
	}
}

func TestWebAuthDiff(t *testing.T) {
	for _, test := range []struct {
		before, after AuthMode
		code          string
	}{
		{AuthModeOptional, AuthModeRequired, "web.auth.tightened"},
		{AuthModeRequired, AuthModeOptional, "web.auth.relaxed"},
		{AuthModeOptional, AuthModeAnonymous, "web.auth.tightened"},
		{AuthModeAnonymous, AuthModeOff, "web.auth.changed"},
	} {
		baseline := &Declaration{Kind: DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &WebSchema{Auth: test.before}}
		candidate := &Declaration{Kind: DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &WebSchema{Auth: test.after}}
		changes := diffChanges(func(diff *_Diff) { diff.compareDeclaration(baseline, candidate) })
		if len(changes) != 1 || changes[0].Code != test.code || changes[0].Impact != ImpactDangerous {
			t.Fatalf("auth diff: %+v", changes)
		}
	}
}
