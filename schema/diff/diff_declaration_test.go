package diff

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestEventDirectionChangesAreBreaking(t *testing.T) {
	for _, ext := range []bool{false, true} {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareData("demo.audit.AuditRecordedEvent", &schema.Data{Ext: ext}, &schema.Data{Ext: !ext})
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
			baseline  *schema.Domain
			candidate *schema.Domain
			code      string
			impact    ImpactLevel
		}{
			{"domain name replaces nested changes", namedTestDomain("baseline", dataDeclaration("State")), namedTestDomain("candidate", enumDeclaration("State")), "domain.name.changed", ImpactBreaking},
			{"domain description", describedTestDomain("old"), describedTestDomain("new"), "domain.description.changed", ImpactCompatible},
			{"declaration removed", newTestDomain(dataDeclaration("User")), newTestDomain(), "declaration.removed", ImpactBreaking},
			{"declaration added", newTestDomain(), newTestDomain(dataDeclaration("User")), "declaration.added", ImpactCompatible},
			{"declaration type", newTestDomain(dataDeclaration("State")), newTestDomain(enumDeclaration("State")), "declaration.type.changed", ImpactBreaking},
			{"visibility increased", visibilityDomain(false), visibilityDomain(true), "declaration.visibility.increased", ImpactCompatible},
			{"visibility reduced", visibilityDomain(true), visibilityDomain(false), "declaration.visibility.reduced", ImpactBreaking},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				report, err := Compare(test.baseline, test.candidate)
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
					diff.compareMetadata(prefix, "symbol", _Metadata{}, _Metadata{
						Description: "new", Deprecated: true, DeprecatedReason: "reason",
					}, schema.Position{}, schema.Position{})
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
			{"data.member", ImpactCompatible},
			{"actor.auth-credential.member", ImpactCompatible},
			{"actor.auth-info.member", ImpactCompatible},
		}
		for _, test := range tests {
			t.Run(test.prefix, func(t *testing.T) {
				changes := diffChanges(func(diff *_Diff) {
					diff.compareMembers("owner", test.prefix,
						[]*schema.DataMember{{Name: "same", Type: scalarType("string"), Example: "old"}},
						[]*schema.DataMember{{Name: "same", Type: scalarType("int"), Sensitive: true, Example: "new"}},
						test.reorderImpact)
					diff.compareMembers("owner", test.prefix,
						[]*schema.DataMember{{Name: "old", Type: scalarType("string")}},
						[]*schema.DataMember{{Name: "new", Type: scalarType("string")}},
						test.reorderImpact)
					diff.compareMembers("owner", test.prefix,
						[]*schema.DataMember{{Name: "first", Type: scalarType("string")}, {Name: "second", Type: scalarType("string")}},
						[]*schema.DataMember{{Name: "second", Type: scalarType("string")}, {Name: "first", Type: scalarType("string")}},
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
			diff.compareEnum("State", &schema.Enum{Items: []*schema.EnumItem{{Name: "OLD"}}}, &schema.Enum{Items: []*schema.EnumItem{{Name: "NEW"}}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"enum.item.removed": ImpactBreaking,
			"enum.item.added":   ImpactDangerous,
		})
	})

	t.Run("data", func(t *testing.T) {
		changes := diffChanges(func(diff *_Diff) {
			diff.compareData("Runtime",
				&schema.Data{Lifecycle: schema.ConfigLifecycleEternal, TypeParameters: []*schema.TypeParameter{{Name: "T"}}, Members: []*schema.DataMember{}},
				&schema.Data{Lifecycle: schema.ConfigLifecycleInstant, Sensitive: true, TypeParameters: []*schema.TypeParameter{{Name: "U"}}, Members: []*schema.DataMember{}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"config.lifecycle.changed":     ImpactDangerous,
			"data.sensitive.changed":       ImpactDangerous,
			"data.type-parameters.changed": ImpactBreaking,
		})
	})

	t.Run("actor", func(t *testing.T) {
		auth := func(enabled, permission bool, vias ...string) *schema.Actor {
			result := &schema.Actor{AuthEnabled: enabled, PermissionEnabled: permission, Vias: []*schema.ActorVia{}}
			for _, via := range vias {
				result.Vias = append(result.Vias, &schema.ActorVia{Name: via})
			}
			if enabled {
				result.AuthCredential = &schema.Data{Members: []*schema.DataMember{}}
				result.AuthInfo = &schema.Data{Members: []*schema.DataMember{}}
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
			"actor.auth.removed":       ImpactBreaking,
			"actor.permission.added":   ImpactDangerous,
			"actor.permission.removed": ImpactBreaking,
		})
	})
}

func _testResourceRules(t *testing.T, coverage *_RuleCoverage) {
	t.Helper()
	t.Run("resource", func(t *testing.T) {
		check := func(name string) *schema.ResourceCheck {
			return &schema.ResourceCheck{Name: name, Method: &schema.Method{Arguments: []*schema.Argument{}}}
		}
		action := func(name, code string) *schema.ResourceAction {
			return &schema.ResourceAction{Name: name, PermissionCode: code, Checks: []*schema.ResourceCheck{}}
		}
		changes := diffChanges(func(diff *_Diff) {
			diff.compareResource("User",
				&schema.Resource{Checks: []*schema.ResourceCheck{check("old")}, Actions: []*schema.ResourceAction{action("old", "old"), action("same", "old")}},
				&schema.Resource{Checks: []*schema.ResourceCheck{check("new")}, Actions: []*schema.ResourceAction{action("same", "new"), action("new", "new")}})
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
		trigger := func(name string) *schema.TaskTrigger {
			return &schema.TaskTrigger{Name: name, SkelName: name, Arguments: []*schema.Argument{}}
		}
		baselineSame := trigger("same")
		candidateSame := trigger("same")
		candidateSame.ArgumentsSensitive = true
		candidateSame.InputDescription = "new"
		changes := diffChanges(func(diff *_Diff) {
			diff.compareTask("Jobs",
				&schema.Task{Triggers: []*schema.TaskTrigger{trigger("old"), baselineSame}},
				&schema.Task{Triggers: []*schema.TaskTrigger{candidateSame, trigger("new")}})
		})
		coverage.assert(t, changes, map[string]ImpactLevel{
			"task.trigger.removed":               ImpactBreaking,
			"task.trigger.added":                 ImpactCompatible,
			"task.trigger.sensitive.changed":     ImpactDangerous,
			"task.trigger.documentation.changed": ImpactCompatible,
		})
	})
}

func namedTestDomain(domain string, declaration *schema.Declaration) *schema.Domain {
	document := newTestDomain(declaration)
	document = schema.NewDomainFromSpec(schema.DomainSpec{Name: domain, Description: domain + " description", Data: document.Data(), Enums: document.Enums()})
	return document
}

func describedTestDomain(description string) *schema.Domain {
	document := newTestDomain()
	document = schema.NewDomainFromSpec(schema.DomainSpec{Name: document.Name(), Description: description})
	return document
}

func visibilityDomain(public bool) *schema.Domain {
	declaration := dataDeclaration("User")
	declaration.Pub = public
	return newTestDomain(declaration)
}

func TestWebMountCompare(t *testing.T) {
	for _, paths := range [][2]string{{"", "/"}, {"", "/portal"}, {"/portal", ""}, {"/portal", "/other"}} {
		baseline := &schema.Declaration{Kind: schema.DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &schema.Web{MountPath: paths[0]}}
		candidate := &schema.Declaration{Kind: schema.DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &schema.Web{MountPath: paths[1]}}
		changes := diffChanges(func(diff *_Diff) { diff.compareDeclaration(baseline, candidate) })
		if len(changes) != 1 || changes[0].Code != "web.mount-path.changed" || changes[0].Impact != ImpactBreaking || changes[0].Change != ChangeModified {
			t.Fatalf("mount diff %q: %+v", paths, changes)
		}
	}
}

func TestWebAuthCompare(t *testing.T) {
	for _, test := range []struct {
		before, after schema.AuthMode
		code          string
	}{
		{schema.AuthModeOptional, schema.AuthModeRequired, "web.auth.tightened"},
		{schema.AuthModeRequired, schema.AuthModeOptional, "web.auth.relaxed"},
		{schema.AuthModeOptional, schema.AuthModeAnonymous, "web.auth.tightened"},
		{schema.AuthModeAnonymous, schema.AuthModeOff, "web.auth.changed"},
	} {
		baseline := &schema.Declaration{Kind: schema.DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &schema.Web{Auth: test.before}}
		candidate := &schema.Declaration{Kind: schema.DeclarationTypeWeb, SkelName: "demo.PortalWeb", Web: &schema.Web{Auth: test.after}}
		changes := diffChanges(func(diff *_Diff) { diff.compareDeclaration(baseline, candidate) })
		impact := ImpactDangerous
		if test.before == schema.AuthModeOptional && test.after != schema.AuthModeOptional {
			impact = ImpactBreaking
		}
		if len(changes) != 1 || changes[0].Code != test.code || changes[0].Impact != impact {
			t.Fatalf("auth diff: %+v", changes)
		}
	}
}

func TestAddedNullableCredentialIsCompatible(t *testing.T) {
	for _, nullable := range []bool{false, true} {
		before, after := actorDomain(true, false), actorDomain(true, false)
		kind := scalarType("string")
		kind.Nullable = nullable
		after.Actors()[0].AuthCredential.Members = []*schema.DataMember{{Name: "extra", Type: kind}}
		report, err := Compare(before, after)
		if err != nil {
			t.Fatal(err)
		}
		if len(report.Changes) != 1 || report.Compatible != nullable {
			t.Fatalf("nullable=%v: %+v", nullable, report)
		}
	}
}

func TestWebOffAuthTransitions(t *testing.T) {
	for _, test := range []struct {
		before, after schema.AuthMode
		impact        ImpactLevel
		code          string
	}{
		{schema.AuthModeOff, schema.AuthModeRequired, ImpactBreaking, "web.auth.tightened"},
		{schema.AuthModeOff, schema.AuthModeOptional, ImpactBreaking, "web.auth.tightened"},
		{schema.AuthModeOff, schema.AuthModeAnonymous, ImpactBreaking, "web.auth.tightened"},
		{schema.AuthModeRequired, schema.AuthModeOff, ImpactDangerous, "web.auth.changed"},
		{schema.AuthModeOptional, schema.AuthModeOff, ImpactDangerous, "web.auth.changed"},
		{schema.AuthModeAnonymous, schema.AuthModeOff, ImpactDangerous, "web.auth.changed"},
		{schema.AuthModeNoAuth, schema.AuthModeRequired, ImpactBreaking, "web.auth.tightened"},
		{schema.AuthModeNoAuth, schema.AuthModeOff, ImpactCompatible, "web.auth.changed"},
	} {
		t.Run(string(test.before)+" to "+string(test.after), func(t *testing.T) {
			makeDomain := func(mode schema.AuthMode) *schema.Domain {
				return newTestDomain(&schema.Declaration{
					Kind: schema.DeclarationTypeWeb, Name: "ExampleWeb", SkelName: "demo.user.ExampleWeb",
					Web: &schema.Web{Auth: mode, Audiences: []*schema.ActorAudience{{Actor: "demo.user.UserActor", Via: "client"}}},
				})
			}
			report, err := Compare(makeDomain(test.before), makeDomain(test.after))
			if err != nil {
				t.Fatal(err)
			}
			if test.before == schema.AuthModeNoAuth && test.after == schema.AuthModeOff {
				if len(report.Changes) != 0 {
					t.Fatalf("equivalent legacy auth changed: %+v", report)
				}
				return
			}
			if len(report.Changes) != 1 || report.Compatible != (test.impact != ImpactBreaking) {
				t.Fatalf("unexpected report: %+v", report)
			}
			if change := report.Changes[0]; change.Impact != test.impact || change.Code != test.code {
				t.Fatalf("unexpected change: %+v", change)
			}
		})
	}
}

func TestActorIdentifierChange(t *testing.T) {
	before, after := actorDomain(true, false), actorDomain(true, false)
	for _, domain := range []*schema.Domain{before, after} {
		domain.Actors()[0].AuthInfo.Members = []*schema.DataMember{{Name: "id", Type: scalarType("string")}}
	}
	after.Actors()[0].IdentifierField = "id"
	for _, pair := range [][2]*schema.Domain{{before, after}, {after, before}} {
		report, err := Compare(pair[0], pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if len(report.Changes) != 1 || report.Changes[0].Code != "actor.identifier.changed" || report.Changes[0].Impact != ImpactDangerous {
			t.Fatalf("lost actor identity change: %+v", report)
		}
	}
}
