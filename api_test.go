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
	"go.yorun.ai/skelc/model"
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
	assertTestFileMissing(t, filepath.Join(goOut, ".skelc-manifest.json"))
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
	assertTestFileMissing(t, filepath.Join(tsOut, ".skelc-manifest.json"))
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
	assertTestFileMissing(t, filepath.Join(skelOut, ".skelc-manifest.json"))
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

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create test directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func assertTestFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
}

func assertTestFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file %s to be missing: %v", path, err)
	}
}

func assertTestFileStartsWithGeneratedMarker(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file %s: %v", path, err)
	}
	const marker = "// Code generated by skelc. DO NOT EDIT.\n"
	if !strings.HasPrefix(string(content), marker) {
		t.Fatalf("expected %s to start with %q, got %q", path, marker, content)
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
