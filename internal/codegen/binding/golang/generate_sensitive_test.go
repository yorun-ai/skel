package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/codegen/binding/golang"
	compiler "go.yorun.ai/skel/internal/compiler"
)

func TestGenerateSensitiveMarkerConflicts(t *testing.T) {
	declarations := map[string]string{
		"data":         "@sensitive\npub data Credential { skelSensitive: string }",
		"generic data": "@sensitive\npub data Credential<TValue> { skelSensitive: TValue }",
		"config":       "@sensitive\nconfig SecretConfig eternal { skelSensitive: string }",
		"event":        "pub event ChangedEvent { @sensitive payload { skelSensitive: string } }",
		"actor credential": `pub actor ClientActor {
    via client {}
    auth {
        @sensitive credential { skelSensitive: string }
        info { @identifier id: int }
    }
}`,
		"actor info": `pub actor ClientActor {
    via client {}
    auth {
        credential { token: string }
        @sensitive info { @identifier skelSensitive: int }
    }
}`,
	}
	for name, declaration := range declarations {
		for _, sensitive := range []bool{false, true} {
			testName := name + "/plain"
			if sensitive {
				testName = name + "/sensitive"
			}
			t.Run(testName, func(t *testing.T) {
				text := declaration
				if !sensitive {
					text = strings.ReplaceAll(text, "@sensitive", "")
				}
				input := api.Input{SkelIn: "contract.skel", Sources: map[string][]byte{
					"contract.skel": []byte("domain demo.marker\n" + text),
				}}
				checked, err := api.Check(api.CheckOption{SkelIn: input.SkelIn, Sources: input.Sources})
				if err != nil || !checked.Valid {
					t.Fatalf("Go field names must pass language validation: %+v, %v", checked, err)
				}
				parsed, err := api.Parse(input)
				if err != nil {
					t.Fatal(err)
				}
				out := filepath.Join(t.TempDir(), "generated")
				writeFileForTest(t, filepath.Join(out, "keep.txt"), "existing content")
				err = api.GenerateGolang(parsed.Domain, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out})
				if sensitive {
					if err == nil || !strings.Contains(err.Error(), "conflicts with generated sensitive marker method SkelSensitive") || !strings.Contains(err.Error(), "contract.skel:") {
						t.Fatalf("expected a positioned Go marker conflict, got %v", err)
					}
					entries, readErr := os.ReadDir(out)
					if readErr != nil || len(entries) != 1 || entries[0].Name() != "keep.txt" {
						t.Fatalf("failed Go generation changed output: %v, %v", entries, readErr)
					}
					if got := readFileForTest(t, filepath.Join(out, "keep.txt")); got != "existing content" {
						t.Fatalf("failed Go generation overwrote existing content: %q", got)
					}
				} else if err != nil {
					t.Fatalf("field without a generated marker must be allowed: %v", err)
				}
			})
		}
	}
}

func TestGenerateSensitiveMarkerConflictUsesSelectedGoView(t *testing.T) {
	parsed, err := api.Parse(api.Input{SkelIn: "contract.skel", Sources: map[string][]byte{
		"contract.skel": []byte(`domain demo.marker
@sensitive
data PrivateCredential { skelSensitive: string }
pub data PublicRecord { id: int }
`),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.GenerateGolang(parsed.Domain, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: filepath.Join(t.TempDir(), "generated"), PubOnly: true}); err != nil {
		t.Fatalf("unselected private declaration must not block public output: %v", err)
	}
	if err := api.GenerateGolang(parsed.Domain, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: filepath.Join(t.TempDir(), "generated"), ApiOnly: true}); err != nil {
		t.Fatalf("unselected private declaration must not block API output: %v", err)
	}
}

func TestGenerateSensitiveTaskInputEndToEnd(t *testing.T) {
	inputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(inputDir, "domain.skel"), []byte("domain demo.task\n"), 0o644); err != nil {
		t.Fatalf("write domain source: %v", err)
	}
	source := `domain demo.task

task RebuildIndexTask {
    trigger atTime {
        @sensitive
        input {
            @sensitive
            token: string
        }
    }
}
`
	if err := os.WriteFile(filepath.Join(inputDir, "task.skel"), []byte(source), 0o644); err != nil {
		t.Fatalf("write task source: %v", err)
	}

	parsed, err := compiler.Compile(compiler.Option{SkelIn: inputDir})
	if err != nil {
		t.Fatalf("parse Skel source: %v", err)
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", parsed.Diagnostics)
	}
	trigger := parsed.Domain.Tasks()[0].Triggers[0]
	if !trigger.ArgumentsSensitive || !trigger.Arguments[0].Sensitive {
		t.Fatalf("sensitive metadata was not preserved by analysis: %+v", trigger)
	}
	if trigger.Hash == "" {
		t.Fatal("expected trigger compatibility hash")
	}

	outputDir := filepath.Join(t.TempDir(), "skeled")
	if err := generateFixture(parsed.Domain, golang.Option{Out: outputDir}); err != nil {
		t.Fatalf("generate Go: %v", err)
	}

	taskContent := readFileForTest(t, filepath.Join(outputDir, "task.go"))
	if !strings.Contains(taskContent, "ArgumentsSensitive: true,") {
		t.Fatalf("expected sensitive trigger input in TriggerSpec, got:\n%s", taskContent)
	}
	if !strings.Contains(taskContent, `json:"token" skel:"sensitive"`) {
		t.Fatalf("expected sensitive task argument tag, got:\n%s", taskContent)
	}

	schemaContent := readFileForTest(t, filepath.Join(outputDir, "descriptor.go"))
	if !strings.Contains(schemaContent, "ArgumentsSensitive: true,") ||
		!strings.Contains(schemaContent, "Sensitive: true,") {
		t.Fatalf("expected sensitive task metadata in DomainSchema, got:\n%s", schemaContent)
	}
}
