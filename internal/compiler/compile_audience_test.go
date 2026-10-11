package compiler

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/schema"
)

func TestCompileActorLists(t *testing.T) {
	for _, declaration := range []struct {
		name, body string
		audiences  func(*schema.Domain) []*schema.ActorAudience
	}{
		{"service", "api service EntryApiService {\n%s\nauth required\nmethod ping {}\n}", func(d *schema.Domain) []*schema.ActorAudience { return d.Services()[0].Audiences }},
		{"web", "web EntryWeb {\n%s\nauth off\n}", func(d *schema.Domain) []*schema.ActorAudience { return d.Webs()[0].Audiences }},
	} {
		for _, test := range []struct {
			name, clause string
			want         []string
		}{
			{"single", "for AActor", []string{"AActor"}},
			{"comma", "for AActor, BActorV10", []string{"AActor", "BActorV10"}},
			{"multiline", "for AActor,\n BActorV10", []string{"AActor", "BActorV10"}},
			{"comma on next line", "for AActor\n, BActorV10", []string{"AActor", "BActorV10"}},
			{"comments", "for AActor /* first */, // next\n BActorV10", []string{"AActor", "BActorV10"}},
		} {
			t.Run(declaration.name+"/"+test.name, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "entry.skel")
				writeFile(t, path, "domain demo\nactor AActor {}\nactor BActorV10 {}\n"+fmt.Sprintf(declaration.body, test.clause))
				result, err := Compile(Option{SkelIn: path, Strict: true})
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Diagnostics) != 0 {
					t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
				}
				var names []string
				for _, audience := range declaration.audiences(result.Domain) {
					names = append(names, audience.Actor)
				}
				if !slices.Equal(names, test.want) {
					t.Fatalf("actors = %v, want %v", names, test.want)
				}
			})
		}
		for _, clause := range []string{"for", "for , AActor", "for AActor,", "for AActor,, BActorV10", "for AActor {} {}", "for AActor { auth off }", "for AActor, AActor", "for AActor, MissingActor"} {
			t.Run(declaration.name+"/invalid/"+clause, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "entry.skel")
				writeFile(t, path, "domain demo\nactor AActor {}\nactor BActorV10 {}\n"+fmt.Sprintf(declaration.body, clause))
				if _, err := Compile(Option{SkelIn: path}); err == nil {
					t.Fatal("accepted invalid audience")
				}
			})
		}
	}
}

// TODO: Remove these migration cases when repeated for clauses and empty blocks are no longer accepted.
func TestCompileLegacyActorListWarnings(t *testing.T) {
	for _, owner := range []string{"api service EntryApiService", "web EntryWeb"} {
		for _, test := range []struct {
			clause string
			codes  []string
		}{
			{"for AActor\nfor BActor", []string{diagnostic.CodeRepeatedFor}},
			{"for AActor for BActor", []string{diagnostic.CodeRepeatedFor}},
			{"for AActor {}, BActor", []string{diagnostic.CodeForBlockIgnored}},
			{"for AActor, BActor {\n/* ignored */\n}", []string{diagnostic.CodeForBlockIgnored}},
			{"for AActor {}\nfor BActor {}", []string{diagnostic.CodeForBlockIgnored, diagnostic.CodeRepeatedFor, diagnostic.CodeForBlockIgnored}},
		} {
			t.Run(owner+"/"+test.clause, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "entry.skel")
				tail := "auth required\n}"
				if strings.HasPrefix(owner, "api") {
					tail = "auth required\nmethod ping {}\n}"
				}
				prefix := "domain demo\nactor AActor {}\nactor BActor {}\n" + owner + " {\n"
				writeFile(t, path, prefix+"for AActor, BActor\n"+tail)
				canonical, err := Compile(Option{SkelIn: path})
				if err != nil {
					t.Fatal(err)
				}
				source := prefix + test.clause + "\n" + tail
				writeFile(t, path, source)
				for _, strict := range []bool{false, true} {
					result, err := Compile(Option{SkelIn: path, Strict: strict})
					if err != nil {
						t.Fatal(err)
					}
					if result.Domain.Hash() != canonical.Domain.Hash() {
						t.Fatal("legacy syntax changed the contract hash")
					}
					var codes []string
					for _, warning := range result.Diagnostics {
						codes = append(codes, warning.Code)
						if warning.Severity != DiagnosticSeverityWarning || warning.Range.Start.File != path {
							t.Fatalf("unexpected diagnostic: %+v", warning)
						}
						line := strings.Split(source, "\n")[warning.Range.Start.Line-1]
						token := "for"
						if warning.Code == diagnostic.CodeForBlockIgnored {
							token = "{"
						}
						if !strings.HasPrefix(line[warning.Range.Start.Column-1:], token) {
							t.Fatalf("warning points at wrong token: %+v", warning)
						}
					}
					if !slices.Equal(codes, test.codes) {
						t.Fatalf("warnings = %v, want %v", codes, test.codes)
					}
				}
			})
		}
	}
}

func TestCompileActorListReferenceLocations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entry.skel")
	for _, name := range []string{"AActor", "MissingActor"} {
		writeFile(t, path, "domain demo\nactor AActor {}\napi service EntryApiService {\nfor AActor, "+name+"\nauth required method ping {}\n}\n")
		result, err := Check(Option{SkelIn: path})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Diagnostics) != 1 {
			t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
		}
		failure := result.Diagnostics[0]
		wantCode := diagnostic.CodeSemanticReference
		if name == "AActor" {
			wantCode = diagnostic.CodeSemanticDuplicate
		}
		if failure.Code != wantCode || failure.Range.Start.Line != 4 || failure.Range.Start.Column != 13 {
			t.Fatalf("wrong reference diagnostic: %+v", failure)
		}
	}
}
