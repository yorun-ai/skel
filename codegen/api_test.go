package codegen_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/codegen"
	"go.yorun.ai/skel/schema"
)

func parseDomain(t *testing.T) *schema.Domain {
	t.Helper()
	root := t.TempDir()
	main, dep := filepath.Join(root, "main.skel"), filepath.Join(root, "shared.skel")
	parsed, err := api.Parse(api.Input{SkelIn: main, SkelImports: map[string]string{"shared": dep}, Sources: map[string][]byte{
		main: []byte(`domain demo
import shared
pub data Box<TItem> { value: TItem? items: list<TItem> }
pub data Result { user: shared.User page: Box<string> }
data Node { next: Node? }
data Hidden { value: int }
`),
		dep: []byte("domain shared\npub data User { name: string }\n"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	return parsed.Domain
}

func TestInputSelectionAndSemanticQueries(t *testing.T) {
	domain := parseDomain(t)
	selection := codegen.Selection{Surface: codegen.SurfaceAPI, API: codegen.ApiFilter{Prune: true, Types: []string{"demo.Result"}}}
	input, err := codegen.Prepare(domain, selection)
	if err != nil {
		t.Fatal(err)
	}
	declarations := input.Declarations()
	var names []string
	for _, data := range declarations.Data {
		names = append(names, data.Name)
	}
	if !reflect.DeepEqual(names, []string{"Box", "Result"}) || len(declarations.Services) != 0 {
		t.Fatalf("wrong selection: %v", names)
	}
	if deps := input.ExternalDomains(); !reflect.DeepEqual(deps, []string{"shared"}) {
		t.Fatalf("dependencies: %v", deps)
	}
	if input.Schema() != domain || input.FindData("demo.Hidden") == nil || input.FindData("shared.User") == nil {
		t.Fatal("complete semantic graph unavailable")
	}
	// Output collections and selection options are defensive copies; schema nodes
	// intentionally retain their semantic identity across views.
	declarations.Data[0] = nil
	selection.API.Types[0] = "demo.Hidden"
	selected := input.Selection()
	selected.API.Types[0] = "demo.Hidden"
	if input.Declarations().Data[0] == nil || input.Selection().API.Types[0] != "demo.Result" {
		t.Fatal("view state leaked")
	}
	full, err := input.Select(codegen.Selection{})
	if err != nil || len(full.Declarations().Data) != 4 {
		t.Fatalf("full selection: %+v, %v", full.Declarations(), err)
	}
	public, err := input.Select(codegen.Selection{Surface: codegen.SurfacePublic})
	if err != nil || len(public.Declarations().Data) != 2 {
		t.Fatalf("public selection: %+v, %v", public.Declarations(), err)
	}
	if _, err := input.Select(codegen.Selection{Surface: "unknown"}); err == nil {
		t.Fatal("invalid surface accepted")
	}
	if _, err := input.Select(codegen.Selection{Surface: codegen.SurfaceAPI, API: codegen.ApiFilter{Prune: true, Types: []string{"demo.Missing"}}}); err == nil {
		t.Fatal("unknown root accepted")
	}
}

func TestGenericInstantiationAndRecursiveTraversal(t *testing.T) {
	input, err := codegen.Prepare(parseDomain(t), codegen.Selection{})
	if err != nil {
		t.Fatal(err)
	}
	result := input.FindData("demo.Result")
	var page *schema.Type
	for _, member := range result.Members {
		if member.Name == "page" {
			page = member.Type
		}
	}
	members, err := codegen.InstantiateMembers(page)
	if err != nil {
		t.Fatal(err)
	}
	if members[0].Type.Scalar != schema.ScalarString || !members[0].Type.Nullable || members[1].Type.List.Element.Scalar != schema.ScalarString {
		t.Fatalf("substitution failed: %+v", members)
	}
	if page.Data.Members[0].Type.Kind != schema.TypeKindTypeParameter || page.TypeArguments[0].Nullable {
		t.Fatal("substitution changed semantic data")
	}
	node := input.FindData("demo.Node")
	visits := 0
	if err := codegen.WalkTypeGraph(new(schema.Type{Kind: schema.TypeKindData, Data: node}), func(*schema.Type) error { visits++; return nil }); err != nil || visits != 2 {
		t.Fatalf("recursive visits=%d, error=%v", visits, err)
	}
	stop := errors.New("stop")
	if err := codegen.WalkTypes(input.TypeRoots(), func(*schema.Type) error { return stop }); !errors.Is(err, stop) {
		t.Fatalf("walk error=%v", err)
	}
}

func TestPrepareRejectsUnresolvedSchemas(t *testing.T) {
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Data: []*schema.Data{{Name: "Pending", Kind: schema.DataKindData, Members: []*schema.DataMember{{Name: "value", Type: new(schema.Type{Kind: schema.TypeKindUnresolvedReference, SkelName: "Missing"})}}}}})
	if _, err := codegen.Prepare(domain, codegen.Selection{}); err == nil {
		t.Fatal("unresolved schema accepted")
	}
	if _, err := codegen.Generate(t.Context(), codegen.Input{}, codegen.GeneratorFunc(func(context.Context, codegen.Input) ([]codegen.File, error) {
		t.Fatal("invalid input reached generator")
		return nil, nil
	})); err == nil {
		t.Fatal("zero input accepted")
	}
}

func TestPrepareRefreshesEffectivePoliciesBeforeBorrowing(t *testing.T) {
	method := new(schema.Method{Name: "read", AuthMode: schema.AuthModeInherit})
	service := new(schema.Service{Name: "Files", AuthMode: schema.AuthModeOptional, Methods: []*schema.Method{method}})
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Services: []*schema.Service{service}})
	if _, err := codegen.Prepare(domain, codegen.Selection{}); err != nil {
		t.Fatal(err)
	}
	if method.EffectiveAuthMode != schema.AuthModeOptional {
		t.Fatal("missing effective policy for programmatic schema")
	}
	// No Input is retained while editing source declarations.
	service.AuthMode = schema.AuthModeRequired
	input, err := codegen.Prepare(domain, codegen.Selection{})
	if err != nil || method.EffectiveAuthMode != schema.AuthModeRequired {
		t.Fatalf("stale effective policy was not refreshed: %v", err)
	}
	for range 4 {
		t.Run("shared graph", func(t *testing.T) {
			t.Parallel()
			if _, err := codegen.Prepare(input.Schema(), codegen.Selection{}); err != nil {
				t.Fatal(err)
			}
			if err := schema.ValidateEffectivePolicy(input.Schema()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestThirdPartyBindingAndManagedFiles(t *testing.T) {
	input, err := codegen.Prepare(parseDomain(t), codegen.Selection{Surface: codegen.SurfacePublic})
	if err != nil {
		t.Fatal(err)
	}
	// This binding and its test import only public Skel packages.
	binding := codegen.GeneratorFunc(func(ctx context.Context, in codegen.Input) ([]codegen.File, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var source strings.Builder
		for _, data := range in.Declarations().Data {
			source.WriteString("class " + data.Name + ":\n    pass\n")
		}
		return []codegen.File{{Path: "types.py", Content: source.String(), CommentPrefix: "#"}}, nil
	})
	files, err := codegen.Generate(t.Context(), input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(files[0].Content, "class Result:") {
		t.Fatal("missing generated data")
	}
	out := filepath.Join(t.TempDir(), "generated")
	if err := codegen.Run(t.Context(), input, binding, map[string]string{"": out}); err != nil {
		t.Fatal(err)
	}
	manual := filepath.Join(out, "manual.py")
	if err := os.WriteFile(manual, []byte("# handwritten\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(filepath.Join(out, "types.py"))
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("binding failed")
	bad := codegen.GeneratorFunc(func(context.Context, codegen.Input) ([]codegen.File, error) {
		return []codegen.File{{Path: "types.py", Content: "partial", CommentPrefix: "#"}}, failure
	})
	if err := codegen.Run(t.Context(), input, bad, map[string]string{"": out}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(out, "types.py"))
	if string(after) != string(old) {
		t.Fatal("failed generation changed output")
	}
	if err := codegen.WriteFiles(t.Context(), nil, map[string]string{"": out}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "types.py")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale Python file retained: %v", err)
	}
	if content, err := os.ReadFile(manual); err != nil || string(content) != "# handwritten\n" {
		t.Fatalf("manual file changed: %v", err)
	}
}

func TestBuiltinGeneratorsUseSDKWithoutPublishing(t *testing.T) {
	domain := parseDomain(t)
	input, err := codegen.Prepare(domain, codegen.Selection{})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, target := range []string{"go", "ts", "skel", "split"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(root, target+"_output")
			var generator codegen.Generator
			var err error
			switch target {
			case "go":
				generator, err = api.NewGolangGenerator(api.GolangOption{Out: out, ApiOnly: true, Imports: map[string]string{"shared": "example.com/sharedapi"}})
			case "split":
				generator, err = api.NewGolangGenerator(api.GolangOption{Out: out, PubOut: filepath.Join(root, "pub"), CompilerVersion: "v0.0.0-dev", AsModule: true, Module: "example.com/demo", ModulePrefix: "example.com"})
			case "ts":
				generator, err = api.NewTypeScriptGenerator(api.TypeScriptOption{Out: out, ApiOnly: true, Imports: map[string]string{"shared": "@example/shared"}})
			case "skel":
				generator, err = api.NewSkeletonGenerator(api.SkeletonOption{Out: out, PubOnly: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			first, err := codegen.Generate(t.Context(), input, generator)
			if err != nil {
				t.Fatal(err)
			}
			second, err := codegen.Generate(t.Context(), input, generator)
			if err != nil {
				t.Fatal(err)
			}
			if len(first) == 0 || !reflect.DeepEqual(first, second) {
				t.Fatal("nondeterministic files")
			}
			if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("render wrote output: %v", err)
			}
			if target == "split" {
				pub := false
				for _, file := range first {
					pub = pub || file.Target == "pub"
				}
				if !pub {
					t.Fatal("public target missing")
				}
			}
		})
	}
	// The source qualifier is retained after Go and TS choose different imports.
	for _, member := range input.FindData("demo.Result").Members {
		if member.Type.ExternalDomain == "shared" && member.Type.ExternalAlias != "shared" {
			t.Fatal("target qualifier leaked into schema")
		}
	}
}
