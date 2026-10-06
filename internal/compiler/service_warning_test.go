package compiler

import (
	"context"
	"strings"
	"testing"

	"go.yorun.ai/skel/diagnostic"
)

func TestServiceWarningsAndApiModifiers(t *testing.T) {
	source := Source{Path: "/workspace/api.skel", Content: []byte(`domain demo.order
service LegacyService { method ping {} }
pub service DualService { method ping { auth optional } }
actor TestActor { via client {} }
api service ClientApiService { for TestActor via client auth required method ping {} }
pub service BackendService { method ping {} }
`)}
	analyzer := NewWorkspaceAnalyzer()
	for range 2 {
		diagnostics, _, err := analyzer.analyze(context.Background(), []Source{source}, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 2 || diagnostics[0].Code != diagnostic.CodeServiceModifier || diagnostics[1].Code != diagnostic.CodeServiceClientRules {
			t.Fatalf("unexpected diagnostics: %v", diagnostics)
		}
		for _, d := range diagnostics {
			if d.Severity != DiagnosticSeverityWarning || d.Range.Start.Line < 2 {
				t.Fatalf("invalid warning: %+v", d)
			}
		}
	}
}

func TestApiServiceNameSuffix(t *testing.T) {
	for _, test := range []struct {
		declaration string
		valid       bool
	}{
		{declaration: "api service OrderApiService", valid: true},
		{declaration: "api service OrderService"},
		{declaration: "api service OrderAPIService"},
		{declaration: "api service ApiService"},
		{declaration: "pub service OrderService", valid: true},
		{declaration: "service OrderService", valid: true},
	} {
		t.Run(test.declaration, func(t *testing.T) {
			source := Source{Path: "/workspace/api.skel", Content: []byte("domain demo.order\n" + test.declaration + " { for TestActor via client method ping {} }\nactor TestActor { via client {} }\n")}
			analyzer := NewWorkspaceAnalyzer()
			for range 2 {
				diagnostics, _, err := analyzer.analyze(context.Background(), []Source{source}, true)
				if err != nil {
					t.Fatal(err)
				}
				hasError := false
				for _, d := range diagnostics {
					if d.Severity == DiagnosticSeverityError {
						hasError = true
						if d.Range.Start.Line != 2 || !strings.Contains(d.Message, "ApiService") {
							t.Fatalf("unexpected naming diagnostic: %+v", d)
						}
					}
				}
				if hasError == test.valid {
					t.Fatalf("unexpected diagnostics: %+v", diagnostics)
				}
			}
		})
	}
}

func TestExtServiceIncrementalAnalysis(t *testing.T) {
	analyzer := NewWorkspaceAnalyzer()
	for _, modifier := range []string{"ext", "pub", "ext"} {
		source := Source{Path: "/workspace/ext.skel", Content: []byte("domain demo.storage\n" + modifier + " service StorageService { method ping {} }\n")}
		diagnostics, domains, err := analyzer.analyze(context.Background(), []Source{source}, true)
		if err != nil || len(diagnostics) != 0 {
			t.Fatalf("%s: %v %v", modifier, diagnostics, err)
		}
		if len(domains) != 1 || domains[0].Model.Services()[0].Ext != (modifier == "ext") {
			t.Fatalf("stale ext modifier after %s", modifier)
		}
	}
}

func TestApiServiceRequiresActorAudience(t *testing.T) {
	analyzer := NewWorkspaceAnalyzer()
	for _, auth := range []string{"", "noauth", "auth"} {
		source := Source{Path: "/workspace/api.skel", Content: []byte("domain demo.order\napi service OrderApiService { " + auth + " method ping {} }\n")}
		for range 2 {
			diagnostics, _, err := analyzer.analyze(context.Background(), []Source{source}, true)
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 1 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
			d := diagnostics[0]
			if d.Code != diagnostic.CodeSemanticValidation || d.Severity != DiagnosticSeverityError || d.Range.Start.File != source.Path || d.Range.Start.Line != 2 || !strings.Contains(d.Message, "at least one for Actor") {
				t.Fatalf("unexpected diagnostic: %+v", d)
			}
		}
		source.Content = []byte("domain demo.order\nactor ClientActor { via client {} }\napi service OrderApiService { for ClientActor via client " + auth + " method ping {} }\n")
		diagnostics, _, err := analyzer.analyze(context.Background(), []Source{source}, true)
		if err != nil || Diagnostics(diagnostics).HasErrors() {
			t.Fatalf("valid audience rejected: %v, %v", diagnostics, err)
		}
	}
}
