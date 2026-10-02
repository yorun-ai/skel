package cli

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

func TestRunSkelcSchemaHelpShowsSubcommandsAndOptions(t *testing.T) {
	result := Run([]string{"schema", "--help"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	for _, expected := range []string{"list [TYPE]", "get TYPE SKEL_NAME", "list OPTIONS:", "get OPTIONS:", "snapshot OPTIONS:", "diff OPTIONS:", "--baseline-skel-in"} {
		if !strings.Contains(result.Stdout, expected) {
			t.Fatalf("expected %q in stdout: %q", expected, result.Stdout)
		}
	}
}

func TestRunSkelcGenHelpShowsSubcommandOptions(t *testing.T) {
	result := Run([]string{"gen", "--help"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	genSkelIndex := strings.Index(result.Stdout, "skel OPTIONS:")
	genGoIndex := strings.Index(result.Stdout, "go OPTIONS:")
	genGoModuleIndex := strings.Index(result.Stdout, "go-module OPTIONS:")
	genTSIndex := strings.Index(result.Stdout, "ts OPTIONS:")
	if genSkelIndex < 0 || genGoIndex < 0 || genGoModuleIndex < 0 || genTSIndex < 0 {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if genSkelIndex > genGoIndex || genSkelIndex > genTSIndex {
		t.Fatalf("expected skel options first, got stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--skel-in") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--skel-out") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--go-out") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--ts-out") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--ts-as-module") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--ts-module-scope") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--ts-module") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--ts-import") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--go-module-prefix") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--go-module") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--skel-import") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--go-import") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "--pub") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
}

func TestRunSkelcGenGoHelpShowsLimitedOptions(t *testing.T) {
	result := Run([]string{"gen", "go", "--help"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	for _, expected := range []string{"--skel-in", "--go-out", "--go-vine-version"} {
		if !strings.Contains(result.Stdout, expected) {
			t.Fatalf("expected %s in stdout: %q", expected, result.Stdout)
		}
	}
	for _, unexpected := range []string{"--go-module-prefix", "--go-module"} {
		if strings.Contains(result.Stdout, unexpected) {
			t.Fatalf("did not expect %s in stdout: %q", unexpected, result.Stdout)
		}
	}
}

func TestRunSkelcGenGoModuleHelpShowsModes(t *testing.T) {
	result := Run([]string{"gen", "go-module", "--help"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	assertHelpOptionsOrder(t, result.Stdout, []string{
		"--skel-in",
		"--skel-import",
		"--go-out",
		"--go-module",
		"--go-pub-out",
		"--go-pub-module",
		"--go-import",
		"--go-module-prefix",
		"--go-vine-version",
	})
	if !strings.Contains(result.Stdout, "--pub") || !strings.Contains(result.Stdout, "--api") {
		t.Fatalf("expected --pub and --api in stdout: %q", result.Stdout)
	}
}

func assertHelpOptionsOrder(t *testing.T, stdout string, options []string) {
	t.Helper()
	lastIndex := -1
	for _, option := range options {
		index := strings.Index(stdout, option)
		if index < 0 {
			t.Fatalf("expected %s in stdout: %q", option, stdout)
		}
		if index < lastIndex {
			t.Fatalf("expected %s after previous options in stdout: %q", option, stdout)
		}
		lastIndex = index
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
