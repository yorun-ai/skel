package skelc_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"go.yorun.ai/skelc"
	"go.yorun.ai/skelc/cli"
	"go.yorun.ai/skelc/diagnostic"
)

func TestScanImportsPreservesSourceDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "domain.skel"), "domain demo\n")
	first, second := filepath.Join(dir, "a.skel"), filepath.Join(dir, "b.skel")
	writeTestFile(t, first, "domain demo\n// import ignored.comment\nimport absent.domain\npub data Value { value: absent.domain.Value }\n")
	writeTestFile(t, second, "domain demo\n/* import ignored.block */\nimport absent.domain as alias\n@desc(\"\"\"\nimport ignored.description\n\"\"\")\npub data Other { value: alias.Value }\n")
	result, err := skelc.ScanImports(skelc.ScanOption{SkelIn: dir, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	// The API and CLI share their existing JSON result type.
	var imports []cli.ScanImport = result.Imports
	want := []skelc.ScanImport{
		{Domain: "absent.domain", File: first, Line: 3, Column: 1},
		{Domain: "absent.domain", Alias: "alias", File: second, Line: 3, Column: 1},
	}
	if !reflect.DeepEqual(imports, want) {
		t.Fatalf("imports=%+v, want %+v", imports, want)
	}
	again, err := skelc.ScanImportsContext(t.Context(), skelc.ScanOption{SkelIn: dir, Strict: true})
	if err != nil || !reflect.DeepEqual(result, again) {
		t.Fatalf("unstable scan: %+v, %v", again, err)
	}
}

func TestScanImportsWarningsAndStrictMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	writeTestFile(t, path, "domain demo\nservice LegacyService { method ping {} }\n")
	result, err := skelc.ScanImports(skelc.ScanOption{SkelIn: path})
	if err != nil || result.Imports == nil || len(result.Imports) != 0 || len(result.Diagnostics) == 0 {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
	for _, entry := range result.Diagnostics {
		if entry.Severity != diagnostic.SeverityWarning || entry.Position.File != path {
			t.Fatalf("unexpected diagnostic: %+v", entry)
		}
	}
	if _, err := skelc.ScanImports(skelc.ScanOption{SkelIn: path, Strict: true}); err == nil {
		t.Fatal("strict mode accepted legacy syntax")
	}
}

func TestScanImportsErrorsAndCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	writeTestFile(t, path, "domain demo\nimport\n")
	for _, input := range []string{"", path, filepath.Join(t.TempDir(), "missing.skel")} {
		if _, err := skelc.ScanImports(skelc.ScanOption{SkelIn: input}); err == nil {
			t.Fatalf("accepted invalid input %q", input)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := skelc.ScanImportsContext(ctx, skelc.ScanOption{SkelIn: path}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
