package compiler

import (
	"errors"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/diagnostic"
)

func TestStrictCompilationChecksImportedLanguageRules(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "shared.skel")
	entry := filepath.Join(dir, "order.skel")
	writeFile(t, dependency, "domain demo.shared\npub data Item { id: string }\npub service BackendService { auth optional method ping {} }\n")
	writeFile(t, entry, "domain demo.order\nimport demo.shared as shared\nactor TestActor { via client {} }\napi service OrderApiService { for TestActor via client auth required method get { output shared.Item } }\n")
	option := Option{SkelIn: entry, SkelImports: map[string]string{"demo.shared": dependency}}
	for _, strict := range []bool{false, true} {
		option.Strict = strict
		_, err := Compile(option)
		var diagnostics Diagnostics
		if !errors.As(err, &diagnostics) {
			t.Fatalf("expected imported validation error, got %v", err)
		}
		found := false
		for _, item := range diagnostics {
			if item.Code == diagnostic.CodeServiceClientRules && item.Severity == DiagnosticSeverityError && item.Range.Start.File == dependency {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing imported diagnostic: %+v", diagnostics)
		}
	}
}

func TestStrictCompilationPreservesLoaderWarnings(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "domain.skel"), "domain demo.order\n")
	writeFile(t, filepath.Join(dir, "service.skel"), "domain demo.order\nactor TestActor { via client {} }\napi service OrderApiService { for TestActor via client auth required method ping {} }\npub service BackendService { method ping {} }\n")
	writeFile(t, filepath.Join(dir, ".hidden.skel"), "ignored")
	for _, compile := range []func(Option) (Result, error){Compile, CompileImport} {
		result, err := compile(Option{SkelIn: dir, Strict: true})
		if err != nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != diagnostic.CodeLoaderHiddenFile || result.Diagnostics[0].Severity != DiagnosticSeverityWarning {
			t.Fatalf("unexpected strict compilation: %+v, %v", result, err)
		}
	}
}
