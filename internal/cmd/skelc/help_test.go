package skelc

import (
	"strings"
	"testing"
)

func TestRunSkelcHelpShowsSubcommandOptions(t *testing.T) {
	result := Run([]string{"--strict", "--help"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "gen") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "schema") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "check") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "format") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "lsp") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--strict") {
		t.Fatalf("missing strict option: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "check OPTIONS:") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "format OPTIONS:") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
}

func TestRunSkelcWithoutArgsShowsHelp(t *testing.T) {
	result := Run(nil)

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	if result.Stderr != "" {
		t.Fatalf("unexpected stderr: %q", result.Stderr)
	}
	if !strings.Contains(result.Stdout, "USAGE:") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "gen") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if strings.Contains(result.Stdout, "gen-go") || strings.Contains(result.Stdout, "gen-ts") || strings.Contains(result.Stdout, "gen-skel") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
}
