package descriptor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/testutil"
	"go.yorun.ai/skel/schema"
)

// Type-check the emitted metadata against the real public descriptor package.
// Runtime registration is tested separately by the Vine integration tests.
func TestGeneratedDescriptorCompilesAgainstPublicTypes(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	source := portableDescriptorSource + `
pub enum Status { READY }
pub data Page<TItem> { items: list<TItem> }
pub data ScalarValues { id: uuid payload: json }
pub config AppConfig eternal { pageSize: int }
pub config DynamicConfig instant { enabled: bool }
pub event ChangedEvent { payload { node: Node } }
web PortalWeb {
    for ClientActor via client
    auth required
    mount /api
}
task RefreshTask { trigger manually { input { page: Page<Node> } } }
`
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	parsed, err := compiler.Compile(compiler.Option{SkelIn: path})
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "generated")
	gen := newGen(Option{Domain: parsed.Domain, View: mustView(t, view.ModeFull, parsed.Domain), Mode: view.ModeFull, PackageName: "generated", Out: out})
	if err := gen.gen(); err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(out, descriptorGoFilename))
	if err != nil {
		t.Fatal(err)
	}
	for _, literal := range []string{"descriptor.ConfigLifecycleEternal", "descriptor.ConfigLifecycleInstant"} {
		codegentest.AssertGoSourceContains(t, string(generated), "Lifecycle: "+literal)
	}
	fileset := token.NewFileSet()
	file, err := parser.ParseFile(fileset, filepath.Join(out, descriptorGoFilename), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var declarations []ast.Decl
	registrations := 0
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "init" {
			var registration bytes.Buffer
			if err := format.Node(&registration, fileset, function); err != nil {
				t.Fatal(err)
			}
			if strings.Join(strings.Fields(registration.String()), " ") != "func init() { skel.RegisterDomainDescriptor(_DomainDescriptor) }" {
				t.Fatalf("unexpected descriptor registration: %s", registration.String())
			}
			registrations++
			continue
		}
		if imports, ok := declaration.(*ast.GenDecl); ok && imports.Tok == token.IMPORT {
			var retained []ast.Spec
			for _, spec := range imports.Specs {
				if spec.(*ast.ImportSpec).Path.Value != `"go.yorun.ai/vine/core/skel"` {
					retained = append(retained, spec)
				}
			}
			imports.Specs = retained
		}
		declarations = append(declarations, declaration)
	}
	if registrations != 1 {
		t.Fatalf("expected one registration, got %d", registrations)
	}
	file.Decls = declarations
	var metadata bytes.Buffer
	if err := format.Node(&metadata, fileset, file); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, descriptorGoFilename), metadata.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "go.mod"), []byte("module example.com/descriptortest\n\ngo 1.27.0\n\nrequire go.yorun.ai/skel v0.0.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	testutil.UseLocalSkel(t, out)
	if err := os.WriteFile(filepath.Join(out, "descriptor_test.go"), []byte(`package generated
import (
    "testing"
    "encoding/json"
    "strings"
    "go.yorun.ai/skel/descriptor"
)
func TestEffectivePolicies(t *testing.T) {
    if err := descriptor.ValidateEffectivePolicy(_DomainDescriptor); err != nil { t.Fatal(err) }
    encoded, err := json.Marshal(_DomainDescriptor)
    if err != nil { t.Fatal(err) }
    var decoded descriptor.Domain
    if err := json.Unmarshal(encoded, &decoded); err != nil { t.Fatal(err) }
    if decoded.Name != "demo" { t.Fatalf("domain name: %q", decoded.Name) }
    if err := descriptor.ValidateEffectivePolicy(&decoded); err != nil { t.Fatal(err) }
}
func TestResourceCheckReferences(t *testing.T) {
    resource := _DomainDescriptor.Resources[0]
    if resource.CheckMethod(resource.Checks[0]) != resource.CheckService.Methods[0] { t.Fatal("resource method reference") }
    for index, action := range resource.Actions {
        method := resource.CheckMethod(action.Checks[0])
        if method == nil || method != resource.CheckService.Methods[index+1] { t.Fatal("action method reference") }
        if len(method.Arguments) != 2 || method.Arguments[0].Name != "code" || method.Arguments[1].Name != []string{"ownerId", "editorId"}[index] { t.Fatal("action arguments") }
    }
}
func TestConfigLifecycles(t *testing.T) {
    want := map[string]descriptor.ConfigLifecycle{
        "AppConfig": descriptor.ConfigLifecycleEternal,
        "DynamicConfig": descriptor.ConfigLifecycleInstant,
    }
    if len(_DomainDescriptor.Configs) != len(want) { t.Fatalf("configs: %d", len(_DomainDescriptor.Configs)) }
    encoded, err := json.Marshal(_DomainDescriptor)
    if err != nil { t.Fatal(err) }
    for _, lifecycle := range []string{"eternal", "instant"} {
        if !strings.Contains(string(encoded), "\"lifecycle\":\""+lifecycle+"\"") { t.Fatalf("missing lowercase lifecycle %q: %s", lifecycle, encoded) }
    }
    var decoded descriptor.Domain
    if err := json.Unmarshal(encoded, &decoded); err != nil { t.Fatal(err) }
    for _, config := range decoded.Configs {
        expected, ok := want[config.Name]
        if !ok || config.Lifecycle != expected { t.Fatalf("config %s lifecycle: %q", config.Name, config.Lifecycle) }
        delete(want, config.Name)
    }
    if len(want) != 0 { t.Fatalf("missing configs: %v", want) }
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	testutil.Go(t, out, "test", "-mod=mod", "./...")
}

func TestGenDescriptorGoRendersActorCapabilities(t *testing.T) {
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{
			{
				Name: "ClientActor",
				Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)},
				Auth: new(schema.ActorAuth{
					IdentifierField: "userId",
					Credential: &schema.Data{
						Name: "ClientActorCredential",
						Members: []*schema.DataMember{
							{Name: "token", Type: codegentest.StringType()},
						},
					},
					Info: &schema.Data{
						Name: "ClientActorInfo",
						Members: []*schema.DataMember{
							{Name: "userId", Type: codegentest.IntType()},
						},
					},
				}),
			},
			{
				Name: "AnonymousActor",
				Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)},
			},
		},
	})

	outputDir := filepath.Join(t.TempDir(), "skeled")
	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         outputDir,
	})
	gen.gen()

	content, err := os.ReadFile(filepath.Join(outputDir, descriptorGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	codegentest.AssertGoSourceContains(t, string(content), `IdentifierField: "userId"`)
	codegentest.AssertGoSourceContains(t, string(content), "Auth: &descriptor.ActorAuth{")
	codegentest.AssertGoSourceContains(t, string(content), `MethodName: "auth"`)
	codegentest.AssertGoSourceContains(t, string(content), `{Name: "AnonymousActor", SkelName: "", Hash: "", Vias: []descriptor.ActorViaKind{descriptor.ActorViaClient}}`)
	for _, old := range []string{"AuthEnabled:", "PermissionEnabled:", "AuthCredential:", "AuthService:"} {
		if strings.Contains(string(content), old) {
			t.Fatalf("generated descriptor retains flattened actor field %s", old)
		}
	}
}

func TestGenDescriptorGoRendersDeprecatedFields(t *testing.T) {
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{{
			Name:             "User",
			Deprecated:       true,
			DeprecatedReason: "Use Profile instead",
			Members: []*schema.DataMember{{
				Name:             "legacyId",
				Type:             codegentest.StringType(),
				Deprecated:       true,
				DeprecatedReason: "Use id instead",
			}},
		}},
	})
	outputDir := filepath.Join(t.TempDir(), "skeled")
	newGen(Option{
		Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull,
		PackageName: "skeled", Out: outputDir,
	}).gen()

	content, err := os.ReadFile(filepath.Join(outputDir, descriptorGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	generated := string(content)
	if strings.Count(generated, "Deprecated:") < 2 {
		t.Fatalf("expected data and member deprecation flags, got:\n%s", generated)
	}
	for _, reason := range []string{"Use Profile instead", "Use id instead"} {
		if !strings.Contains(generated, `"`+reason+`"`) {
			t.Fatalf("expected reason %q, got:\n%s", reason, generated)
		}
	}
}

func TestGenDescriptorGoRendersWebAuthModes(t *testing.T) {
	for _, mode := range []schema.AuthMode{schema.AuthModeUnset, schema.AuthModeRequired, schema.AuthModeOptional, schema.AuthModeAnonymous, schema.AuthModeOff, schema.AuthModeAuth, schema.AuthModeNoAuth} {
		t.Run(string(mode), func(t *testing.T) {
			pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{Name: "demo.user", Webs: []*schema.Web{{Name: "ConsoleWeb", AuthMode: mode}}})
			out := filepath.Join(t.TempDir(), "skeled")
			newGen(Option{Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull, PackageName: "skeled", Out: out}).gen()
			content, err := os.ReadFile(filepath.Join(out, descriptorGoFilename))
			if err != nil {
				t.Fatal(err)
			}
			expected := mode
			switch mode {
			case schema.AuthModeUnset, schema.AuthModeAuth:
				expected = schema.AuthModeRequired
			case schema.AuthModeNoAuth:
				expected = schema.AuthModeOff
			}
			literal, err := renderAuthModeLiteral(descriptor.AuthMode(expected))
			if err != nil {
				t.Fatal(err)
			}
			codegentest.AssertGoSourceContains(t, string(content), "AuthMode: "+literal)

		})
	}
}
