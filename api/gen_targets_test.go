package api_test

import (
	"context"
	"go/ast"
	stdparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/codegen"
	"go.yorun.ai/skel/internal/testutil"
	"go.yorun.ai/skel/schema"
)

func TestBuiltinGeneratorsPreserveNormalizedMethodAuth(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "domain.skel")
	source := `domain demo
pub actor ClientActor { via client {} }
pub service BackendService { method read {} }
api service EntryApiService { for ClientActor auth required method read {} }
`
	var inputs []codegen.Input
	for _, normalized := range []bool{false, true} {
		parsed, err := api.Parse(api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte(source)}})
		if err != nil {
			t.Fatal(err)
		}
		for _, service := range parsed.Domain.Services() {
			clientApi := service.Api
			if normalized {
				for _, method := range service.Methods {
					method.AuthMode = method.NormalizedAuth()
				}
			}
			if service.Api != clientApi {
				t.Fatal("normalizing inherited method auth changed service selection")
			}
		}
		input, err := codegen.Prepare(parsed.Domain, codegen.Selection{})
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, input)
	}
	for _, target := range []string{"go", "go-api", "ts", "skel"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(root, target, "generated")
			var generator codegen.Generator
			var err error
			switch target {
			case "go", "go-api":
				generator, err = api.NewGolangGenerator(api.GolangOption{Out: out, CompilerVersion: "v0.0.0-dev", ApiOnly: target == "go-api"})
			case "ts":
				generator, err = api.NewTypeScriptGenerator(api.TypeScriptOption{Out: out, ApiOnly: true})
			case "skel":
				generator, err = api.NewSkeletonGenerator(api.SkeletonOption{Out: out, PubOnly: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := codegen.Generate(t.Context(), inputs[0], generator)
			if err != nil {
				t.Fatal(err)
			}
			after, err := codegen.Generate(t.Context(), inputs[1], generator)
			if err != nil {
				t.Fatal(err)
			}
			if len(before) == 0 || !reflect.DeepEqual(before, after) {
				t.Fatal("normalizing inherited method auth changed generated files")
			}
		})
	}
}

func TestOptionalActorCredentialGeneration(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeTestFile(t, filepath.Join(source, "domain.skel"), "domain demo.auth")
	writeTestFile(t, filepath.Join(source, "actor.skel"), `domain demo.auth
pub actor UserActor {
    via client {}
    auth {
        credential {
            token: string
            session: string?
        }
        info { userId: string }
    }
}
`)
	input := api.Input{SkelIn: source}
	goOut, tsOut, skelOut := filepath.Join(root, "golang"), filepath.Join(root, "ts"), filepath.Join(root, "skel")
	if _, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: goOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, Out: tsOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := api.CompileSkeleton(input, api.SkeletonOption{Out: skelOut, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	goFile, err := stdparser.ParseFile(token.NewFileSet(), filepath.Join(goOut, "actor.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	ast.Inspect(goFile, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "UserActorCredential" {
			return true
		}
		fields := spec.Type.(*ast.StructType).Fields.List
		for _, field := range fields {
			if len(field.Names) == 0 {
				continue
			}
			if field.Names[0].Name == "Session" {
				pointer, ok := field.Type.(*ast.StarExpr)
				if !ok {
					t.Fatal("optional credential must generate a Go pointer")
				}
				if name, ok := pointer.X.(*ast.Ident); !ok || name.Name != "string" {
					t.Fatal("optional credential must point to string")
				}
				found = true
			}
		}
		return false
	})
	if !found {
		t.Fatal("missing generated optional credential")
	}
	roundTrip, err := api.Parse(api.Input{SkelIn: skelOut})
	if err != nil {
		t.Fatal(err)
	}
	credential := roundTrip.Domain.Actors()[0].Auth.Credential
	if credential.Members[0].Type.Nullable || !credential.Members[1].Type.Nullable {
		t.Fatal("public Skel did not preserve required and optional credential fields")
	}
}

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
api service UserApiService { auth required
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
	if _, err := api.CompileGolang(api.Input{SkelIn: input}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, AsModule: true, Module: "example.com/permissions"}); err != nil {
		t.Fatal(err)
	}
	resource := strings.Join(strings.Fields(readTestFile(t, filepath.Join(out, "resource.go"))), " ")
	if !strings.Contains(resource, `UserReadPermission string = "demo.User:read"`) || !strings.Contains(resource, "code string") {
		t.Fatalf("expected string constant and check parameter: %s", resource)
	}
	tsOut := filepath.Join(root, "ts")
	if _, err := api.CompileTypeScript(api.Input{SkelIn: input}, api.TypeScriptOption{ApiOnly: true, Out: tsOut}); err != nil {
		t.Fatal(err)
	}
	ts := strings.Join(strings.Fields(readTestFile(t, filepath.Join(tsOut, "data.ts"))), " ")
	if !strings.Contains(ts, "code: string") || !strings.Contains(ts, "optional: string | null") {
		t.Fatalf("expected string TypeScript fields: %s", ts)
	}
	public := filepath.Join(root, "public")
	if _, err := api.CompileSkeleton(api.Input{SkelIn: input}, api.SkeletonOption{Out: public, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := api.CompileGolang(api.Input{SkelIn: public}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: filepath.Join(root, "roundtrip")}); err != nil {
		t.Fatal(err)
	}
	testutil.UseLocalSkel(t, out)
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
api service UserApiService { auth required
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
			if _, err := api.CompileGolang(api.Input{SkelIn: input}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out}); err != nil {
				t.Fatal(err)
			}
			generatedSchema := strings.Join(strings.Fields(readTestFile(t, filepath.Join(out, "descriptor.go"))), " ")
			for _, want := range []string{`CodeArgumentName: "code2"`, `CodeArgumentName: "code"`, `Name: "code", JsonPath: "code"`, `Name: "code1", JsonPath: "code1"`} {
				if !strings.Contains(generatedSchema, want) {
					t.Fatalf("missing %s in schema:\n%s", want, generatedSchema)
				}
			}
			resource := strings.Join(strings.Fields(readTestFile(t, filepath.Join(out, "resource.go"))), " ")
			for _, want := range []string{`json:"code2" skel:"index(0)"`, `json:"code" skel:"index(1)"`, `json:"code1" skel:"index(2)"`} {
				if !strings.Contains(resource, want) {
					t.Fatalf("missing %s in resource:\n%s", want, resource)
				}
			}
			public := filepath.Join(root, "public")
			if _, err := api.CompileSkeleton(api.Input{SkelIn: input}, api.SkeletonOption{Out: public, PubOnly: true}); err != nil {
				t.Fatal(err)
			}
			out2 := filepath.Join(root, "roundtrip")
			if _, err := api.CompileGolang(api.Input{SkelIn: public}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out2}); err != nil {
				t.Fatal(err)
			}
			if got := readTestFile(t, filepath.Join(out2, "resource.go")); !strings.Contains(got, `json:"code2" skel:"index(0)"`) {
				t.Fatalf("public round trip lost injection name: %s", got)
			}
			// Resolve the same check through an imported public resource.
			consumer := filepath.Join(root, "consumer.skel")
			source := `domain consumer
import demo
actor ClientActor { via client {} }
api service ConsumerApiService { auth required
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
			if _, err := api.CompileGolang(api.Input{SkelIn: consumer, SkelImports: map[string]string{"demo": public}}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: consumerOut, Imports: map[string]string{"demo": "example.com/demo"}}); err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(strings.Fields(readTestFile(t, filepath.Join(consumerOut, "descriptor.go"))), " "); !strings.Contains(got, `CodeArgumentName: "code2"`) {
				t.Fatalf("imported check lost injection name: %s", got)
			}
		})
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

func TestGenerationReusesSemanticSchemaAcrossTargetsAndGoroutines(t *testing.T) {
	root := t.TempDir()
	shared, consumer := filepath.Join(root, "shared.skel"), filepath.Join(root, "consumer.skel")
	writeTestFile(t, shared, "domain shared.user\npub enum State { READY }\npub data Value { state: State }\n")
	writeTestFile(t, consumer, "domain consumer\nimport shared.user\npub data Box<TItem> { value: TItem }\nactor TestActor { via client {} }\napi service ReadApiService { auth required  for TestActor via client method get { output Box<shared.user.Value> } }\n")
	compiled, err := api.Parse(api.Input{SkelIn: consumer, SkelImports: map[string]string{"shared.user": shared}})
	if err != nil {
		t.Fatal(err)
	}
	domain := compiled.Domain
	before := map[*schema.Type]schema.Type{}
	roots := (codegen.Declarations{Data: domain.Data(), Services: domain.Services()}).TypeRoots(true)
	if err := codegen.WalkTypes(roots, func(kind *schema.Type) error {
		before[kind] = *kind
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	renderGo := func(out string) error {
		return api.GenerateGolang(domain, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, ApiOnly: true, AsModule: true, Module: "example.com/consumerapi", ModulePrefix: "example.com"})
	}
	renderTS := func(out string) error {
		return api.GenerateTypeScript(domain, api.TypeScriptOption{ApiOnly: true, Out: out, AsModule: true, ModuleScope: "@example"})
	}
	first, ts, second := filepath.Join(root, "first"), filepath.Join(root, "ts"), filepath.Join(root, "second")
	if err := renderGo(first); err != nil {
		t.Fatal(err)
	}
	if err := renderTS(ts); err != nil {
		t.Fatal(err)
	}
	if err := renderGo(second); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(apiOutputSnapshot(t, first), apiOutputSnapshot(t, second)) {
		t.Fatal("Go output depends on previous target generation")
	}
	var group sync.WaitGroup
	failures := make(chan error, 4)
	for i := range 4 {
		out := filepath.Join(root, string(rune('a'+i)))
		group.Go(func() {
			if i%2 == 0 {
				failures <- renderGo(out)
			} else {
				failures <- renderTS(out)
			}
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	for kind, original := range before {
		if !reflect.DeepEqual(original, *kind) {
			t.Fatalf("generation mutated semantic type %s", kind.SkelName)
		}
	}
}
