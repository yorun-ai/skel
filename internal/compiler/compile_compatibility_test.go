package compiler

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/schema"
)

// TODO: Remove this compatibility test when legacy via syntax and warnings are removed.
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

// TODO: Remove legacy via cases when its syntax and warnings are removed;
// keep bare audience validation in the regular compiler tests.
func TestViaWarningsPreserveAudienceValidation(t *testing.T) {
	for _, declaration := range []struct {
		name, body string
		audiences  func(*schema.Domain) []*schema.ActorAudience
	}{
		{"service", "api service EntryApiService { %s auth required method ping {} }", func(d *schema.Domain) []*schema.ActorAudience { return d.Services()[0].Audiences }},
		{"web", "web EntryWeb { %s auth required }", func(d *schema.Domain) []*schema.ActorAudience { return d.Webs()[0].Audiences }},
	} {
		for _, test := range []struct {
			name, audience string
			valid          bool
			warnings       int
		}{
			{"missing actor", "for MissingActor via legacy", false, 0},
			{"duplicate bare", "for UserActor for UserActor", false, 0},
			{"bare", "for UserActor", true, 0},
			{"legacy", "for UserActor via first for UserActor via second", true, 2},
			{"bare then legacy", "for UserActor for UserActor via legacy", true, 1},
			{"legacy then bare", "for UserActor via legacy for UserActor", true, 1},
			{"bare legacy bare", "for UserActor for UserActor via legacy for UserActor", true, 1},
			{"bare bare legacy", "for UserActor for UserActor for UserActor via legacy", false, 0},
		} {
			t.Run(declaration.name+"/"+test.name, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "domain.skel")
				writeFile(t, path, "domain demo\nactor UserActor {}\n"+fmt.Sprintf(declaration.body, test.audience))
				result, err := Compile(Option{SkelIn: path, Strict: true})
				if !test.valid {
					if err == nil {
						t.Fatal("accepted invalid audience")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				audiences := declaration.audiences(result.Domain)
				if len(audiences) != 1 || audiences[0].Actor != "UserActor" {
					t.Fatalf("unexpected audiences: %+v", audiences)
				}
				wantWarnings := test.warnings + strings.Count(test.audience, "for ") - 1
				if len(result.Diagnostics) != wantWarnings {
					t.Fatalf("got diagnostics %+v, want %d warnings", result.Diagnostics, wantWarnings)
				}
				for _, warning := range result.Diagnostics {
					if (warning.Code != diagnostic.CodeViaIgnored && warning.Code != diagnostic.CodeRepeatedFor) || warning.Severity != DiagnosticSeverityWarning {
						t.Fatalf("unexpected diagnostic: %+v", warning)
					}
				}
			})
		}
	}
}
