package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"slices"
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
	if result.ExitCode != 0 {
		t.Fatalf("complete dependency query failed: %+v", result)
	}
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

func TestSchemaDepViews(t *testing.T) {
	root := t.TempDir()
	source, shared, deep := filepath.Join(root, "demo.skel"), filepath.Join(root, "shared.skel"), filepath.Join(root, "deep.skel")
	writeCLIFile(t, deep, "domain deep\npub data Detail { value: string }\n")
	writeCLIFile(t, shared, `domain shared
import deep
pub data Box<TItem> { value: TItem detail: deep.Detail }
pub enum State { READY }
pub actor CallerActor { via client {} }
pub resource Record { action read }
`)
	writeCLIFile(t, source, `domain demo
import shared as s
pub data Unused { state: s.State }
data Hidden { box: s.Box<string> }
pub data Output { next: Output? box: s.Box<map<string,list<s.State>>> }
pub service ReadService { for s.CallerActor require s.Record:read method read { output Output } }
api service ReadApiService { for s.CallerActor via client auth anonymous method read { output Output } }
`)
	base := []string{"schema", "dep", "--skel-in", source, "--skel-import", "shared=" + shared, "--skel-import", "deep=" + deep}
	for _, mode := range []string{"", "--pub", "--api"} {
		args := append([]string{}, base...)
		if mode != "" {
			args = append(args, mode)
		}
		result := Run(args)
		var report skelc.SchemaDependencyReport
		if result.ExitCode != 0 || json.Unmarshal([]byte(result.Stdout), &report) != nil {
			t.Fatalf("%s: %+v", mode, result)
		}
		if strings.Contains(result.Stdout, "null") {
			t.Fatal(result.Stdout)
		}
		wantData := []string{"demo.Output", "demo.Unused"}
		if mode == "" {
			wantData = []string{"demo.Hidden", "demo.Output", "demo.Unused"}
		}
		if !slices.Equal(report.Data, wantData) {
			t.Fatalf("%s: %+v", mode, report)
		}
		want := []skelc.SchemaDeclarationDependency{{Domain: "shared", Name: "Box", Kind: "data"}, {Domain: "shared", Name: "State", Kind: "enum"}}
		if mode != "--api" {
			want = []skelc.SchemaDeclarationDependency{{Domain: "shared", Name: "Box", Kind: "data"}, {Domain: "shared", Name: "CallerActor", Kind: "actor"}, {Domain: "shared", Name: "Record", Kind: "resource"}, {Domain: "shared", Name: "State", Kind: "enum"}}
		}
		if !reflect.DeepEqual(report.Dependencies, want) {
			t.Fatalf("%s: deps=%+v", mode, report.Dependencies)
		}
		// Foreign members are not expanded; deep.Detail belongs to the shared query.
		if strings.Contains(result.Stdout, "deep.Detail") {
			t.Fatal(result.Stdout)
		}
		reversed := []string{"schema", "dep", "--skel-in", source, "--skel-import", "deep=" + deep, "--skel-import", "shared=" + shared}
		if mode != "" {
			reversed = append(reversed, mode)
		}
		if other := Run(reversed); other.ExitCode != 0 || other.Stdout != result.Stdout {
			t.Fatalf("unstable output: %+v", other)
		}
	}
	for _, extra := range [][]string{{"--pub", "--api"}, {"--actor", "shared.CallerActor"}, {"--prune"}, {"--name", "demo.Output"}, {"--pub", "--actor", "shared.CallerActor"}, {"--pub", "--prune"}, {"--pub", "--name", "demo.Output"}, {"--api", "--name", "demo.Output"}} {
		result := Run(append(append([]string{}, base...), extra...))
		if result.ExitCode == 0 || !strings.Contains(result.Stdout, "INVALID_ARGUMENT") {
			t.Fatalf("accepted %v: %+v", extra, result)
		}
	}
	for _, mode := range []string{"", "--pub", "--api"} {
		args := []string{"schema", "dep", "--skel-in", source, "--skel-import", "shared=" + shared}
		if mode != "" {
			args = append(args, mode)
		}
		result := Run(args)
		if result.ExitCode == 0 || !strings.Contains(result.Stdout, "deep") {
			t.Fatalf("missing transitive mapping ignored: %+v", result)
		}
	}
}

func TestSchemaDepEmptyAndStrict(t *testing.T) {
	source := filepath.Join(t.TempDir(), "empty.skel")
	writeCLIFile(t, source, "domain empty\n")
	for _, mode := range []string{"", "--pub", "--api"} {
		args := []string{"schema", "dep", "--skel-in", source}
		if mode != "" {
			args = append(args, mode)
		}
		result := Run(args)
		var report skelc.SchemaDependencyReport
		if result.ExitCode != 0 || json.Unmarshal([]byte(result.Stdout), &report) != nil || report.Domain != "empty" || strings.Contains(result.Stdout, "null") {
			t.Fatalf("%s: %+v", mode, result)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(result.Stdout), &fields); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"actors", "configs", "events", "resources", "webs", "tasks"} {
			value, present := fields[field]
			if present != (mode != "--api") || present && string(value) != "[]" {
				t.Fatalf("%s: unexpected %s: %s", mode, field, value)
			}
		}

	}
	writeCLIFile(t, source, "domain empty\nservice LegacyService { method ping {} }\n")
	for _, mode := range []string{"", "--pub", "--api"} {
		args := []string{"schema", "dep", "--skel-in", source}
		if mode != "" {
			args = append(args, mode)
		}
		result := Run(args)
		if result.ExitCode != 0 || !strings.Contains(result.Stderr, `"severity":"warning"`) {
			t.Fatalf("warnings lost: %+v", result)
		}
		result = Run(append([]string{"--strict"}, args...))
		if result.ExitCode == 0 || !strings.Contains(result.Stdout, "COMPILATION_FAILED") {
			t.Fatalf("strict ignored: %+v", result)
		}
	}
}

func TestSchemaDepAPIPreservesJSONOutput(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.skel")
	writeCLIFile(t, source, `domain demo
actor UserActor { via client {} }
actor IdleActor { via client {} }
pub data Unused { value: string }
pub enum Status { READY }
data Output { status: Status }
api service ReadApiService { for UserActor via client auth anonymous method read { output Output } }
`)
	cases := []struct {
		name      string
		flags     []string
		selection skelc.ApiFilter
	}{
		{name: "complete API"},
		{name: "filtered API", flags: []string{"--actor", "demo.UserActor"}, selection: skelc.ApiFilter{Actors: []string{"demo.UserActor"}}},
		{name: "pruned service", flags: []string{"--prune", "--actor", "demo.UserActor"}, selection: skelc.ApiFilter{Prune: true, Actors: []string{"demo.UserActor"}}},
		{name: "type only", flags: []string{"--prune", "--name", "demo.Unused"}, selection: skelc.ApiFilter{Prune: true, Types: []string{"demo.Unused"}}},
		{name: "empty", flags: []string{"--prune", "--actor", "demo.IdleActor"}, selection: skelc.ApiFilter{Prune: true, Actors: []string{"demo.IdleActor"}}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			legacy, err := skelc.QueryApiDependencies(skelc.Input{SkelIn: source}, test.selection)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := json.Marshal(legacy.Report)
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"schema", "dep", "--api", "--skel-in", source}, test.flags...)
			result := Run(args)
			if result.ExitCode != 0 {
				t.Fatalf("%+v", result)
			}
			var compact bytes.Buffer
			if err := json.Compact(&compact, []byte(result.Stdout)); err != nil {
				t.Fatal(err)
			}
			if compact.String() != string(expected) {
				t.Fatalf("API JSON changed:\ngot %s\nwant %s", compact.String(), expected)
			}
		})
	}
}
