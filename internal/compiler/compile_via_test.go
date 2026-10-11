package compiler

import (
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/diagnostic"
)

func TestViaWarningsAcrossCompilerEntryPoints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeFile(t, path, `domain demo
actor UserActor { via UNKNOWN {} }
api service EntryApiService { for UserActor via other auth required method ping {} }
web EntryWeb { for UserActor via another auth required }
`)
	for _, strict := range []bool{false, true} {
		option := Option{SkelIn: path, Strict: strict}
		checked, err := Check(option)
		if err != nil {
			t.Fatal(err)
		}
		sets := []Diagnostics{checked.Diagnostics}
		for _, compile := range []func(Option) (Result, error){Compile, CompileImport} {
			result, err := compile(option)
			if err != nil {
				t.Fatal(err)
			}
			sets = append(sets, result.Diagnostics)
		}
		for _, diagnostics := range sets {
			if len(diagnostics) != 3 || diagnostics.HasErrors() {
				t.Fatalf("expected three warnings: %+v", diagnostics)
			}
			for index, warning := range diagnostics {
				if warning.Code != diagnostic.CodeViaIgnored || warning.Severity != DiagnosticSeverityWarning || warning.Range.Start.Line != index+2 || warning.Range.Start.File != path {
					t.Fatalf("unexpected warning: %+v", warning)
				}
			}
		}
	}
}

func TestViaWarningsPreserveAudienceValidation(t *testing.T) {
	for _, audience := range []string{"for MissingActor via legacy", "for UserActor for UserActor"} {
		path := filepath.Join(t.TempDir(), "domain.skel")
		writeFile(t, path, "domain demo\nactor UserActor {}\napi service EntryApiService { "+audience+" auth required method ping {} }")
		if _, err := Compile(Option{SkelIn: path}); err == nil {
			t.Fatalf("accepted invalid audience: %s", audience)
		}
	}
}
