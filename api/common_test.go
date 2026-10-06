package api_test

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

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/schema"
)

func TestParseFrozenImportGraph(t *testing.T) {
	root := t.TempDir()
	entry, dep := filepath.Join(root, "entry.skel"), filepath.Join(root, "dep.skel")
	// Disk has a different domain; the frozen graph must take precedence.
	writeTestFile(t, entry, "domain wrong\n")
	input := api.Input{SkelIn: entry, SkelImports: map[string]string{"shared": dep}, Sources: map[string][]byte{
		entry: []byte("domain demo\nimport shared\npub data Value { value: shared.Value }\n"),
		dep:   []byte("domain shared\npub data Value { value: string }\n"),
	}}
	parsed, err := api.ParseContext(t.Context(), input)
	if err != nil || parsed.Domain.Name() != "demo" {
		t.Fatalf("parse=%+v, err=%v", parsed, err)
	}
	output := filepath.Join(root, "generated")
	if _, err := api.CompileSkeleton(input, api.SkeletonOption{Out: output, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	delete(input.Sources, dep)
	if _, err := api.Parse(input); err == nil {
		t.Fatal("missing snapshot dependency accepted")
	}
	input.Sources = map[string][]byte{}
	if _, err := api.Parse(input); err == nil {
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
		_, err := api.ScanImports(api.ScanOption{SkelIn: path, Sources: files})
		if err == nil || !(strings.Contains(err.Error(), "path") || strings.Contains(err.Error(), "source")) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestPublicReadAPIsHonorCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	input := api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte("domain demo\n")}}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	calls := []func() error{
		func() error { _, err := api.ParseContext(ctx, input); return err },
		func() error {
			_, err := api.CheckContext(ctx, api.CheckOption{SkelIn: path, Sources: input.Sources})
			return err
		},
		func() error {
			_, err := api.ScanImportsContext(ctx, api.ScanOption{SkelIn: path, Sources: input.Sources})
			return err
		},
		func() error {
			_, err := api.FormatFilesContext(ctx, api.FormatOption{SkelIn: path, Sources: input.Sources})
			return err
		},
		func() error {
			_, err := api.QuerySchemaContext(ctx, input, api.SchemaQueryOption{})
			return err
		},
		func() error {
			_, err := api.QueryApiDependenciesContext(ctx, input, api.ApiFilter{})
			return err
		},
		func() error {
			_, err := api.QuerySchemaDependenciesContext(ctx, input, api.SchemaDependencyOption{})
			return err
		},
		func() error {
			_, err := api.DiffSchemaSourcesContext(ctx, input, api.SchemaDiffOption{Baseline: &input})
			return err
		},
	}
	for i, call := range calls {
		if err := call(); !errors.Is(err, context.Canceled) {
			t.Fatalf("call %d lost cancellation: %v", i, err)
		}
	}
}

func TestParseEnforcesLanguageRulesInAllModes(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "order.skel")
	writeTestFile(t, entry, "domain demo.order\nservice OrderService { method ping {} }\n")
	for _, strict := range []bool{false, true} {
		_, err := api.Parse(api.Input{SkelIn: entry, Strict: strict})
		var diagnostics diagnostic.Diagnostics
		if !errors.As(err, &diagnostics) || len(diagnostics) != 1 || diagnostics[0].Code != diagnostic.CodeServiceModifier || diagnostics[0].Severity != diagnostic.SeverityError {
			t.Fatalf("unexpected diagnostics: %v", err)
		}
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

	result, err := api.Parse(api.Input{SkelIn: skelDir})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Domain.Name())

	// Output: demo.user
}

func TestParseExposesSemanticSchema(t *testing.T) {
	skelDir := t.TempDir()
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")

	result, err := api.Parse(api.Input{SkelIn: skelDir})
	if err != nil {
		t.Fatalf("parse Skel: %v", err)
	}
	var domain *schema.Domain = result.Domain
	if domain.Name() != "demo.user" {
		t.Fatalf("unexpected domain: %s", domain.Name())
	}
}

func TestParseReturnsStructuredWarnings(t *testing.T) {
	skelDir := t.TempDir()
	writeTestFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")
	writeTestFile(t, filepath.Join(skelDir, ".ignored.skel"), "domain ignored")

	result, err := api.Parse(api.Input{SkelIn: skelDir})
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
	const internalPrefix = "go.yorun.ai/skel/internal/"
	rules := []struct {
		directory   string
		forbidden   func(string) bool
		packageOnly bool
	}{
		{
			directory: "internal/parser",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) &&
					path != internalPrefix+"parser/grammar"
			},
		},
		{
			directory: "internal/analyzer",
			forbidden: func(path string) bool {
				for _, prefix := range []string{"compiler", "hasher", "loader", "lsp", "codegen", "projection", "sourcediff"} {
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
				return strings.HasPrefix(path, internalPrefix)
			},
		},
		{
			directory: "diagnostic",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"location"
			},
		},
		{
			directory: "cmd/skelc/output",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"cmd/skelc/output"
			},
		},
		{
			directory: "internal/cmd/skelc/output",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) || strings.HasPrefix(path, "go.yorun.ai/skel/api")
			},
		},
		{
			directory: "api",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"api"
			},
		},
		{
			directory: "codegen",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, internalPrefix) && path != internalPrefix+"codegen"
			},
		},

		{
			directory:   "schema",
			packageOnly: true,
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, "go.yorun.ai/skel/") && path != internalPrefix+"location" && path != internalPrefix+"policy" && !strings.HasPrefix(path, internalPrefix+"util/")
			},
		},
		{
			directory: "descriptor",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, "go.yorun.ai/") && path != internalPrefix+"policy"
			},
		},
		{
			// Shared policy rules must stay independent of both representations,
			// the compiler, and runtime frameworks.
			directory: "internal/policy",
			forbidden: func(path string) bool {
				return strings.Contains(strings.Split(path, "/")[0], ".")
			},
		},
		{
			directory: "schema/diff",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, "go.yorun.ai/skel/") && path != "go.yorun.ai/skel/schema"
			},
		},
		{
			directory: "types",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, "go.yorun.ai/skel/") || strings.HasPrefix(path, "go.yorun.ai/vine/") || strings.HasPrefix(path, "go.yorun.ai/vrpc/")
			},
		},
		{
			directory: "internal/api",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, "go.yorun.ai/skel/api") || strings.HasPrefix(path, internalPrefix+"cmd/") || strings.HasPrefix(path, internalPrefix+"lsp") || path == internalPrefix+"codegen/output"
			},
		},
		{
			directory: "internal/compiler",
			forbidden: func(path string) bool {
				return strings.HasPrefix(path, "go.yorun.ai/skel/api") || path == internalPrefix+"api" || strings.HasPrefix(path, internalPrefix+"cmd/") || strings.HasPrefix(path, internalPrefix+"lsp") || targetCodegenImport(path)
			},
		},
		{
			directory:   "internal/codegen",
			packageOnly: true,
			forbidden: func(path string) bool {
				return path == internalPrefix+"codegen/binding" || targetCodegenImport(path)
			},
		},
		{
			directory:   "internal/codegen/binding",
			packageOnly: true,
			forbidden:   targetCodegenImport,
		},
		{
			directory: "internal/codegen/output",
			forbidden: func(path string) bool {
				return path == internalPrefix+"codegen" || path == internalPrefix+"codegen/binding" || targetCodegenImport(path)
			},
		},
	}

	// Shared capabilities must not depend on API adapters or command implementations.
	for _, directory := range []string{"internal/compiler", "internal/parser", "internal/symbol", "internal/analyzer", "schema", "schema/diff", "descriptor", "types", "internal/formatter", "internal/codegen", "internal/lsp", "internal/sourcediff", "internal/loader", "internal/source", "internal/hasher", "internal/optionvalidation"} {
		t.Run(directory+"/no-api-or-command-dependencies", func(t *testing.T) {
			inspectProductionImports(t, directory, func(path string) bool {
				return path == internalPrefix+"api" || path == "go.yorun.ai/skel/api" || strings.HasPrefix(path, internalPrefix+"cmd/") || strings.HasPrefix(path, "go.yorun.ai/skel/cmd/")
			})
		})
	}
	for _, rule := range rules {
		t.Run(filepath.ToSlash(rule.directory), func(t *testing.T) {
			inspectProductionImports(t, rule.directory, rule.forbidden, rule.packageOnly)
		})
	}
}

func targetCodegenImport(path string) bool {
	return strings.HasPrefix(path, "go.yorun.ai/skel/internal/codegen/binding/")
}

func inspectProductionImports(t *testing.T, directory string, forbidden func(string) bool, packageOnly ...bool) {
	t.Helper()
	err := filepath.WalkDir(filepath.Join("..", directory), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// Independently owned subpackages have their own dependency rules.
		if entry.IsDir() && (!strings.HasPrefix(directory, "internal/") || directory == "internal/compiler" || (len(packageOnly) > 0 && packageOnly[0])) && path != filepath.Join("..", directory) {
			return filepath.SkipDir
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
