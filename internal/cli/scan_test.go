package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/command"
)

func TestScanImportsListsDirectDeclarationsWithoutLoadingDependencies(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), "domain demo.main\n")
	first := filepath.Join(dir, "a.skel")
	second := filepath.Join(dir, "b.skel")
	writeCLIFile(t, first, "domain demo.main\n// import ignored.comment\nimport external.shared\n\npub data Payload { value: external.shared.Value }\n")
	writeCLIFile(t, second, "domain demo.main\n/* import ignored.block */\nimport external.shared as shared\n@desc(\"\"\"\nimport ignored.description\n\"\"\")\npub data Other { value: shared.Value }\n")
	// The imported domain is not present on disk and needs no mapping.
	args := []string{"scan", "imports", "--skel-in", dir}
	result := Run(args)
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("import query failed: %+v", result)
	}
	var imports []command.ScanImport
	if err := json.Unmarshal([]byte(result.Stdout), &imports); err != nil {
		t.Fatal(err)
	}
	if len(imports) != 2 {
		t.Fatalf("expected both source declarations, got %+v", imports)
	}
	for i, want := range []command.ScanImport{
		{Domain: "external.shared", File: first, Line: 3, Column: 1},
		{Domain: "external.shared", Alias: "shared", File: second, Line: 3, Column: 1},
	} {
		if imports[i] != want {
			t.Fatalf("import %d = %+v, want %+v", i, imports[i], want)
		}
	}
	if again := Run(args); again.ExitCode != ExitCodeSuccess || again.Stdout != result.Stdout {
		t.Fatalf("unstable import query: %+v", again)
	}
}

func TestScanImportsEmptyAndInvalidInputs(t *testing.T) {
	root := t.TempDir()
	empty, broken := filepath.Join(root, "empty.skel"), filepath.Join(root, "broken.skel")
	writeCLIFile(t, empty, "domain demo.empty\npub data Value { id: string }\n")
	writeCLIFile(t, broken, "domain demo.broken\nimport\n")
	result := Run([]string{"scan", "imports", "--skel-in", empty})
	if result.ExitCode != ExitCodeSuccess || strings.TrimSpace(result.Stdout) != "[]" {
		t.Fatalf("expected empty array, got %+v", result)
	}
	for _, tc := range []struct {
		args []string
		code command.ErrorCode
	}{
		{[]string{"scan", "imports"}, command.ErrorCodeInvalidArgument},
		{[]string{"scan", "imports", "extra", "--skel-in", empty}, command.ErrorCodeInvalidArgument},
		{[]string{"scan", "imports", "--skel-in", broken}, command.ErrorCodeCompilationFailed},
		{[]string{"scan", "imports", "--skel-in", filepath.Join(root, "missing.skel")}, command.ErrorCodeCompilationFailed},
	} {
		result := Run(tc.args)
		var failure command.Error
		if err := json.Unmarshal([]byte(result.Stdout), &failure); err != nil {
			t.Fatal(err)
		}
		if result.ExitCode != ExitCodeError || failure.Code != tc.code {
			t.Fatalf("args %v: got %+v, want %s", tc.args, result, tc.code)
		}
	}
}

func TestScanImportsStrictAndCommandRemoval(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.skel")
	writeCLIFile(t, source, "domain demo\nservice LegacyService { method ping {} }\n")
	result := Run([]string{"scan", "imports", "--skel-in", source})
	if result.ExitCode != ExitCodeSuccess || result.Stdout != "[]\n" || !strings.Contains(result.Stderr, `"severity":"warning"`) {
		t.Fatalf("unexpected scan result: %+v", result)
	}
	result = Run([]string{"--strict", "scan", "imports", "--skel-in", source})
	failure := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || failure.Code != command.ErrorCodeCompilationFailed {
		t.Fatalf("strict ignored: %+v", result)
	}
	result = Run([]string{"schema", "import", "--skel-in", source})
	if result.ExitCode == ExitCodeSuccess {
		t.Fatalf("removed schema import accepted: %+v", result)
	}
}

func TestScanHelp(t *testing.T) {
	result := Run([]string{"scan", "--help"})
	if result.ExitCode != ExitCodeSuccess || !strings.Contains(result.Stdout, "imports OPTIONS:") || !strings.Contains(result.Stdout, "--skel-in") {
		t.Fatalf("unexpected help: %+v", result)
	}
	result = Run([]string{"schema", "--help"})
	if strings.Contains(result.Stdout, "import OPTIONS:") {
		t.Fatal("schema help still exposes import")
	}
	result = Run([]string{"--help"})
	if !strings.Contains(result.Stdout, "scan") {
		t.Fatal("root help does not expose scan")
	}
}
