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
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/testutil"
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
    via client {}
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

func TestGeneratorGoRendersDescriptorFile(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")
	appContext := &schema.Data{
		Pub:  true,
		Name: "AppContext",
		Members: []*schema.DataMember{
			{Name: "name", Type: codegentest.StringType()},
		},
	}

	pkg := newSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.app",
		Data: []*schema.Data{
			appContext,
		},
		Configs: []*schema.Data{
			{
				Pub:       true,
				Name:      "AppConfig",
				Lifecycle: schema.ConfigLifecycleEternal,
				Members: []*schema.DataMember{
					{Name: "title", Type: codegentest.StringType()},
				},
			},
		},
		Actors: []*schema.Actor{
			{Pub: true, Name: "ClientActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}},
		},
		Services: []*schema.Service{
			{
				Pub:  true,
				Name: "AppService",
				Methods: []*schema.Method{
					methodForTest("AppService", &schema.Method{Name: "getContext", ResultType: codegentest.DataType(appContext)}),
				}},
		},
	})

	pubOutDir := filepath.Join(t.TempDir(), "pub")
	if err := generateFixture(pkg, golang.Option{
		Out:          goOutDir,
		AsModule:     true,
		PubOut:       pubOutDir,
		ModulePrefix: "github.com/acme/skel",
	}); err != nil {
		t.Fatal(err)
	}

	goDescriptorContent, err := os.ReadFile(filepath.Join(pubOutDir, "descriptor.go"))
	if err != nil {
		t.Fatalf("read go descriptor file: %v", err)
	}
	if !strings.Contains(string(goDescriptorContent), "skel.RegisterDomainDescriptor(_DomainDescriptor)") {
		t.Fatalf("expected descriptor registration, got:\n%s", string(goDescriptorContent))
	}
	codegentest.AssertGoSourceContains(t, string(goDescriptorContent), `Name: "AppContext"`)
	if !strings.Contains(string(goDescriptorContent), `"AppConfig"`) ||
		!strings.Contains(string(goDescriptorContent), `"demo.app.AppConfig"`) {
		t.Fatalf("expected pub descriptor config declaration, got:\n%s", string(goDescriptorContent))
	}
	codegentest.AssertGoSourceContains(t, string(goDescriptorContent), "Pub: true")
	if !strings.Contains(string(goDescriptorContent), `descriptor.ActorViaClient`) {
		t.Fatalf("expected pub descriptor actor via, got:\n%s", string(goDescriptorContent))
	}
}

// TestGeneratorGoDescriptorHasNoBlankLineInsideDeclarations guards against stray
// blank lines inside rendered schema blocks. gofmt preserves blank lines a
// template emits, so a conditional block that leaves one behind would open a
// gap in the compact declaration layout.
func TestGeneratorGoDescriptorHasNoBlankLineInsideDeclarations(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")

	userData := &schema.Data{
		Pub:         true,
		Name:        "User",
		Description: "User record",
		Members: []*schema.DataMember{
			{Name: "id", Type: codegentest.StringType()},
			{Name: "status", Sensitive: true, Type: codegentest.ScalarType(schema.ScalarInt)},
		},
	}

	pkg := newSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{
			userData,
		},
		Configs: []*schema.Data{
			{
				Pub:       true,
				Name:      "AppConfig",
				Lifecycle: schema.ConfigLifecycleEternal,
				Members: []*schema.DataMember{
					{Name: "title", Type: codegentest.StringType()},
				},
			},
		},
		Enums: []*schema.Enum{
			{Name: "Status", Items: []*schema.EnumItem{{Name: "ACTIVE"}}},
		},
		Actors: []*schema.Actor{
			{Pub: true, Name: "ClientActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}},
		},
		Services: []*schema.Service{
			{
				Pub:  true,
				Name: "UserService",
				Methods: []*schema.Method{
					methodForTest("UserService", &schema.Method{Name: "getUser", ResultType: codegentest.DataType(userData)}),
				}},
		},
		Tasks: []*schema.Task{
			{
				Name: "RebuildTask",
				Triggers: []*schema.TaskTrigger{
					triggerForTest("RebuildTask", &schema.TaskTrigger{
						Name:      "atTime",
						Arguments: []*schema.Argument{{Name: "startAt", Type: codegentest.LocalDateTimeType()}},
					}),
				},
			},
		},
		Events: []*schema.Data{
			{
				Name: "UserCreated",
				Members: []*schema.DataMember{
					{Name: "userId", Type: codegentest.StringType()},
				},
			},
		},
	})

	if err := generateFixture(pkg, golang.Option{Out: goOutDir}); err != nil {
		t.Fatal(err)
	}

	goDescriptorContent := readFileForTest(t, filepath.Join(goOutDir, "descriptor.go"))
	assertNoBlankLineInsideDeclaration(t, goDescriptorContent)
}

// assertNoBlankLineInsideDeclaration reports a failure when a blank line opens
// or closes a rendered block, which the compact layout only allows between
// top-level schema sections.
func assertNoBlankLineInsideDeclaration(t *testing.T, content string) {
	t.Helper()
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			continue
		}
		previous := ""
		if i > 0 {
			previous = strings.TrimSpace(lines[i-1])
		}
		if strings.HasSuffix(previous, "{") {
			t.Fatalf("unexpected blank line after %q at line %d:\n%s", previous, i, content)
		}
		next := ""
		if i+1 < len(lines) {
			next = strings.TrimSpace(lines[i+1])
		}
		if strings.HasPrefix(next, "}") {
			t.Fatalf("unexpected blank line before %q at line %d:\n%s", next, i+2, content)
		}
	}
}

func TestGeneratorGoRendersWebMountInSpecAndDescriptor(t *testing.T) {
	for _, path := range []string{"", "/", "/portal/v1-assets/"} {
		pkg := newSchemaDomainForTest(t, schema.DomainSpec{
			Name:   "demo.web",
			Actors: []*schema.Actor{{Name: "ClientActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}}},
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

func TestGeneratedAuthModesWithPublishedVine(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	input := filepath.Join(root, "domain.skel")
	writeFileForTest(t, input, `domain demo.auth
actor ClientActor { via client {} }
api service SessionApiService {
 for ClientActor via client
 auth required
 method profile {}
 method browse { auth optional }
 method login { auth anonymous }
}
web RequiredWeb { for ClientActor via client auth required }
web OptionalWeb { for ClientActor via client auth optional }
web AnonymousWeb { for ClientActor via client auth anonymous }
web OffWeb { for ClientActor via client auth off }
`)
	out := filepath.Join(root, "auth")
	if _, err := api.CompileGolang(api.Input{SkelIn: input, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Module: "example.com/auth", AsModule: true}); err != nil {
		t.Fatal(err)
	}
	writeFileForTest(t, filepath.Join(out, "auth_test.go"), `package auth
import (
 "testing"
 "go.yorun.ai/skel/descriptor"
)
func TestAuthModes(t *testing.T) {
 if _DomainDescriptor.Services[0].AuthMode != descriptor.AuthModeRequired { t.Fatal("wrong service auth") }
 methods:=_DomainDescriptor.Services[0].Methods
 want:=map[string]descriptor.AuthMode{"profile":descriptor.AuthModeInherit,"browse":descriptor.AuthModeOptional,"login":descriptor.AuthModeAnonymous}
 for _,method:=range methods { if method.AuthMode!=want[method.Name] { t.Fatalf("method %s: %s",method.Name,method.AuthMode) } }
 webs:=map[string]descriptor.AuthMode{"RequiredWeb":descriptor.AuthModeRequired,"OptionalWeb":descriptor.AuthModeOptional,"AnonymousWeb":descriptor.AuthModeAnonymous,"OffWeb":descriptor.AuthModeOff}
 for _,web:=range _DomainDescriptor.Webs { if web.AuthMode!=webs[web.Name] { t.Fatalf("web %s: %s",web.Name,web.AuthMode) } }
}
`)
	testutil.UseLocalSkel(t, out)
	testutil.Go(t, out, "test", "-mod=mod", "./...")
}

func TestApiBackendDescriptorAndClientBoundary(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "order.skel")
	writeFileForTest(t, entry, `domain demo.order
actor TestActor { via client {} }
api service OrderApiService { auth required  for TestActor via client method ping {} }
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
