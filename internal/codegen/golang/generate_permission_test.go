package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel"
	"go.yorun.ai/skel/internal/testutil"
)

func TestPermissionGenerationUsesStrings(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	input := filepath.Join(root, "domain.skel")
	contract := `domain demo
pub data Permissions { code: string
 optional: string? }
pub resource User {
    check byId { input {
        id: string
    } }
    action read
    action update
}
actor ClientActor { via client {} }
service UserService {
    for ClientActor
    method update {
        require User:update:byId(id)
        input { id: string }
    }
}`
	if err := os.WriteFile(input, []byte(contract), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "generated")
	if _, err := skel.CompileGolang(skel.Input{SkelIn: input}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, AsModule: true, Module: "example.com/permissions"}); err != nil {
		t.Fatal(err)
	}
	resource := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "resource.go"))), " ")
	if !strings.Contains(resource, `UserReadPermission string = "demo.User:read"`) || !strings.Contains(resource, "code string") {
		t.Fatalf("expected string constant and check parameter: %s", resource)
	}
	tsOut := filepath.Join(root, "ts")
	if _, err := skel.CompileTypeScript(skel.Input{SkelIn: input}, skel.TypeScriptOption{ApiOnly: true, Out: tsOut}); err != nil {
		t.Fatal(err)
	}
	ts := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(tsOut, "data.ts"))), " ")
	if !strings.Contains(ts, "code: string") || !strings.Contains(ts, "optional: string | null") {
		t.Fatalf("expected string TypeScript fields: %s", ts)
	}
	public := filepath.Join(root, "public")
	if _, err := skel.CompileSkeleton(skel.Input{SkelIn: input}, skel.SkeletonOption{Out: public, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := skel.CompileGolang(skel.Input{SkelIn: public}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: filepath.Join(root, "roundtrip")}); err != nil {
		t.Fatal(err)
	}
	testutil.Go(t, out, "test", "-mod=mod", "./...")
}

func TestPermissionCodeArgumentNameGeneration(t *testing.T) {
	for _, scope := range []string{"resource", "action"} {
		t.Run(scope, func(t *testing.T) {
			root := t.TempDir()
			check := `check byCode { input { code: string
code1: string } }`
			resourceBody := check + "\naction read"
			if scope == "action" {
				resourceBody = "action read {\n" + check + "\n}"
			}
			input := filepath.Join(root, "domain.skel")
			contract := "domain demo\npub resource User {\n" + resourceBody + "\ncheck enabled {}\n}\n" + `
pub actor ClientActor { via client {} }
pub service UserService {
 for ClientActor via client
 method read {
  require all(User:read:byCode(code, code1), User:read:enabled())
  input { code: string
   code1: string }
 }
}`
			if err := os.WriteFile(input, []byte(contract), 0o600); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(root, "generated")
			if _, err := skel.CompileGolang(skel.Input{SkelIn: input}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out}); err != nil {
				t.Fatal(err)
			}
			generatedSchema := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "schema.go"))), " ")
			for _, want := range []string{`CodeArgumentName: "code2"`, `CodeArgumentName: "code"`, `Name: "code", JsonPath: "code"`, `Name: "code1", JsonPath: "code1"`} {
				if !strings.Contains(generatedSchema, want) {
					t.Fatalf("missing %s in schema:\n%s", want, generatedSchema)
				}
			}
			resource := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "resource.go"))), " ")
			for _, want := range []string{`json:"code2" skel:"index(0)"`, `json:"code" skel:"index(1)"`, `json:"code1" skel:"index(2)"`} {
				if !strings.Contains(resource, want) {
					t.Fatalf("missing %s in resource:\n%s", want, resource)
				}
			}
			public := filepath.Join(root, "public")
			if _, err := skel.CompileSkeleton(skel.Input{SkelIn: input}, skel.SkeletonOption{Out: public, PubOnly: true}); err != nil {
				t.Fatal(err)
			}
			out2 := filepath.Join(root, "roundtrip")
			if _, err := skel.CompileGolang(skel.Input{SkelIn: public}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out2}); err != nil {
				t.Fatal(err)
			}
			if got := readFileForTest(t, filepath.Join(out2, "resource.go")); !strings.Contains(got, `json:"code2" skel:"index(0)"`) {
				t.Fatalf("public round trip lost injection name: %s", got)
			}
			// Resolve the same check through an imported public resource.
			consumer := filepath.Join(root, "consumer.skel")
			source := `domain consumer
import demo
actor ClientActor { via client {} }
service ConsumerService {
 for ClientActor via client
 method read {
  require demo.User:read:byCode(code, code1)
  input { code: string
   code1: string }
 }
}`
			if err := os.WriteFile(consumer, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			consumerOut := filepath.Join(root, "consumer")
			if _, err := skel.CompileGolang(skel.Input{SkelIn: consumer, SkelImports: map[string]string{"demo": public}}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: consumerOut, Imports: map[string]string{"demo": "example.com/demo"}}); err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(consumerOut, "schema.go"))), " "); !strings.Contains(got, `CodeArgumentName: "code2"`) {
				t.Fatalf("imported check lost injection name: %s", got)
			}
		})
	}
}
