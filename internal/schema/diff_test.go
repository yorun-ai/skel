package schema

import (
	"strings"
	"testing"
)

func TestStableChangeRuleMatrix(t *testing.T) {
	coverage := &_RuleCoverage{covered: map[string]ImpactLevel{}}
	_testDeclarationRules(t, coverage)
	_testResourceRules(t, coverage)
	_testServiceRules(t, coverage)
	_testTaskRules(t, coverage)
	if len(coverage.covered) != 121 {
		t.Fatalf("stable change rule matrix covers %d codes, expected 121", len(coverage.covered))
	}
}

func TestDiffClassifiesAndOrdersChanges(t *testing.T) {
	baseline := newTestDocument(
		dataDeclaration("User", "id"),
		enumDeclaration("UserStatus", "ACTIVE"),
		serviceDeclaration("UserService", "getUser"),
	)
	candidate := newTestDocument(
		dataDeclaration("User", "id", "name"),
		enumDeclaration("UserStatus", "ACTIVE", "DISABLED"),
		serviceDeclaration("UserService", "listUsers"),
	)
	report, err := Diff(baseline, candidate)
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
	baseline := newTestDocument(dataDeclaration("User", "id"))
	candidate := newTestDocument(dataDeclaration("User", "id"))
	candidate.Declarations[0].Description = "User data"
	report, err := Diff(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Compatible || report.Summary.Compatible != 1 || report.Changes[0].Change != ChangeModified ||
		report.Changes[0].Code != "declaration.description.changed" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffTreatsDomainDescriptionAsCompatible(t *testing.T) {
	baseline := newTestDocument(dataDeclaration("User", "id"))
	baseline.Description = "Baseline domain"
	candidate := newTestDocument(dataDeclaration("User", "id"))
	candidate.Description = "Candidate domain"
	report, err := Diff(baseline, candidate)
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
	report, err := Diff(newTestDocument(baselineConfig), newTestDocument(candidateConfig))
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
	baseline := newTestDocument(dataDeclaration("State", "value"))
	candidate := newTestDocument(enumDeclaration("State", "ACTIVE"))
	report, err := Diff(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Breaking != 1 || len(report.Changes) != 1 || report.Changes[0].Code != "declaration.type.changed" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDiffTreatsAuthenticationAndPermissionSemanticsAsDangerous(t *testing.T) {
	readRequirement := func() *Requirement {
		return &Requirement{Mode: "code", Code: "identity.User:read"}
	}
	writeRequirement := func() *Requirement {
		return &Requirement{Mode: "code", Code: "identity.User:write"}
	}
	tests := []struct {
		name      string
		baseline  *Document
		candidate *Document
		code      string
	}{
		{"actor authentication added", actorDocument(false, false), actorDocument(true, false), "actor.auth.added"},
		{"actor authentication removed", actorDocument(true, false), actorDocument(false, false), "actor.auth.removed"},
		{"actor permission added", actorDocument(false, false), actorDocument(false, true), "actor.permission.added"},
		{"actor permission removed", actorDocument(false, true), actorDocument(false, false), "actor.permission.removed"},
		{"resource permission code changed", resourceDocument("identity.User:read"), resourceDocument("identity.User:write"), "resource.action.code.changed"},
		{"service authentication tightened", servicePolicyDocument("optional", nil), servicePolicyDocument("auth", nil), "service.auth.tightened"},
		{"service authentication relaxed", servicePolicyDocument("auth", nil), servicePolicyDocument("optional", nil), "service.auth.relaxed"},
		{"service permission added", servicePolicyDocument("unset", nil), servicePolicyDocument("unset", readRequirement()), "service.require.added"},
		{"service permission changed", servicePolicyDocument("unset", readRequirement()), servicePolicyDocument("unset", writeRequirement()), "service.require.changed"},
		{"service permission removed", servicePolicyDocument("unset", readRequirement()), servicePolicyDocument("unset", nil), "service.require.removed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report, err := Diff(test.baseline, test.candidate)
			if err != nil {
				t.Fatal(err)
			}
			if !report.Compatible || report.Summary != (Summary{Dangerous: 1}) || len(report.Changes) != 1 ||
				report.Changes[0].Impact != ImpactDangerous || report.Changes[0].Code != test.code {
				t.Fatalf("unexpected report: %+v", report)
			}
		})
	}
}

func TestDiffApiBoundaryIsBreaking(t *testing.T) {
	baseline := newTestDocument(serviceDeclaration("OrderService", "get"))
	candidate := newTestDocument(serviceDeclaration("OrderService", "get"))
	baseline.Declarations[0].Pub = false
	candidate.Declarations[0].Pub = false
	candidate.Declarations[0].Service.Api = true
	report, err := Diff(baseline, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if report.Compatible || len(report.Changes) != 1 || report.Changes[0].Code != "service.api.changed" || report.Changes[0].Impact != ImpactBreaking {
		t.Fatalf("unexpected report: %+v", report)
	}
}
