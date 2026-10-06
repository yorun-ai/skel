package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/testutil"
)

func TestGeneratorGoRendersSchemaFile(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")
	appContext := &model.Data{
		Pub:  true,
		Name: "AppContext",
		Members: []*model.DataMember{
			{Name: "name", Type: codegentest.StringType()},
		},
	}

	pkg := newModelDomainForTest(t, model.DomainSpec{
		Name: "demo.app",
		Data: []*model.Data{
			appContext,
		},
		Configs: []*model.Data{
			{
				Pub:       true,
				Name:      "AppConfig",
				Lifecycle: model.ConfigLifecycleEternal,
				Members: []*model.DataMember{
					{Name: "title", Type: codegentest.StringType()},
				},
			},
		},
		Actors: []*model.Actor{
			{Pub: true, Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}},
		},
		Services: []*model.Service{
			{
				Pub:       true,
				Name:      "AppService",
				Audiences: []*model.ActorAudience{{Actor: "ClientActor", Via: string(model.ActorViaClient)}},
				Methods: []*model.Method{
					methodForTest("AppService", &model.Method{Name: "getContext", ResultType: codegentest.DataType(appContext)}),
				},
			},
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

	goSchemaContent, err := os.ReadFile(filepath.Join(pubOutDir, "schema.go"))
	if err != nil {
		t.Fatalf("read go schema file: %v", err)
	}
	if !strings.Contains(string(goSchemaContent), "skel.RegisterDomainSchema(_DomainSchema)") {
		t.Fatalf("expected schema go registration, got:\n%s", string(goSchemaContent))
	}
	codegentest.AssertGoSourceContains(t, string(goSchemaContent), `Name: "AppContext"`)
	if !strings.Contains(string(goSchemaContent), `"AppConfig"`) ||
		!strings.Contains(string(goSchemaContent), `"demo.app.AppConfig"`) {
		t.Fatalf("expected pub schema config declaration, got:\n%s", string(goSchemaContent))
	}
	codegentest.AssertGoSourceContains(t, string(goSchemaContent), "Pub: true")
	if !strings.Contains(string(goSchemaContent), `Via: skel.ActorViaClient`) {
		t.Fatalf("expected pub schema actor via, got:\n%s", string(goSchemaContent))
	}
}

// TestGeneratorGoSchemaHasNoBlankLineInsideDeclarations guards against stray
// blank lines inside rendered schema blocks. gofmt preserves blank lines a
// template emits, so a conditional block that leaves one behind would open a
// gap in the compact declaration layout.
func TestGeneratorGoSchemaHasNoBlankLineInsideDeclarations(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")

	userData := &model.Data{
		Pub:         true,
		Name:        "User",
		Description: "User record",
		Members: []*model.DataMember{
			{Name: "id", Type: codegentest.StringType()},
			{Name: "status", Sensitive: true, Type: codegentest.ScalarType(model.ScalarInt)},
		},
	}

	pkg := newModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{
			userData,
		},
		Configs: []*model.Data{
			{
				Pub:       true,
				Name:      "AppConfig",
				Lifecycle: model.ConfigLifecycleEternal,
				Members: []*model.DataMember{
					{Name: "title", Type: codegentest.StringType()},
				},
			},
		},
		Enums: []*model.Enum{
			{Name: "Status", Items: []*model.EnumItem{{Name: "ACTIVE"}}},
		},
		Actors: []*model.Actor{
			{Pub: true, Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}},
		},
		Services: []*model.Service{
			{
				Pub:       true,
				Name:      "UserService",
				Audiences: []*model.ActorAudience{{Actor: "ClientActor", Via: string(model.ActorViaClient)}},
				Methods: []*model.Method{
					methodForTest("UserService", &model.Method{Name: "getUser", ResultType: codegentest.DataType(userData)}),
				},
			},
		},
		Tasks: []*model.Task{
			{
				Name: "RebuildTask",
				Triggers: []*model.TaskTrigger{
					triggerForTest("RebuildTask", &model.TaskTrigger{
						Name:      "atTime",
						Arguments: []*model.Argument{{Name: "startAt", Type: codegentest.LocalDateTimeType()}},
					}),
				},
			},
		},
		Events: []*model.Data{
			{
				Name: "UserCreated",
				Members: []*model.DataMember{
					{Name: "userId", Type: codegentest.StringType()},
				},
			},
		},
	})

	if err := generateFixture(pkg, golang.Option{Out: goOutDir}); err != nil {
		t.Fatal(err)
	}

	goSchemaContent := readFileForTest(t, filepath.Join(goOutDir, "schema.go"))
	assertNoBlankLineInsideDeclaration(t, goSchemaContent)
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

func TestGeneratorGoRendersWebMountInSpecAndSchema(t *testing.T) {
	for _, path := range []string{"", "/", "/portal/v1-assets/"} {
		pkg := newModelDomainForTest(t, model.DomainSpec{
			Name:   "demo.web",
			Actors: []*model.Actor{{Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}}},
			Webs:   []*model.Web{{Name: "PortalWeb", MountPath: path, Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}}},
		})
		out := filepath.Join(t.TempDir(), "skeled")
		if err := generateFixture(pkg, golang.Option{Out: out}); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"web.go", "schema.go"} {
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
 "go.yorun.ai/vine/core/skel"
)
func TestAuthModes(t *testing.T) {
 if _DomainSchema.Services[0].AuthMode != skel.AuthModeRequired { t.Fatal("wrong service auth") }
 methods:=_DomainSchema.Services[0].Methods
 want:=map[string]skel.AuthMode{"profile":skel.AuthModeInherit,"browse":skel.AuthModeOptional,"login":skel.AuthModeAnonymous}
 for _,method:=range methods { if method.AuthMode!=want[method.Name] { t.Fatalf("method %s: %s",method.Name,method.AuthMode) } }
 webs:=map[string]skel.AuthMode{"RequiredWeb":skel.AuthModeRequired,"OptionalWeb":skel.AuthModeOptional,"AnonymousWeb":skel.AuthModeAnonymous,"OffWeb":skel.AuthModeOff}
 for _,web:=range _DomainSchema.Webs { if web.AuthMode!=webs[web.Name] { t.Fatalf("web %s: %s",web.Name,web.AuthMode) } }
}
`)
	testutil.UseLocalSkel(t, out)
	testutil.Go(t, out, "test", "-mod=mod", "./...")
}

func TestApiBackendSchemaAndClientBoundary(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "order.skel")
	writeFileForTest(t, entry, `domain demo.order
actor TestActor { via client {} }
api service OrderApiService { for TestActor via client method ping {} }
pub service BackendService { method ping {} }
`)
	input := api.Input{SkelIn: entry}
	out := filepath.Join(root, "server")
	option := api.GolangOption{CompilerVersion: "v0.0.0-dev", AsModule: true, Module: "example.com/orderserver", Out: out}
	if _, err := api.CompileGolang(input, option); err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(filepath.Join(out, "schema.go"))
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
