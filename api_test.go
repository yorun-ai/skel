package skelc_test

import (
	"errors"
	"fmt"
	"go/ast"
	stdparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"go.yorun.ai/skelc"
	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/testutil"
	"go.yorun.ai/skelc/model"
	"golang.org/x/mod/modfile"
)

func TestParseStrictMigrationRules(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "order.skel")
	writeTestFile(t, entry, "domain demo.order\nservice OrderService { method ping {} }\n")
	result, err := skelc.Parse(skelc.Input{SkelIn: entry})
	if err != nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].Severity != skelc.DiagnosticSeverityWarning {
		t.Fatalf("unexpected compatible parse: %+v, %v", result, err)
	}
	_, err = skelc.Parse(skelc.Input{SkelIn: entry, Strict: true})
	var diagnostics skelc.Diagnostics
	if !errors.As(err, &diagnostics) || len(diagnostics) != 1 || diagnostics[0].Code != skelc.DiagnosticCodeServiceModifier || diagnostics[0].Severity != skelc.DiagnosticSeverityError {
		t.Fatalf("unexpected strict diagnostics: %v", err)
	}
}

func ExampleParse() {
	skelDir, err := os.MkdirTemp("", "skelc-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(skelDir)

	if err := os.WriteFile(filepath.Join(skelDir, "domain.skel"), []byte("domain demo.user"), 0o644); err != nil {
		panic(err)
	}

	result, err := skelc.Parse(skelc.Input{SkelIn: skelDir})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Domain.Name())

	// Output: demo.user
}

func TestParseExposesSemanticModel(t *testing.T) {
	skelDir := t.TempDir()
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")

	result, err := skelc.Parse(skelc.Input{SkelIn: skelDir})
	if err != nil {
		t.Fatalf("parse Skel: %v", err)
	}
	var domain *model.Domain = result.Domain
	if domain.Name() != "demo.user" {
		t.Fatalf("unexpected domain: %s", domain.Name())
	}
}

func TestDiagnosticAliasesMatchPublicDiagnosticPackage(t *testing.T) {
	tests := []struct {
		name    string
		alias   string
		defined string
	}{
		{"severity error", string(skelc.DiagnosticSeverityError), string(diagnostic.SeverityError)},
		{"severity warning", string(skelc.DiagnosticSeverityWarning), string(diagnostic.SeverityWarning)},
		{"syntax unexpected", skelc.DiagnosticCodeSyntaxUnexpected, diagnostic.CodeSyntaxUnexpected},
		{"syntax eof", skelc.DiagnosticCodeSyntaxEOF, diagnostic.CodeSyntaxEOF},
		{"syntax finalize", skelc.DiagnosticCodeSyntaxFinalize, diagnostic.CodeSyntaxFinalize},
		{"semantic validation", skelc.DiagnosticCodeSemanticValidation, diagnostic.CodeSemanticValidation},
		{"semantic duplicate", skelc.DiagnosticCodeSemanticDuplicate, diagnostic.CodeSemanticDuplicate},
		{"semantic naming", skelc.DiagnosticCodeSemanticNaming, diagnostic.CodeSemanticNaming},
		{"semantic reference", skelc.DiagnosticCodeSemanticReference, diagnostic.CodeSemanticReference},
		{"semantic warning", skelc.DiagnosticCodeSemanticWarning, diagnostic.CodeSemanticWarning},
		{"service modifier", skelc.DiagnosticCodeServiceModifier, diagnostic.CodeServiceModifier},
		{"service client rules", skelc.DiagnosticCodeServiceClientRules, diagnostic.CodeServiceClientRules},
		{"import missing", skelc.DiagnosticCodeImportMissing, diagnostic.CodeImportMissing},
		{"import cycle", skelc.DiagnosticCodeImportCycle, diagnostic.CodeImportCycle},
		{"domain missing", skelc.DiagnosticCodeDomainMissing, diagnostic.CodeDomainMissing},
		{"domain mismatch", skelc.DiagnosticCodeDomainMismatch, diagnostic.CodeDomainMismatch},
		{"domain file content", skelc.DiagnosticCodeDomainFileContent, diagnostic.CodeDomainFileContent},
		{"domain decorator", skelc.DiagnosticCodeDomainDecorator, diagnostic.CodeDomainDecorator},
		{"loader directory", skelc.DiagnosticCodeLoaderDirectory, diagnostic.CodeLoaderDirectory},
		{"loader hidden file", skelc.DiagnosticCodeLoaderHiddenFile, diagnostic.CodeLoaderHiddenFile},
		{"loader unsupported", skelc.DiagnosticCodeLoaderUnsupported, diagnostic.CodeLoaderUnsupported},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.alias != test.defined {
				t.Fatalf("alias = %q, want %q", test.alias, test.defined)
			}
		})
	}
}

func TestPublicPackagesRespectDependencyBoundaries(t *testing.T) {
	const internalPrefix = "go.yorun.ai/skelc/internal/"
	rules := []struct {
		directory string
		forbidden func(string) bool
	}{
		{
			directory: "internal/parser",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) &&
					path != internalPrefix+"parser/grammar" &&
					path != internalPrefix+"model"
			},
		},
		{
			directory: "internal/analyzer",
			forbidden: func(path string) bool {
				for _, prefix := range []string{"compiler", "hasher", "loader", "lsp", "codegen"} {
					if strings.HasPrefix(path, internalPrefix+prefix) {
						return true
					}
				}
				return false
			},
		},
		{
			directory: "internal/hasher",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"model"
			},
		},
		{
			directory: "model",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"model"
			},
		},
		{
			directory: "diagnostic",
			forbidden: func(path string) bool { return strings.HasPrefix(path, internalPrefix) },
		},
		{
			directory: "schema",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"schema"
			},
		},
		{
			directory: "internal/codegen/common",
			forbidden: targetCodegenImport,
		},
		{
			directory: "internal/codegen/output",
			forbidden: targetCodegenImport,
		},
	}

	for _, rule := range rules {
		t.Run(filepath.ToSlash(rule.directory), func(t *testing.T) {
			inspectProductionImports(t, rule.directory, rule.forbidden)
		})
	}
}

func targetCodegenImport(path string) bool {
	for _, target := range []string{"golang", "skeleton", "typescript"} {
		if strings.HasPrefix(path, "go.yorun.ai/skelc/internal/codegen/"+target) {
			return true
		}
	}
	return false
}

func inspectProductionImports(t *testing.T, directory string, forbidden func(string) bool) {
	t.Helper()
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := stdparser.ParseFile(token.NewFileSet(), path, nil, stdparser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			importDecl, ok := declaration.(*ast.GenDecl)
			if !ok || importDecl.Tok != token.IMPORT {
				continue
			}
			for _, spec := range importDecl.Specs {
				pathValue, err := strconv.Unquote(spec.(*ast.ImportSpec).Path.Value)
				if err != nil {
					return err
				}
				if forbidden(pathValue) {
					t.Errorf("%s imports forbidden dependency %s", path, pathValue)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseReturnsStructuredWarnings(t *testing.T) {
	skelDir := t.TempDir()
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")
	writeTestFile(t, filepath.Join(skelDir, ".ignored.skel"), "domain ignored")

	result, err := skelc.Parse(skelc.Input{SkelIn: skelDir})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != "loader.ignored-hidden-file" || diagnostic.Severity != skelc.DiagnosticSeverityWarning {
		t.Fatalf("unexpected warning: %+v", diagnostic)
	}
}

func TestGenerateGolang(t *testing.T) {
	skelDir := t.TempDir()
	goOut := filepath.Join(t.TempDir(), "generated")
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")

	result, err := skelc.CompileGolang(
		skelc.Input{SkelIn: skelDir},
		skelc.GolangOption{Out: goOut, CompilerVersion: "v1.2.3"},
	)
	if err != nil {
		t.Fatalf("generate Go: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
	}
	assertTestFileExists(t, filepath.Join(goOut, "doc.go"))
	assertTestFileExists(t, filepath.Join(goOut, "schema.go"))
	assertTestFileStartsWithGeneratedMarker(t, filepath.Join(goOut, "doc.go"))
}

func TestCompileTargetsWithImportedDomainNamedTypeReferences(t *testing.T) {
	baseDir := t.TempDir()
	writeTestFile(t, filepath.Join(baseDir, "domain.skel"), "domain base")
	writeTestFile(t, filepath.Join(baseDir, "types.skel"), `
domain base

pub enum ItemType {
    STANDARD
}

pub data Item {
    type: ItemType
    types: list<ItemType>
    itemsByType: map<ItemType, Item>
}

pub config ItemConfig eternal {
    defaultType: ItemType
}

pub event ItemCreatedEvent {
    payload {
        item: Item
    }
}

pub actor BaseActor {
    via client {}
    auth {
        credential {
            subject: string
        }
        info {
            item: Item
        }
    }
}

pub resource ItemResource {
    check byItem {
        input {
            item: Item
        }
    }
    action read
}

pub service BaseService {
    for BaseActor

    method getItem {
        input {
            type: ItemType
        }
        output Item
    }
}

task SyncItemTask {
    trigger manually {
        input {
            item: Item
        }
    }
}
`)
	appDir := t.TempDir()
	writeTestFile(t, filepath.Join(appDir, "domain.skel"), "domain app")
	writeTestFile(t, filepath.Join(appDir, "types.skel"), `
domain app

import base

pub data AppItem {
    item: base.Item
    kind: base.ItemType
}
`)

	input := skelc.Input{SkelIn: appDir, SkelImports: map[string]string{"base": baseDir}}
	tests := []struct {
		name     string
		file     string
		compile  func(out string) error
		expected []string
	}{
		{
			name: "Go",
			file: "data.go",
			compile: func(out string) error {
				_, err := skelc.CompileGolang(input, skelc.GolangOption{
					CompilerVersion: "v0.0.0-dev",
					Out:             out,
					Imports:         map[string]string{"base": "example.com/basepub"},
				})
				return err
			},
			expected: []string{"Item basepub.Item", "Kind basepub.ItemType"},
		},
		{
			name: "TypeScript",
			file: "data.ts",
			compile: func(out string) error {
				_, err := skelc.CompileTypeScript(input, skelc.TypeScriptOption{ApiOnly: true,
					Out:     out,
					Imports: map[string]string{"base": "@example/base"},
				})
				return err
			},
			expected: []string{"item: baseapi.Item;", "kind: baseapi.ItemType;"},
		},
		{
			name: "Skel",
			file: "types.skel",
			compile: func(out string) error {
				_, err := skelc.CompileSkeleton(input, skelc.SkeletonOption{
					Out:     out,
					PubOnly: true,
				})
				return err
			},
			expected: []string{"item: base.Item", "kind: base.ItemType"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The Go generator derives the package name from the output
			// directory, so it must be a valid identifier.
			out := filepath.Join(t.TempDir(), "generated")
			if err := test.compile(out); err != nil {
				t.Fatalf("compile %s with imported named type references: %v", test.name, err)
			}
			content, err := os.ReadFile(filepath.Join(out, test.file))
			if err != nil {
				t.Fatalf("read generated %s: %v", test.file, err)
			}
			for _, fragment := range test.expected {
				if !strings.Contains(string(content), fragment) {
					t.Fatalf("generated %s missing %q:\n%s", test.file, fragment, content)
				}
			}
		})
	}
}

func TestCompileGolangAnalyzesTransitiveImportsButGeneratesOnlyTarget(t *testing.T) {
	baseDir := t.TempDir()
	writeTestFile(t, filepath.Join(baseDir, "domain.skel"), "domain base")
	writeTestFile(t, filepath.Join(baseDir, "types.skel"), `
domain base

pub enum UserStatus {
    ACTIVE
}
`)
	userDir := t.TempDir()
	writeTestFile(t, filepath.Join(userDir, "domain.skel"), "domain user")
	writeTestFile(t, filepath.Join(userDir, "types.skel"), `
domain user

import base

pub data User {
    status: base.UserStatus
}
`)
	appDir := t.TempDir()
	writeTestFile(t, filepath.Join(appDir, "domain.skel"), "domain app")
	writeTestFile(t, filepath.Join(appDir, "types.skel"), `
domain app

import user

data AppUser {
    user: user.User
}
`)

	goOut := filepath.Join(t.TempDir(), "generated")
	_, err := skelc.CompileGolang(
		skelc.Input{
			SkelIn: appDir,
			SkelImports: map[string]string{
				"base": baseDir,
				"user": userDir,
			},
		},
		skelc.GolangOption{
			CompilerVersion: "v0.0.0-dev",
			Out:             goOut,
			Imports:         map[string]string{"user": "example.com/userpub"},
		},
	)
	if err != nil {
		t.Fatalf("compile Go with transitive imports: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(goOut, "data.go"))
	if err != nil {
		t.Fatalf("read generated target data: %v", err)
	}
	generated := string(content)
	if !strings.Contains(generated, "type AppUser struct") {
		t.Fatalf("target data was not generated:\n%s", generated)
	}
	if strings.Contains(generated, "type User struct") || strings.Contains(generated, "type UserStatus ") {
		t.Fatalf("dependency declarations were generated into target output:\n%s", generated)
	}
}

func TestGenerateTargetsShareParsedDomain(t *testing.T) {
	skelDir := t.TempDir()
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")

	parsed, err := skelc.Parse(skelc.Input{SkelIn: skelDir})
	if err != nil {
		t.Fatalf("parse Skel: %v", err)
	}
	goOut := filepath.Join(t.TempDir(), "golang")
	if err := skelc.GenerateGolang(parsed.Domain, skelc.GolangOption{CompilerVersion: "v0.0.0-dev", Out: goOut}); err != nil {
		t.Fatalf("generate Go: %v", err)
	}
	tsOut := filepath.Join(t.TempDir(), "typescript")
	if err := skelc.GenerateTypeScript(parsed.Domain, skelc.TypeScriptOption{ApiOnly: true, Out: tsOut}); err != nil {
		t.Fatalf("generate TypeScript: %v", err)
	}
	assertTestFileExists(t, filepath.Join(goOut, "schema.go"))
	assertTestFileExists(t, filepath.Join(tsOut, "index.ts"))
	assertTestFileStartsWithGeneratedMarker(t, filepath.Join(tsOut, "index.ts"))
}

func TestGenerateTypeScript(t *testing.T) {
	skelDir := t.TempDir()
	tsOut := filepath.Join(t.TempDir(), "generated")
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")

	_, err := skelc.CompileTypeScript(
		skelc.Input{SkelIn: skelDir},
		skelc.TypeScriptOption{ApiOnly: true, Out: tsOut},
	)
	if err != nil {
		t.Fatalf("generate TypeScript: %v", err)
	}
	assertTestFileExists(t, filepath.Join(tsOut, "index.ts"))
}

func TestGenerateSkeleton(t *testing.T) {
	skelDir := t.TempDir()
	skelOut := filepath.Join(t.TempDir(), "generated")
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")

	_, err := skelc.CompileSkeleton(
		skelc.Input{SkelIn: skelDir},
		skelc.SkeletonOption{Out: skelOut, PubOnly: true},
	)
	if err != nil {
		t.Fatalf("generate Skel: %v", err)
	}
	assertTestFileExists(t, filepath.Join(skelOut, "domain.skel"))
	assertTestFileStartsWithGeneratedMarker(t, filepath.Join(skelOut, "domain.skel"))
	if _, err := skelc.Parse(skelc.Input{SkelIn: skelOut}); err != nil {
		t.Fatalf("parse generated Skel with ownership marker: %v", err)
	}
}

func TestGenerateGolangReturnsErrorBeforeCleaningOutput(t *testing.T) {
	skelDir := t.TempDir()
	goOut := filepath.Join(t.TempDir(), "generated")
	writeTestFile(t, filepath.Join(skelDir, "invalid.skel"), "data User { id: string }")
	oldFile := filepath.Join(goOut, "old.go")
	writeTestFile(t, oldFile, "old")

	_, err := skelc.CompileGolang(
		skelc.Input{SkelIn: skelDir},
		skelc.GolangOption{CompilerVersion: "v0.0.0-dev", Out: goOut},
	)
	if err == nil {
		t.Fatal("expected generation error")
	}
	assertTestFileExists(t, oldFile)
}

func TestGeneratorsReturnErrorsForMalformedProgrammaticModels(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo.invalid",
		Data: []*model.Data{{
			Name: "Broken",
			Pub:  true,
			Members: []*model.DataMember{{
				Name: "value",
				Type: &model.Type{Kind: model.TypeKind(999)},
			}},
		}},
	})
	tests := []struct {
		name     string
		generate func() error
	}{
		{
			name: "Go",
			generate: func() error {
				return skelc.GenerateGolang(domain, skelc.GolangOption{CompilerVersion: "v0.0.0-dev", Out: filepath.Join(t.TempDir(), "generated")})
			},
		},
		{
			name: "TypeScript",
			generate: func() error {
				return skelc.GenerateTypeScript(domain, skelc.TypeScriptOption{ApiOnly: true, Out: t.TempDir()})
			},
		},
		{
			name: "Skel",
			generate: func() error {
				return skelc.GenerateSkeleton(domain, skelc.SkeletonOption{Out: t.TempDir(), PubOnly: true})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.generate()
			if err == nil || !strings.Contains(err.Error(), "unsupported type kind 999") {
				t.Fatalf("expected malformed model error, got %v", err)
			}
		})
	}
}

func TestGeneratorsReturnErrorsForMalformedNestedModels(t *testing.T) {
	domains := []struct {
		name     string
		domain   *model.Domain
		expected string
	}{
		{
			name: "incomplete actor auth",
			domain: model.NewDomainFromSpec(model.DomainSpec{
				Name: "demo.invalid", Actors: []*model.Actor{{Name: "Client", AuthEnabled: true}},
			}),
			expected: "incomplete auth support",
		},
		{
			name: "nil import",
			domain: model.NewDomainFromSpec(model.DomainSpec{
				Name: "demo.invalid", Imports: []*model.Import{nil},
			}),
			expected: "nil import",
		},
	}
	for _, malformed := range domains {
		t.Run(malformed.name, func(t *testing.T) {
			generators := []struct {
				name     string
				generate func() error
			}{
				{name: "Go", generate: func() error {
					return skelc.GenerateGolang(malformed.domain, skelc.GolangOption{CompilerVersion: "v0.0.0-dev", Out: filepath.Join(t.TempDir(), "generated")})
				}},
				{name: "TypeScript", generate: func() error {
					return skelc.GenerateTypeScript(malformed.domain, skelc.TypeScriptOption{ApiOnly: true, Out: t.TempDir()})
				}},
				{name: "Skel", generate: func() error {
					return skelc.GenerateSkeleton(malformed.domain, skelc.SkeletonOption{Out: t.TempDir(), PubOnly: true})
				}},
			}
			for _, generator := range generators {
				t.Run(generator.name, func(t *testing.T) {
					err := generator.generate()
					if err == nil || !strings.Contains(err.Error(), malformed.expected) {
						t.Fatalf("expected error containing %q, got %v", malformed.expected, err)
					}
				})
			}
		})
	}
}

func TestGeneratorsRejectNilTypeParameter(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo.invalid", Data: []*model.Data{{Name: "Box", SkelName: "demo.invalid.Box", Kind: model.DataKindData, Pub: true, TypeParameters: []*model.TypeParameter{nil}}}})
	for _, target := range []string{"Go", "TypeScript", "Skel"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "generated")
			if err := os.Mkdir(out, 0o755); err != nil {
				t.Fatal(err)
			}
			var err error
			switch target {
			case "Go":
				err = skelc.GenerateGolang(domain, skelc.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out})
			case "TypeScript":
				err = skelc.GenerateTypeScript(domain, skelc.TypeScriptOption{ApiOnly: true, Out: out})
			case "Skel":
				err = skelc.GenerateSkeleton(domain, skelc.SkeletonOption{PubOnly: true, Out: out})
			}
			if err == nil || !strings.Contains(err.Error(), "nil type parameter") {
				t.Fatalf("expected invalid model error, got %v", err)
			}
			entries, err := os.ReadDir(out)
			if err != nil || len(entries) != 0 {
				t.Fatalf("invalid model wrote output: %v, %v", entries, err)
			}
		})
	}
}

// b's public contract needs c for parsing, even when a uses only b.Token or its actor.
func TestCompileModulesSeparateResolutionMappingsFromDependencies(t *testing.T) {
	for _, tc := range []struct {
		name, source               string
		full, public, regular, api []string
	}{
		{"external actor", "pub data Payload { id: uuid }\nweb GatewayWeb { for b.AgentActor }", nil, nil, nil, nil},
		{"unused transitive contract", "pub data Payload { token: b.Token }", []string{"b"}, []string{"b"}, nil, []string{"b"}},
		{"split output boundaries", "pub data Payload { access: c.Access }\ndata PrivatePayload { token: b.Token }", []string{"b", "c"}, []string{"c"}, []string{"b"}, []string{"c"}},
		{"generic arguments", "pub data Payload { values: map<string, b.Box<c.Access>> }", []string{"b", "c"}, []string{"b", "c"}, nil, []string{"b", "c"}},
		{"public service implementation", "pub service ExampleService { method get { output b.Token } }", []string{"b"}, []string{"b"}, []string{"b"}, nil},
		{"actor and resource helpers", `pub actor LocalActor {
    via client {}
    auth { credential { token: string } info { access: c.Access } }
}
pub resource LocalResource {
    check byToken { input { token: b.Token } }
    action read
}
api service ExampleApiService { for LocalActor via client method ping {} }`, []string{"b", "c"}, []string{"b", "c"}, nil, nil},
		{"config event and task", `pub config ExampleConfig eternal { token: b.Token }
pub event ExampleEvent { payload { access: c.Access } }
task ExampleTask { trigger manually { input { token: b.Token } } }`, []string{"b", "c"}, []string{"b", "c"}, []string{"b", "c"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			a, b, c := filepath.Join(root, "a.skel"), filepath.Join(root, "b.skel"), filepath.Join(root, "c.skel")
			imports := "import b\n"
			if strings.Contains(tc.source, "c.") {
				imports += "import c\n"
			}
			writeTestFile(t, a, "domain a\n"+imports+tc.source+"\n")
			writeTestFile(t, b, "domain b\nimport c\npub data Token { id: uuid access: c.Access }\npub data Box<TItem> { items: list<TItem> }\npub actor AgentActor { via client {} }\npub service AccessService { method get { output c.Access } }\n")
			writeTestFile(t, c, "domain c\npub data Access { id: uuid }\n")
			input := skelc.Input{SkelIn: a, SkelImports: map[string]string{"b": b, "c": c}}
			mappings := map[string]string{"b": "example.com/bpub@v1.2.3", "c": "example.com/cpub@v1.3.0", "unused": "example.com/unused@v1.4.0"}
			dependencyOutputs := map[bool]map[string]string{}
			for _, api := range []bool{false, true} {
				dependencyOutputs[api] = map[string]string{}
				for name, entry := range map[string]string{"b": b, "c": c} {
					out := filepath.Join(root, map[bool]string{false: "pubdeps", true: "apideps"}[api], name)
					dependencyImports := map[string]string{}
					if name == "b" {
						dependencyImports["c"] = c
					}
					if _, err := skelc.CompileGolang(skelc.Input{SkelIn: entry, SkelImports: dependencyImports}, skelc.GolangOption{
						CompilerVersion: "v0.0.0-dev", PubOnly: !api, ApiOnly: api, AsModule: true, Module: "example.com/" + name + "pub", Out: out, Imports: mappings,
					}); err != nil {
						t.Fatal(err)
					}
					dependencyOutputs[api][name] = out
				}
			}
			for _, mode := range []string{"full", "pub", "api", "split"} {
				t.Run(mode, func(t *testing.T) {
					out := filepath.Join(root, mode)
					opts := skelc.GolangOption{CompilerVersion: "v0.0.0-dev", AsModule: true, Module: "example.com/a", Out: out, Imports: mappings}
					want := tc.full
					switch mode {
					case "pub":
						opts.PubOnly = true
						want = tc.public
					case "api":
						opts.ApiOnly = true
						want = tc.api
					case "split":
						opts.PubOut = out + "pub"
						opts.PubModule = "example.com/apub"
						want = tc.regular
					}
					if _, err := skelc.CompileGolang(input, opts); err != nil {
						t.Fatal(err)
					}
					assertGeneratedDomainRequires(t, out, want)
					if mode == "split" {
						assertGeneratedDomainRequires(t, opts.PubOut, tc.public)
						content, err := os.ReadFile(filepath.Join(out, "go.mod"))
						if err != nil {
							t.Fatal(err)
						}
						if !strings.Contains(string(content), "example.com/apub ") {
							t.Fatal("split output lost its own public-module dependency")
						}
					}
					t.Run("compile", func(t *testing.T) {
						testutil.RequireToolchain(t)
						// Use generated local dependencies; go.work must not hide a
						// missing requirement in any generated module.
						for name, dir := range dependencyOutputs[mode == "api"] {
							testutil.Go(t, out, "mod", "edit", "-replace=example.com/"+name+"pub="+dir)
						}
						if mode == "split" {
							testutil.Go(t, out, "mod", "edit", "-replace=example.com/apub="+opts.PubOut)
						}
						testutil.Go(t, out, "mod", "tidy")
						files, err := filepath.Glob(filepath.Join(out, "*.go"))
						if err != nil {
							t.Fatal(err)
						}
						if len(files) > 0 {
							testutil.Go(t, out, "test", "./...")
						}
					})
				})
			}
			tsOut := filepath.Join(root, "ts")
			if _, err := skelc.CompileTypeScript(input, skelc.TypeScriptOption{ApiOnly: true, AsModule: true, Module: "@example/a", Out: tsOut, Imports: map[string]string{"b": "@example/b@1.2.3", "c": "@example/c@1.3.0", "unused": "@example/unused@1.4.0"}}); err != nil {
				t.Fatal(err)
			}
			manifest, err := os.ReadFile(filepath.Join(tsOut, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"b", "c", "unused"} {
				want := false
				for _, dep := range tc.api {
					want = want || dep == name
				}
				if strings.Contains(string(manifest), `"@example/`+name+`"`) != want {
					t.Fatalf("wrong TypeScript dependency %s: %s", name, manifest)
				}
			}
			if mappings["unused"] != "example.com/unused@v1.4.0" || len(mappings) != 3 {
				t.Fatal("generation mutated caller's import mappings")
			}
		})
	}
}

func assertGeneratedDomainRequires(t *testing.T, out string, want []string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	file, err := modfile.Parse("go.mod", content, nil)
	if err != nil {
		t.Fatal(err)
	}
	requires := map[string]string{}
	for _, require := range file.Require {
		requires[require.Mod.Path] = require.Mod.Version
	}
	// Cross-check model-based dependencies against independently parsed output.
	paths := map[string]bool{}
	files, err := filepath.Glob(filepath.Join(out, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		parsed, err := stdparser.ParseFile(token.NewFileSet(), path, nil, stdparser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			paths[path] = true
		}
	}
	for _, path := range []string{"example.com/bpub", "example.com/cpub", "example.com/unused"} {
		if paths[path] != (requires[path] != "") {
			t.Fatalf("dependency %s disagrees with generated imports: %s", path, content)
		}
	}
	for name, version := range map[string]string{"b": "v1.2.3", "c": "v1.3.0", "unused": "v1.4.0"} {
		used := false
		for _, dep := range want {
			used = used || dep == name
		}
		got := requires["example.com/"+name+"pub"]
		if name == "unused" {
			got = requires["example.com/unused"]
		}
		if used && got != version || !used && got != "" {
			t.Fatalf("dependency %s = %q, used=%v: %s", name, got, used, content)
		}
	}
}

func TestCompileGolangIncludesUsedPrefixDerivedDependencies(t *testing.T) {
	for _, api := range []bool{false, true} {
		t.Run(map[bool]string{false: "backend", true: "api"}[api], func(t *testing.T) {
			root := t.TempDir()
			a, b := filepath.Join(root, "a.skel"), filepath.Join(root, "b.skel")
			writeTestFile(t, a, "domain a\nimport b\npub data Payload { token: b.Token }\n")
			writeTestFile(t, b, "domain b\npub data Token { id: string }\n")
			out := filepath.Join(root, "out")
			if _, err := skelc.CompileGolang(skelc.Input{SkelIn: a, SkelImports: map[string]string{"b": b}}, skelc.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: api, AsModule: true, ModulePrefix: "example.com/gen", Out: out}); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join(out, "go.mod"))
			if err != nil {
				t.Fatal(err)
			}
			suffix := map[bool]string{false: "pub", true: "api"}[api]
			if !strings.Contains(string(content), "example.com/gen/b"+suffix+" v0.0.0-00010101000000-000000000000") {
				t.Fatalf("missing inferred dependency: %s", content)
			}
		})
	}
}

func TestCompileGolangDependencyConflictCompatibility(t *testing.T) {
	for _, api := range []bool{false, true} {
		for _, used := range []int{0, 1, 2} {
			t.Run(fmt.Sprintf("api=%v/used=%d", api, used), func(t *testing.T) {
				root := t.TempDir()
				a := filepath.Join(root, "a.skel")
				first, second := filepath.Join(root, "first.skel"), filepath.Join(root, "second.skel")
				writeTestFile(t, first, "domain first\npub data Value { id: string }\n")
				writeTestFile(t, second, "domain second\npub data Value { id: string }\n")
				source := "domain a\nimport first\nimport second\npub data Payload { id: string\n"
				if used > 0 {
					source += "first: first.Value\n"
				}
				if used > 1 {
					source += "second: second.Value\n"
				}
				writeTestFile(t, a, source+"}\n")
				out := filepath.Join(root, "out")
				sentinel := filepath.Join(out, "sentinel.go")
				original := "// Code generated by skelc. DO NOT EDIT.\npackage sentinel\n"
				writeTestFile(t, sentinel, original)
				_, err := skelc.CompileGolang(skelc.Input{SkelIn: a, SkelImports: map[string]string{"first": first, "second": second}}, skelc.GolangOption{
					CompilerVersion: "v0.0.0-dev", ApiOnly: api, AsModule: true, Module: "example.com/a", Out: out,
					Imports: map[string]string{"first": "example.com/shared@v1.0.0", "second": "example.com/shared@v1.1.0"},
				})
				if !api || used == 2 {
					if err == nil || !strings.Contains(err.Error(), "conflicting") {
						t.Fatalf("expected dependency conflict, got %v", err)
					}
					got, readErr := os.ReadFile(sentinel)
					if readErr != nil || string(got) != original {
						t.Fatalf("conflict changed existing output: %s, %v", got, readErr)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				content, err := os.ReadFile(filepath.Join(out, "go.mod"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(content), "example.com/shared v1.0.0") != (used == 1) || strings.Contains(string(content), "v1.1.0") {
					t.Fatalf("unused mapping affected API dependencies: %s", content)
				}
			})
		}
	}
}
