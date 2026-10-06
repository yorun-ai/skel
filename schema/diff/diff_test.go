package diff

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestStableChangeRuleMatrix(t *testing.T) {
	coverage := &_RuleCoverage{covered: map[string]ImpactLevel{}}
	_testDeclarationRules(t, coverage)
	_testResourceRules(t, coverage)
	_testServiceRules(t, coverage)
	_testTaskRules(t, coverage)
	if len(coverage.covered) != 123 {
		t.Fatalf("stable change rule matrix covers %d codes, expected 123", len(coverage.covered))
	}
}

func TestDiffClassifiesAndOrdersChanges(t *testing.T) {
	baseline := newTestDomain(
		dataDeclaration("User", "id"),
		enumDeclaration("UserStatus", "ACTIVE"),
		serviceDeclaration("UserService", "getUser"),
	)
	candidate := newTestDomain(
		dataDeclaration("User", "id", "name"),
		enumDeclaration("UserStatus", "ACTIVE", "DISABLED"),
		serviceDeclaration("UserService", "listUsers"),
	)
	report, err := Compare(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if report.Compatible {
		t.Fatal("expected report to be incompatible")
	}
	if report.Summary != (Summary{Breaking: 2, Dangerous: 1, Compatible: 1}) {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	codes := make([]string, 0, len(report.Changes))
	for _, change := range report.Changes {
		codes = append(codes, change.Code)
	}
	expected := []string{
		"data.member.added",
		"service.method.removed",
		"enum.item.added",
		"service.method.added",
	}
	if strings.Join(codes, ",") != strings.Join(expected, ",") {
		t.Fatalf("unexpected changes: %v", codes)
	}
	if report.Changes[0].Change != ChangeAdded || report.Changes[1].Change != ChangeRemoved ||
		report.Changes[2].Change != ChangeAdded || report.Changes[3].Change != ChangeAdded {
		t.Fatalf("unexpected change kinds: %+v", report.Changes)
	}
}

func TestDiffTreatsDocumentationAsCompatible(t *testing.T) {
	baseline := newTestDomain(dataDeclaration("User", "id"))
	candidate := newTestDomain(dataDeclaration("User", "id"))
	candidate.Data()[0].Description = "User data"
	report, err := Compare(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Compatible || report.Summary.Compatible != 1 || report.Changes[0].Change != ChangeModified ||
		report.Changes[0].Code != "declaration.description.changed" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffTreatsDomainDescriptionAsCompatible(t *testing.T) {
	baseline := describedTestDomain("Baseline domain")
	candidate := describedTestDomain("Candidate domain")
	report, err := Compare(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Compatible || report.Summary != (Summary{Compatible: 1}) || len(report.Changes) != 1 ||
		report.Changes[0].Change != ChangeModified || report.Changes[0].Impact != ImpactCompatible ||
		report.Changes[0].Code != "domain.description.changed" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffTreatsConfigLifecycleAsDangerous(t *testing.T) {
	baselineConfig := dataDeclaration("Runtime", "endpoint")
	baselineConfig.Kind = "config"
	baselineConfig.Data.Lifecycle = "eternal"
	candidateConfig := dataDeclaration("Runtime", "endpoint")
	candidateConfig.Kind = "config"
	candidateConfig.Data.Lifecycle = "instant"
	report, err := Compare(newTestDomain(baselineConfig), newTestDomain(candidateConfig))
	if err != nil {
		t.Fatal(err)
	}
	if !report.Compatible || report.Summary != (Summary{Dangerous: 1}) || len(report.Changes) != 1 ||
		report.Changes[0].Change != ChangeModified || report.Changes[0].Impact != ImpactDangerous ||
		report.Changes[0].Code != "config.lifecycle.changed" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffRecognizesDeclarationTypeChangeWithoutNamespaceCollision(t *testing.T) {
	baseline := newTestDomain(dataDeclaration("State", "value"))
	candidate := newTestDomain(enumDeclaration("State", "ACTIVE"))
	report, err := Compare(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Breaking != 1 || len(report.Changes) != 1 || report.Changes[0].Code != "declaration.type.changed" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffAuthenticationAndPermissionImpacts(t *testing.T) {
	readRequirement := func() *schema.PermissionRequire {
		return &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: "code", Code: "identity.User:read"}}
	}
	writeRequirement := func() *schema.PermissionRequire {
		return &schema.PermissionRequire{Expression: &schema.PermissionExpression{Mode: "code", Code: "identity.User:write"}}
	}
	tests := []struct {
		name      string
		baseline  *schema.Domain
		candidate *schema.Domain
		code      string
	}{
		{"actor authentication added", actorDomain(false, false), actorDomain(true, false), "actor.auth.added"},
		{"actor authentication removed", actorDomain(true, false), actorDomain(false, false), "actor.auth.removed"},
		{"actor permission added", actorDomain(false, false), actorDomain(false, true), "actor.permission.added"},
		{"actor permission removed", actorDomain(false, true), actorDomain(false, false), "actor.permission.removed"},
		{"resource permission code changed", resourceDomain("identity.User:read"), resourceDomain("identity.User:write"), "resource.action.code.changed"},
		{"service authentication tightened", servicePolicyDomain("optional", nil), servicePolicyDomain("required", nil), "service.auth.tightened"},
		{"service authentication relaxed", servicePolicyDomain("required", nil), servicePolicyDomain("optional", nil), "service.auth.relaxed"},
		{"service permission added", servicePolicyDomain("unset", nil), servicePolicyDomain("unset", readRequirement()), "service.require.added"},
		{"service permission changed", servicePolicyDomain("unset", readRequirement()), servicePolicyDomain("unset", writeRequirement()), "service.require.changed"},
		{"service permission removed", servicePolicyDomain("unset", readRequirement()), servicePolicyDomain("unset", nil), "service.require.removed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, err := Compare(test.baseline, test.candidate)
			if err != nil {
				t.Fatal(err)
			}
			impact := ImpactDangerous
			switch test.code {
			case "actor.auth.removed", "actor.permission.removed", "service.auth.tightened", "service.require.added":
				impact = ImpactBreaking
			}
			if report.Compatible != (impact != ImpactBreaking) || len(report.Changes) != 1 ||
				report.Changes[0].Impact != impact || report.Changes[0].Code != test.code {
				t.Fatalf("unexpected report: %+v", report)
			}
		})
	}
}

func TestDiffApiBoundaryIsBreaking(t *testing.T) {
	baseline := newTestDomain(serviceDeclaration("OrderService", "get"))
	candidate := newTestDomain(serviceDeclaration("OrderService", "get"))
	baseline.Services()[0].Pub = false
	candidate.Services()[0].Pub = false
	candidate.Services()[0].Api = true
	report, err := Compare(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if report.Compatible || len(report.Changes) != 1 || report.Changes[0].Code != "service.api.changed" || report.Changes[0].Impact != ImpactBreaking {
		t.Fatalf("unexpected report: %+v", report)
	}
}
