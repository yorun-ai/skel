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
	for _, extra := range [][]string{nil, {"--prune", "--actor", "demo.UserActor"}, {"--prune", "--type", "demo.Status"}, {"--prune", "--actor", "demo.UserActor", "--type", "demo.Unused"}} {
		result := Run(append(append([]string{}, base...), extra...))
		var report skelc.ApiDependencyReport
		if result.ExitCode != 0 || json.Unmarshal([]byte(result.Stdout), &report) != nil {
			t.Fatalf("%+v", result)
		}
		if report.Domain != "demo" {
			t.Fatal(report)
		}
		if len(extra) == 3 && extra[1] == "--type" && (len(report.Services) != 0 || len(report.Enums) != 1) {
			t.Fatal(report)
		}
	}
	invalid := [][]string{{"--prune"}, {"--type", "demo.Status"}, {"--prune", "--type", "Status"}, {"--prune", "--type", "foreign.Status"}, {"--prune", "--type", "demo.UserActor"}, {"--prune", "--actor", "demo.MissingActor"}, {"--prune", "--type", "demo.Missing"}, {"--prune", "--type", ""}, {"extra"}}
	for _, args := range invalid {
		result := Run(append(append([]string{}, base...), args...))
		if result.ExitCode == 0 || !strings.Contains(result.Stdout, "INVALID_ARGUMENT") {
			t.Fatalf("accepted %v: %+v", args, result)
		}
	}
	result := Run([]string{"schema", "dep", "--skel-in", source})
	assertCommandErrorMessage(t, result, "flag api is required for schema dep")
	for _, target := range []string{"go", "go-module", "ts"} {
		args := []string{"gen", target, "--api", "--prune", "--type", "demo.Status", "--skel-in", source}
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
		noAPI := append(append([]string{}, args[:2]...), args[3:]...)
		assertCommandErrorMessage(t, Run(noAPI), "flag prune requires api")
	}
}
