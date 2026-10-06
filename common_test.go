package skelc_test

import (
	"context"
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

func TestParseFrozenImportGraph(t *testing.T) {
	root := t.TempDir()
	entry, dep := filepath.Join(root, "entry.skel"), filepath.Join(root, "dep.skel")
	// Disk has a different domain; the frozen graph must take precedence.
	writeTestFile(t, entry, "domain wrong\n")
	input := skelc.Input{SkelIn: entry, SkelImports: map[string]string{"shared": dep}, Sources: map[string][]byte{
		entry: []byte("domain demo\nimport shared\npub data Value { value: shared.Value }\n"),
		dep:   []byte("domain shared\npub data Value { value: string }\n"),
	}}
	parsed, err := skelc.ParseContext(t.Context(), input)
	if err != nil || parsed.Domain.Name() != "demo" {
		t.Fatalf("parse=%+v, err=%v", parsed, err)
	}
	output := filepath.Join(root, "generated")
	if _, err := skelc.CompileSkeleton(input, skelc.SkeletonOption{Out: output, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	delete(input.Sources, dep)
	if _, err := skelc.Parse(input); err == nil {
		t.Fatal("missing snapshot dependency accepted")
	}
	input.Sources = map[string][]byte{}
	if _, err := skelc.Parse(input); err == nil {
		t.Fatal("empty snapshot fell back to disk")
	}
}

func TestFrozenInputRejectsAmbiguousPaths(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.skel")
	source := []byte("domain demo\n")
	for _, files := range []map[string][]byte{
		{"": source},
		{path: source, filepath.Join(path, "child.skel"): source},
		{path: source, filepath.Join(path, "child.skel", "nested.skel"): source},
		{path: source, root + "/./source.skel": source},
	} {
		_, err := skelc.ScanImports(skelc.ScanOption{SkelIn: path, Sources: files})
		if err == nil || !(strings.Contains(err.Error(), "path") || strings.Contains(err.Error(), "source")) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestPublicReadAPIsHonorCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	input := skelc.Input{SkelIn: path, Sources: map[string][]byte{path: []byte("domain demo\n")}}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	calls := []func() error{
		func() error { _, err := skelc.ParseContext(ctx, input); return err },
		func() error {
			_, err := skelc.CheckContext(ctx, skelc.CheckOption{SkelIn: path, Sources: input.Sources})
			return err
		},
		func() error {
			_, err := skelc.ScanImportsContext(ctx, skelc.ScanOption{SkelIn: path, Sources: input.Sources})
			return err
		},
		func() error {
			_, err := skelc.FormatFilesContext(ctx, skelc.FormatOption{SkelIn: path, Sources: input.Sources})
			return err
		},
		func() error { _, err := skelc.QuerySchemaContext(ctx, input, skelc.SchemaQueryOption{}); return err },
		func() error { _, err := skelc.QueryApiDependenciesContext(ctx, input, skelc.ApiFilter{}); return err },
		func() error {
			_, err := skelc.QuerySchemaDependenciesContext(ctx, input, skelc.SchemaDependencyOption{})
			return err
		},
		func() error {
			_, err := skelc.DiffSchemaSourcesContext(ctx, input, skelc.SchemaDiffOption{Baseline: &input})
			return err
		},
	}
	for i, call := range calls {
		if err := call(); !errors.Is(err, context.Canceled) {
			t.Fatalf("call %d lost cancellation: %v", i, err)
		}
	}
}

func TestParseStrictMigrationRules(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "order.skel")
	writeTestFile(t, entry, "domain demo.order\nservice OrderService { method ping {} }\n")
	result, err := skelc.Parse(skelc.Input{SkelIn: entry})
	if err != nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].Severity != diagnostic.SeverityWarning {
		t.Fatalf("unexpected compatible parse: %+v, %v", result, err)
	}
	_, err = skelc.Parse(skelc.Input{SkelIn: entry, Strict: true})
	var diagnostics diagnostic.Diagnostics
	if !errors.As(err, &diagnostics) || len(diagnostics) != 1 || diagnostics[0].Code != diagnostic.CodeServiceModifier || diagnostics[0].Severity != diagnostic.SeverityError {
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
	entry := result.Diagnostics[0]
	if entry.Code != "loader.ignored-hidden-file" || entry.Severity != diagnostic.SeverityWarning {
		t.Fatalf("unexpected warning: %+v", entry)
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
