package analyzer

import (
	"errors"
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestDiagnosticReporterDoesNotUsePanicControlFlow(t *testing.T) {
	reporter := newDiagnosticReporter()

	if reporter.check(false, "invalid value") {
		t.Fatal("expected failed check")
	}
	if len(reporter.result()) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(reporter.result()))
	}
}

func TestDiagnosticReporterUsesExplicitMetadata(t *testing.T) {
	reporter := newDiagnosticReporter()
	reporter.reportDuplicatef("conflict")
	reporter.reportReferencef("missing declaration")
	reporter.reportNamingf("UserProfile", "wrong name")
	reporter.reportf("duplicated unknown value not found; expected=WrongName")

	expectedCodes := map[string]string{
		"conflict":            DiagnosticCodeDuplicate,
		"missing declaration": DiagnosticCodeReference,
		"wrong name":          DiagnosticCodeNaming,
		"duplicated unknown value not found; expected=WrongName": DiagnosticCodeValidation,
	}
	for _, err := range reporter.result() {
		var failure *Failure
		if !errors.As(err, &failure) {
			t.Fatalf("expected structured failure, got %T", err)
		}
		if failure.Code != expectedCodes[failure.Message] {
			t.Fatalf("unexpected code for %q: %s", failure.Message, failure.Code)
		}
		if failure.Message == "wrong name" {
			if failure.Suggestion == nil || failure.Suggestion.Replacement != "UserProfile" {
				t.Fatalf("unexpected naming suggestion: %+v", failure.Suggestion)
			}
		} else if failure.Suggestion != nil {
			t.Fatalf("unexpected suggestion for %q: %+v", failure.Message, failure.Suggestion)
		}
	}
}

func TestAnalyzeReturnsErrorsWithoutPanicking(t *testing.T) {
	_, diagnostics := Analyze(&grammar.SkelContent{
		Domain: domainContent("demo"),
		Entries: []*grammar.SkelEntry{
			{Data: &grammar.Data{Name: ident("User"), Members: []*grammar.DataMember{
				{Name: ident("first"), Type: refGrammarType("MissingFirst")},
				{Name: ident("second"), Type: refGrammarType("MissingSecond")},
			}}},
		},
	}, nil)

	if len(diagnostics) != 2 {
		t.Fatalf("expected two diagnostics, got %d: %v", len(diagnostics), diagnostics)
	}
}
