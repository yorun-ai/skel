package api_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/diagnostic"
)

func TestScanImportsPreservesSourceDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "domain.skel"), "domain demo\n")
	first, second := filepath.Join(dir, "a.skel"), filepath.Join(dir, "b.skel")
	writeTestFile(t, first, "domain demo\n// import ignored.comment\nimport absent.domain\npub data Value { value: absent.domain.Value }\n")
	writeTestFile(t, second, "domain demo\n/* import ignored.block */\nimport absent.domain as alias\n@desc(\"\"\"\nimport ignored.description\n\"\"\")\npub data Other { value: alias.Value }\n")
	result, err := api.ScanImports(api.ScanOption{SkelIn: dir, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	imports := result.Imports
	want := []api.ScanImport{
		{Domain: "absent.domain", File: first, Line: 3, Column: 1},
		{Domain: "absent.domain", Alias: "alias", File: second, Line: 3, Column: 1},
	}
	if !reflect.DeepEqual(imports, want) {
		t.Fatalf("imports=%+v, want %+v", imports, want)
	}
	again, err := api.ScanImportsContext(t.Context(), api.ScanOption{SkelIn: dir, Strict: true})
	if err != nil || !reflect.DeepEqual(result, again) {
		t.Fatalf("unstable scan: %+v, %v", again, err)
	}
}

func TestScanImportsRejectsInvalidDeclarationsInAllModes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	writeTestFile(t, path, "domain demo\nservice LegacyService { method ping {} }\n")
	for _, strict := range []bool{false, true} {
		_, err := api.ScanImports(api.ScanOption{SkelIn: path, Strict: strict})
		var diagnostics diagnostic.Diagnostics
		if !errors.As(err, &diagnostics) || len(diagnostics) != 1 || diagnostics[0].Severity != diagnostic.SeverityError || diagnostics[0].Position.File != path {
			t.Fatalf("unexpected diagnostics: %v", err)
		}
	}
}

func TestScanImportsErrorsAndCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	writeTestFile(t, path, "domain demo\nimport\n")
	for _, input := range []string{"", path, filepath.Join(t.TempDir(), "missing.skel")} {
		if _, err := api.ScanImports(api.ScanOption{SkelIn: input}); err == nil {
			t.Fatalf("accepted invalid input %q", input)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := api.ScanImportsContext(ctx, api.ScanOption{SkelIn: path}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
