package golang_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/schema"
)

func TestActorRegistrationImportDoesNotCollideWithSourceAlias(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, "shared.skel")
	entry := filepath.Join(root, "actor.skel")
	writeFileForTest(t, shared, "domain shared\npub data Credential { token: string }\n")
	writeFileForTest(t, entry, `domain demo
import shared as meta
actor ClientActor {
    auth {
        credential { token: string }
        info { id: string details: meta.Credential }
    }
}`)
	out := filepath.Join(root, "generated")
	_, err := api.CompileGolang(api.Input{SkelIn: entry, SkelImports: map[string]string{"shared": shared}}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Imports: map[string]string{"shared": "example.com/shared"}})
	if err != nil {
		t.Fatal(err)
	}
	content := readFileForTest(t, filepath.Join(out, "actor.go"))
	file, err := parser.ParseFile(token.NewFileSet(), "actor.go", content, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range file.Imports {
		if item.Path.Value == `"example.com/shared"` && (item.Name == nil || item.Name.Name == "meta") {
			t.Fatalf("source alias collides with actor registration import: %s", content)
		}
	}
	for _, fragment := range []string{
		`"go.yorun.ai/vine/core/meta"`,
		"meta.RegisterActor(meta.ActorSpec{",
		"reflect.TypeFor[*ClientActorInfo]()",
		"type ClientActorCredential struct",
		"type ClientActorInfo struct",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("missing %s in generated actor: %s", fragment, content)
		}
	}
}

func TestGeneratorGoRendersWebMountInSpecAndDescriptor(t *testing.T) {
	for _, path := range []string{"", "/", "/portal/v1-assets/"} {
		pkg := newSchemaDomainForTest(t, schema.DomainSpec{
			Name:   "demo.web",
			Actors: []*schema.Actor{{Name: "ClientActor"}},
			Webs:   []*schema.Web{{Name: "PortalWeb", AuthMode: schema.AuthModeRequired, MountPath: path, Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}}},
		})
		out := filepath.Join(t.TempDir(), "skeled")
		if err := generateFixture(pkg, golang.Option{Out: out}); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"web.go", "descriptor.go"} {
			first := readFileForTest(t, filepath.Join(out, name))
			normalized := strings.Join(strings.Fields(first), " ")
			if path == "" {
				if strings.Contains(normalized, "MountPath:") {
					t.Fatalf("unspecified mount emitted in %s", name)
				}
			} else if !strings.Contains(normalized, `MountPath: "`+path+`",`) {
				t.Fatalf("mount path missing from %s:\n%s", name, first)
			}
			if err := generateFixture(pkg, golang.Option{Out: out}); err != nil {
				t.Fatal(err)
			}
			if second := readFileForTest(t, filepath.Join(out, name)); first != second {
				t.Fatal("non-deterministic generation")
			}
		}
	}
}

func TestApiBackendDescriptorAndClientBoundary(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "order.skel")
	writeFileForTest(t, entry, `domain demo.order
actor TestActor {}
api service OrderApiService { auth required  for TestActor method ping {} }
pub service BackendService { method ping {} }
`)
	input := api.Input{SkelIn: entry}
	out := filepath.Join(root, "server")
	option := api.GolangOption{CompilerVersion: "v0.0.0-dev", AsModule: true, Module: "example.com/orderserver", Out: out}
	if _, err := api.CompileGolang(input, option); err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(filepath.Join(out, "descriptor.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(schema), "Api:") != 1 {
		t.Fatalf("missing explicit API schema: %s", schema)
	}
	service, err := os.ReadFile(filepath.Join(out, "service.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(service), "NewOrderApiServiceClient") || !strings.Contains(string(service), "NewBackendServiceClient") {
		t.Fatalf("wrong backend client generation: %s", service)
	}
	mod, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "go.yorun.ai/vine "+api.DefaultGolangVineVersion) {
		t.Fatalf("unsupported Vine requirement: %s", mod)
	}
	option.VineVersion = "v0.15.6"
	if _, err := api.CompileGolang(input, option); err == nil {
		t.Fatal("accepted runtime below the minimum Vine version")
	}
}
