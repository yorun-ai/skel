package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc"
)

func TestSchemaDepAndPruneFlags(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.skel")
	writeCLIFile(t, source, `domain demo
actor UserActor { via client {} }
pub data Unused { value: string }
pub enum Status { READY DONE }
data Value { status: Status }
api service ReadApiService { for UserActor via client auth anonymous method read { output Value } }
`)
	base := []string{"schema", "dep", "--api", "--skel-in", source}
	// Repeated names select a union without selecting any services.
	result := Run(append(append([]string{}, base...), "--prune", "--name", "demo.Status", "--name", "demo.Unused"))
	var report skelc.ApiDependencyReport
	if result.ExitCode != 0 || json.Unmarshal([]byte(result.Stdout), &report) != nil {
		t.Fatalf("%+v", result)
	}
	if len(report.Services) != 0 || len(report.Data) != 1 || report.Data[0] != "demo.Unused" || len(report.Enums) != 1 || report.Enums[0] != "demo.Status" {
		t.Fatalf("%+v", report)
	}
	assertCommandErrorMessage(t, Run(append(append([]string{}, base...), "--name", "demo.Status")), "flag name requires prune")
	assertCommandErrorMessage(t, Run(append(append([]string{}, base...), "--prune", "--type", "demo.Status")), "flag provided but not defined: -type")
	for _, extra := range [][]string{nil, {"--prune", "--actor", "demo.UserActor"}, {"--prune", "--name", "demo.Status"}, {"--prune", "--actor", "demo.UserActor", "--name", "demo.Unused"}} {
		result := Run(append(append([]string{}, base...), extra...))
		var report skelc.ApiDependencyReport
		if result.ExitCode != 0 || json.Unmarshal([]byte(result.Stdout), &report) != nil {
			t.Fatalf("%+v", result)
		}
		if report.Domain != "demo" {
			t.Fatal(report)
		}
		if len(extra) == 3 && extra[1] == "--name" && (len(report.Services) != 0 || len(report.Enums) != 1) {
			t.Fatal(report)
		}
	}
	invalid := [][]string{{"--prune"}, {"--name", "demo.Status"}, {"--prune", "--name", "Status"}, {"--prune", "--name", "foreign.Status"}, {"--prune", "--name", "demo.UserActor"}, {"--prune", "--actor", "demo.MissingActor"}, {"--prune", "--name", "demo.Missing"}, {"--prune", "--name", ""}, {"extra"}}
	for _, args := range invalid {
		result := Run(append(append([]string{}, base...), args...))
		if result.ExitCode == 0 || !strings.Contains(result.Stdout, "INVALID_ARGUMENT") {
			t.Fatalf("accepted %v: %+v", args, result)
		}
	}
	result = Run([]string{"schema", "dep", "--skel-in", source})
	assertCommandErrorMessage(t, result, "flag api is required for schema dep")
	for _, target := range []string{"go", "go-module", "ts"} {
		args := []string{"gen", target, "--api", "--prune", "--name", "demo.Status", "--skel-in", source}
		out := filepath.Join(t.TempDir(), "api")
		if target == "ts" {
			args = append(args, "--ts-out", out)
		} else {
			args = append(args, "--go-out", out)
		}
		if target == "go-module" {
			args = append(args, "--go-module", "example.com/demoapi")
		}
		assertGenerationResult(t, Run(args))
		legacyArgs := append([]string{}, args...)
		legacyArgs[4] = "--type"
		assertCommandErrorMessage(t, Run(legacyArgs), "flag provided but not defined: -type")
		noAPI := append(append([]string{}, args[:2]...), args[3:]...)
		assertCommandErrorMessage(t, Run(noAPI), "flag prune requires api")
	}
}
