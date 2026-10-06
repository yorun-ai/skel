package golang_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/codegen"
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

func TestGenerateSensitiveFieldInOtherBindings(t *testing.T) {
	parsed, err := api.Parse(api.Input{SkelIn: "contract.skel", Sources: map[string][]byte{
		"contract.skel": []byte("domain demo.marker\n@sensitive\npub data Credential { skelSensitive: string }"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	input, err := codegen.Prepare(parsed.Domain, codegen.Selection{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"typescript", "skel"} {
		t.Run(name, func(t *testing.T) {
			var generator codegen.Generator
			var err error
			if name == "typescript" {
				generator, err = api.NewTypeScriptGenerator(api.TypeScriptOption{ApiOnly: true, Out: filepath.Join(t.TempDir(), "generated")})
			} else {
				generator, err = api.NewSkeletonGenerator(api.SkeletonOption{Out: filepath.Join(t.TempDir(), "generated"), PubOnly: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			files, err := codegen.Generate(context.Background(), input, generator)
			if err != nil {
				t.Fatalf("Go-specific constraint leaked into %s: %v", name, err)
			}
			for _, file := range files {
				if strings.Contains(file.Content, "skelSensitive") {
					return
				}
			}
			t.Fatalf("%s output lost the field", name)
		})
	}
}
