package skelc

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/cmd/skelc/output"
	"go.yorun.ai/skel/internal/testutil"
	schemadiff "go.yorun.ai/skel/schema/diff"
)

func TestRunSkelcStrictSchemaCommands(t *testing.T) {
	dir := t.TempDir()
	entry, baseline := filepath.Join(dir, "order.skel"), filepath.Join(dir, "baseline.skel")
	writeCLIFile(t, entry, "domain demo.order\nservice OrderService { method ping {} }\n")
	writeCLIFile(t, baseline, "domain demo.order\nservice OrderService { method ping {} }\n")
	for _, args := range [][]string{
		{"schema", "list"},
		{"schema", "get", "service", "demo.order.OrderService"},
		{"schema", "diff", "--baseline-skel-in", baseline},
	} {
		args = append(args, "--skel-in", entry)
		if result := Run(args); result.ExitCode != ExitCodeSuccess {
			t.Fatalf("compatible schema failed: %+v", result)
		}
		result := Run(append([]string{"--strict"}, args...))
		failure := decodeCommandError(t, result)
		if result.ExitCode != ExitCodeError || failure.Code != output.ErrorCodeCompilationFailed || !strings.Contains(result.Stderr, `"severity":"error"`) {
			t.Fatalf("expected strict schema failure: %+v", result)
		}
	}
	writeCLIFile(t, entry, "domain demo.order\npub service OrderService { method ping {} }\n")
	result := Run([]string{"--strict", "schema", "diff", "--skel-in", entry, "--baseline-skel-in", baseline})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("historical baseline blocked migration comparison: %+v", result)
	}
}

func TestRunSkelcSchemaListAndGet(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(dir, "schema.skel"), `domain demo.user

pub data User {
    id: string
}

pub resource User {
    action view
}
`)

	listResult := Run([]string{"schema", "list", "--skel-in", dir})
	if listResult.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected list result: %+v", listResult)
	}
	var entries []*output.SchemaEntry
	if err := json.Unmarshal([]byte(listResult.Stdout), &entries); err != nil {
		t.Fatalf("decode list result: %v\n%s", err, listResult.Stdout)
	}
	if len(entries) != 2 || entries[0].Kind != "data" || entries[1].Kind != "resource" {
		t.Fatalf("unexpected list entries: %+v", entries)
	}
	filteredResult := Run([]string{"schema", "list", "data", "--skel-in", dir})
	if filteredResult.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected filtered list result: %+v", filteredResult)
	}
	entries = nil
	if err := json.Unmarshal([]byte(filteredResult.Stdout), &entries); err != nil || len(entries) != 1 || entries[0].Kind != "data" {
		t.Fatalf("unexpected filtered list entries: %+v, err=%v", entries, err)
	}

	missingTypeResult := Run([]string{"schema", "get", "demo.user.User", "--skel-in", dir})
	missingTypeError := decodeCommandError(t, missingTypeResult)
	if missingTypeResult.ExitCode != ExitCodeError || missingTypeResult.Stderr != "" ||
		missingTypeError.Code != output.ErrorCodeInvalidArgument || !strings.Contains(missingTypeError.Message, "expected TYPE SKEL_NAME") {
		t.Fatalf("expected missing type error: %+v", missingTypeResult)
	}

	getResult := Run([]string{"schema", "get", "data", "demo.user.User", "--skel-in", dir})
	if getResult.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected get result: %+v", getResult)
	}
	var declaration output.SchemaDeclaration
	if err := json.Unmarshal([]byte(getResult.Stdout), &declaration); err != nil {
		t.Fatalf("decode declaration: %v\n%s", err, getResult.Stdout)
	}
	if declaration.Data == nil || len(declaration.Data.Members) != 1 || declaration.Data.Members[0].Name != "id" {
		t.Fatalf("unexpected declaration: %+v", declaration)
	}

	missingResult := Run([]string{"schema", "get", "web", "demo.user.Missing", "--skel-in", dir})
	if missingResult.ExitCode != ExitCodeSuccess || missingResult.Stdout != "null\n" || missingResult.Stderr != "" {
		t.Fatalf("unexpected missing declaration result: %+v", missingResult)
	}
}

func TestRunSkelcSchemaGetActorCapabilities(t *testing.T) {
	for _, test := range []struct {
		name       string
		auth       bool
		permission bool
	}{
		{name: "neither"},
		{name: "auth", auth: true},
		{name: "permission", permission: true},
		{name: "both", auth: true, permission: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			entry := filepath.Join(t.TempDir(), "actor.skel")
			source := "domain demo.user\npub actor ClientActor { via client {}\n"
			if test.auth {
				source += "auth { credential { token: string } info { @identifier id: uuid } }\n"
			}
			if test.permission {
				source += "permission {}\n"
			}
			writeCLIFile(t, entry, source+"}\n")
			result := Run([]string{"schema", "get", "actor", "demo.user.ClientActor", "--skel-in", entry})
			if result.ExitCode != ExitCodeSuccess {
				t.Fatalf("actor query failed: %+v", result)
			}
			var declaration struct {
				Actor map[string]json.RawMessage `json:"actor"`
			}
			if err := json.Unmarshal([]byte(result.Stdout), &declaration); err != nil {
				t.Fatal(err)
			}
			actor := declaration.Actor
			if string(actor["vias"]) == "" || (actor["auth"] != nil) != test.auth || (actor["permission"] != nil) != test.permission {
				t.Fatalf("unexpected actor capabilities: %s", result.Stdout)
			}
			for key := range actor {
				if key != "vias" && key != "auth" && key != "permission" {
					t.Fatalf("unexpected flattened actor field %q", key)
				}
			}
			if test.permission && strings.TrimSpace(string(actor["permission"])) != "{}" {
				t.Fatalf("expected permission declaration without derived services: %s", actor["permission"])
			}
			if test.auth {
				var auth map[string]json.RawMessage
				if err := json.Unmarshal(actor["auth"], &auth); err != nil {
					t.Fatal(err)
				}
				if len(auth) != 3 || auth["credential"] == nil || auth["info"] == nil || string(auth["identifierField"]) != `"id"` {
					t.Fatalf("unexpected authentication declaration: %s", actor["auth"])
				}
			}
		})
	}
}

func TestRunSkelcSchemaQueryRejectsInvalidType(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), `domain demo.user`)

	for _, args := range [][]string{
		{"schema", "list", "unknown", "--skel-in", dir},
		{"schema", "get", "unknown", "demo.user.User", "--skel-in", dir},
	} {
		result := Run(args)
		commandError := decodeCommandError(t, result)
		if result.ExitCode != ExitCodeError || result.Stderr != "" ||
			commandError.Code != output.ErrorCodeInvalidArgument || !strings.Contains(commandError.Message, "invalid schema declaration type") {
			t.Fatalf("expected invalid type error for %v: %+v", args, result)
		}
	}
}

func TestRunSkelcSchemaKeepsWarningsOnStderr(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(dir, ".hidden.skel"), `domain demo.user`)

	result := Run([]string{"schema", "list", "--skel-in", dir})
	var entries []*output.SchemaEntry
	if err := json.Unmarshal([]byte(result.Stdout), &entries); err != nil {
		t.Fatalf("decode schema result: %v\n%s", err, result.Stdout)
	}
	logEntry := new(struct {
		Level string `json:"level"`
	})
	if err := json.Unmarshal([]byte(strings.TrimSpace(result.Stderr)), logEntry); err != nil {
		t.Fatalf("decode schema log: %v\n%s", err, result.Stderr)
	}
	if result.ExitCode != ExitCodeSuccess || len(entries) != 0 || logEntry.Level != logLevelWarn {
		t.Fatalf("unexpected schema result: %+v", result)
	}
}

func TestRunSkelcSchemaErrorsUseStdoutResult(t *testing.T) {
	invalidSource := t.TempDir()
	writeCLIFile(t, filepath.Join(invalidSource, "domain.skel"), "domain demo.user")
	writeCLIFile(t, filepath.Join(invalidSource, "data.skel"), "domain demo.user\n\ndata User { id string }")
	validSource := t.TempDir()
	writeCLIFile(t, filepath.Join(validSource, "domain.skel"), "domain demo.user")

	for _, test := range []struct {
		name string
		args []string
		code output.ErrorCode
	}{
		{name: "diff argument", args: []string{"schema", "diff"}, code: output.ErrorCodeInvalidArgument},
		{name: "compilation", args: []string{"schema", "list", "--skel-in", invalidSource}, code: output.ErrorCodeCompilationFailed},
		{name: "diff candidate compilation", args: []string{"schema", "diff", "--baseline-skel-in", validSource, "--skel-in", invalidSource}, code: output.ErrorCodeCompilationFailed},
		{name: "diff baseline compilation", args: []string{"schema", "diff", "--baseline-skel-in", invalidSource, "--skel-in", validSource}, code: output.ErrorCodeCompilationFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := Run(test.args)
			if result.ExitCode != ExitCodeError {
				t.Fatalf("expected error result: %+v", result)
			}
			commandError := decodeCommandError(t, result)
			if commandError.Code != test.code || commandError.Message == "" {
				t.Fatalf("unexpected command error: %+v", commandError)
			}
		})
	}
}

func TestRunSkelcSchemaParserErrorsUseStdoutResult(t *testing.T) {
	for _, args := range [][]string{
		{"schema", "unknown"},
		{"schema", "snapshot"},
		{"schema", "list", "--unknown"},
		{"--log-format", "unknown", "schema", "list"},
	} {
		result := Run(args)
		commandError := decodeCommandError(t, result)
		if result.ExitCode != ExitCodeError || result.Stderr != "" ||
			commandError.Code != output.ErrorCodeInvalidArgument || commandError.Message == "" {
			t.Fatalf("expected structured parser error for %v: %+v", args, result)
		}
	}
}

func TestRunSkelcSchemaGetKeepsImportsOpaque(t *testing.T) {
	source := t.TempDir()
	aliasSource := t.TempDir()
	writeCLIFile(t, filepath.Join(source, "domain.skel"), `domain demo.order`)
	writeCLIFile(t, filepath.Join(aliasSource, "domain.skel"), `domain demo.order`)
	sourceSchema := `domain demo.order

import demo.user as identity

pub data Order {
    owner: identity.User
}

pub service OrderService {
    require identity.User:read

    method get {
        output Order
    }
}
`
	writeCLIFile(t, filepath.Join(source, "schema.skel"), sourceSchema)
	writeCLIFile(t, filepath.Join(aliasSource, "schema.skel"), strings.ReplaceAll(sourceSchema, "identity", "account"))

	get := func(kind, name string) *output.SchemaDeclaration {
		t.Helper()
		result := Run([]string{"schema", "get", kind, name, "--skel-in", source})
		aliasResult := Run([]string{"schema", "get", kind, name, "--skel-in", aliasSource})
		if result.ExitCode != ExitCodeSuccess || aliasResult.ExitCode != ExitCodeSuccess {
			t.Fatalf("query with opaque imports failed: %+v; %+v", result, aliasResult)
		}
		if result.Stdout != aliasResult.Stdout {
			t.Fatalf("import alias changed schema output: %s vs %s", result.Stdout, aliasResult.Stdout)
		}
		var declaration output.SchemaDeclaration
		if err := json.Unmarshal([]byte(result.Stdout), &declaration); err != nil {
			t.Fatal(err)
		}
		return &declaration
	}
	order := get("data", "demo.order.Order")
	if order.Data.Members[0].Type.Kind != "importedReference" || order.Data.Members[0].Type.Name != "demo.user.User" {
		t.Fatalf("unexpected imported data reference: %+v", order)
	}
	service := get("service", "demo.order.OrderService")
	if service.Service.Require == nil || service.Service.Require.Mode != "reference" ||
		service.Service.Require.Check.Resource != "demo.user.User" {
		t.Fatalf("unexpected imported permission reference: %+v", service)
	}
	diff := Run([]string{
		"schema", "diff", "--baseline-skel-in", source, "--skel-in", aliasSource,
	})
	var diffReport schemadiff.Report
	if err := json.Unmarshal([]byte(diff.Stdout), &diffReport); err != nil {
		t.Fatalf("decode diff report: %v\n%s", err, diff.Stdout)
	}
	if diff.ExitCode != ExitCodeSuccess || !diffReport.Compatible || len(diffReport.Changes) != 0 {
		t.Fatalf("source diff should not require imports or observe aliases: %+v", diff)
	}
}

func TestRunSkelcSchemaDiffOutputsCompleteReport(t *testing.T) {
	baseline := t.TempDir()
	candidate := t.TempDir()
	writeCLIFile(t, filepath.Join(baseline, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(baseline, "data.skel"), `domain demo.user

pub data User {
    id: string
}
`)
	writeCLIFile(t, filepath.Join(candidate, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(candidate, "data.skel"), `domain demo.user

pub data User {
    id: string
    name: string
}
`)

	result := Run([]string{"schema", "diff", "--baseline-skel-in", baseline, "--skel-in", candidate})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("expected completed diff, got %+v", result)
	}
	var report schemadiff.Report
	if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
		t.Fatalf("decode diff report: %v\n%s", err, result.Stdout)
	}
	if result.Stderr != "" || report.Compatible || report.Summary.Breaking != 1 ||
		report.Changes[0].Change != "ADDED" || report.Changes[0].Impact != "BREAKING" ||
		report.Changes[0].Code != "data.member.added" {
		t.Fatalf("unexpected diff output: %+v", result)
	}
}

func TestRunSkelcSchemaDiffChecksPrivateDeclarations(t *testing.T) {
	baseline := t.TempDir()
	candidate := t.TempDir()
	writeCLIFile(t, filepath.Join(baseline, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(baseline, "data.skel"), `domain demo.user

data InternalUser {
    id: string
}
`)
	writeCLIFile(t, filepath.Join(candidate, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(candidate, "data.skel"), `domain demo.user

data InternalUser {
    id: string
    name: string
}
`)

	result := Run([]string{
		"schema", "diff", "--baseline-skel-in", baseline, "--skel-in", candidate,
	})
	if result.ExitCode != ExitCodeSuccess || !strings.Contains(result.Stdout, "data.member.added") {
		t.Fatalf("expected private declaration incompatibility: %+v", result)
	}
}

func TestRunSkelcSchemaDiffReportsDangerousSourceChange(t *testing.T) {
	baseline := t.TempDir()
	candidate := t.TempDir()
	writeCLIFile(t, filepath.Join(baseline, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(baseline, "data.skel"), `domain demo.user

pub enum UserStatus {
    ACTIVE
}
`)
	writeCLIFile(t, filepath.Join(candidate, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(candidate, "data.skel"), `domain demo.user

pub enum UserStatus {
    ACTIVE
    DISABLED
}
`)

	result := Run([]string{
		"schema", "diff", "--baseline-skel-in", baseline, "--skel-in", candidate,
	})
	if result.ExitCode != ExitCodeSuccess || result.Stderr != "" {
		t.Fatalf("unexpected diff result: %+v", result)
	}
	var report schemadiff.Report
	if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
		t.Fatalf("decode report: %v\n%s", err, result.Stdout)
	}
	if !report.Compatible || report.Summary.Dangerous != 1 || report.Changes[0].Code != "enum.item.added" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRunSkelcSchemaDiffUsesGitHeadBaseline(t *testing.T) {
	testutil.RequireGit(t)
	for _, test := range []struct {
		name             string
		directoryInput   bool
		expectedBaseline string
	}{
		{name: "directory", directoryInput: true, expectedBaseline: "HEAD:skel/data.skel"},
		{name: "single file", expectedBaseline: "HEAD:user.skel"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := t.TempDir()
			skelIn := filepath.Join(repository, "user.skel")
			sourceFile := skelIn
			baselineSource := `domain demo.user

pub data User {
    id: string
}
`
			candidateSource := strings.Replace(baselineSource, "id: string", "id: int", 1)
			if test.directoryInput {
				skelIn = filepath.Join(repository, "skel")
				sourceFile = filepath.Join(skelIn, "data.skel")
				writeCLIFile(t, filepath.Join(skelIn, "domain.skel"), "domain demo.user")
			}
			writeCLIFile(t, sourceFile, baselineSource)
			testutil.InitRepository(t, repository)
			testutil.Commit(t, repository, "baseline")
			writeCLIFile(t, sourceFile, candidateSource)

			result := Run([]string{"schema", "diff", "--skel-in", skelIn})
			if result.ExitCode != ExitCodeSuccess || result.Stderr != "" {
				t.Fatalf("unexpected Git diff result: %+v", result)
			}
			var report schemadiff.Report
			if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
				t.Fatalf("decode Git diff report: %v\n%s", err, result.Stdout)
			}
			if len(report.Changes) != 1 || report.Changes[0].Code != "data.member.type.changed" {
				t.Fatalf("unexpected Git diff report: %+v", report)
			}
			if report.Changes[0].Baseline == nil || report.Changes[0].Baseline.File != test.expectedBaseline {
				t.Fatalf("unexpected Git baseline position: %+v", report.Changes[0].Baseline)
			}
		})
	}
}

func TestRunSkelcSchemaDiffRequiresBaselineWithoutGitHistory(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), "domain demo.user")

	result := Run([]string{"schema", "diff", "--skel-in", dir})
	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || result.Stderr != "" ||
		commandError.Code != output.ErrorCodeGitHistoryNotFound ||
		!strings.Contains(commandError.Message, "git history not found") ||
		!strings.Contains(commandError.Message, "--baseline-skel-in") {
		t.Fatalf("expected missing Git history guidance: %+v", result)
	}
}

func TestRunSkelcSchemaDiffRemapsInvalidGitBaselinePath(t *testing.T) {
	testutil.RequireGit(t)
	repository := t.TempDir()
	skelDir := filepath.Join(repository, "skel")
	writeCLIFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.user")
	dataPath := filepath.Join(skelDir, "data.skel")
	writeCLIFile(t, dataPath, `domain demo.user

pub data User {
    id string
}
`)
	testutil.InitRepository(t, repository)
	testutil.Commit(t, repository, "invalid baseline")
	writeCLIFile(t, dataPath, `domain demo.user

pub data User {
    id: string
}
`)

	result := Run([]string{"schema", "diff", "--skel-in", skelDir})
	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || commandError.Code != output.ErrorCodeCompilationFailed ||
		!strings.Contains(commandError.Message, "HEAD:skel/data.skel") ||
		strings.Contains(commandError.Message, "skelc-schema-baseline-") {
		t.Fatalf("expected stable Git baseline error path: %+v", result)
	}
}

func TestSchemaListGenerationViews(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.skel")
	writeCLIFile(t, source, `domain demo
pub actor UserActor { via client {} }
actor WriterActor { via client {} }
enum Status { READY DONE }
data Value { status: Status }
pub data Unused { value: string }
pub config SettingsConfig eternal { enabled: bool }
pub resource Record { action view }
pub event ChangedEvent { payload { value: string } }
ext event StoredEvent { payload { value: string } }
pub service BackendService { method read { output Value } }
ext service StorageService { method read { output Value } }
api service ReadApiService { for UserActor via client auth anonymous method read { output Value } }
api service WriteApiService { for WriterActor via client auth anonymous method write { output string } }
`)
	for _, test := range []struct {
		name  string
		flags []string
		want  []string
	}{
		{"public", []string{"--pub"}, []string{"actor:UserActor", "config:SettingsConfig", "data:Unused", "data:Value", "enum:Status", "event:ChangedEvent", "event:StoredEvent", "resource:Record", "service:BackendService", "service:StorageService"}},
		{"api", []string{"--api"}, []string{"data:Unused", "data:Value", "enum:Status", "service:ReadApiService", "service:WriteApiService"}},
		{"actor without pruning", []string{"--api", "--actor", "demo.UserActor"}, []string{"data:Unused", "data:Value", "enum:Status", "service:ReadApiService"}},
		{"pruned actor", []string{"--api", "--prune", "--actor", "demo.UserActor"}, []string{"data:Value", "enum:Status", "service:ReadApiService"}},
		{"type only", []string{"--api", "--prune", "--name", "demo.Value"}, []string{"data:Value", "enum:Status"}},
		{"repeated names", []string{"--api", "--prune", "--name", "demo.Value", "--name", "demo.Unused"}, []string{"data:Unused", "data:Value", "enum:Status"}},
		{"actor and name union", []string{"--api", "--prune", "--actor", "demo.UserActor", "--name", "demo.Unused"}, []string{"data:Unused", "data:Value", "enum:Status", "service:ReadApiService"}},
		{"kind after selection", []string{"--api", "--prune", "--actor", "demo.UserActor", "enum"}, []string{"enum:Status"}},
		{"no matching services", []string{"--api", "--prune", "--actor", "demo.WriterActor", "data"}, []string{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			args := append([]string{"schema", "list", "--skel-in", source}, test.flags...)
			result := Run(args)
			var entries []*output.SchemaEntry
			if result.ExitCode != ExitCodeSuccess || json.Unmarshal([]byte(result.Stdout), &entries) != nil {
				t.Fatalf("%+v", result)
			}
			got := make([]string, 0, len(entries))
			for _, entry := range entries {
				got = append(got, string(entry.Kind)+":"+entry.Name)
				if entry.SkelName != "demo."+entry.Name {
					t.Fatalf("noncanonical entry: %+v", entry)
				}
				wantPub := slices.Contains([]string{"UserActor", "Unused", "SettingsConfig", "Record", "ChangedEvent", "BackendService"}, entry.Name)
				if entry.Pub != wantPub {
					t.Fatalf("selection changed public attribute: %+v", entry)
				}
				if test.name == "public" {
					result := Run([]string{"schema", "get", string(entry.Kind), entry.SkelName, "--skel-in", source})
					var declaration output.SchemaDeclaration
					if result.ExitCode != ExitCodeSuccess || json.Unmarshal([]byte(result.Stdout), &declaration) != nil || declaration.Pub != wantPub {
						t.Fatalf("get changed declaration pub: %+v", result)
					}
					if entry.Name == "StorageService" && (declaration.Service == nil || !declaration.Service.Ext) ||
						entry.Name == "StoredEvent" && (declaration.Data == nil || !declaration.Data.Ext) {
						t.Fatalf("get lost extension flag: %+v", result)
					}
				}
			}
			slices.Sort(got)
			if !slices.Equal(got, test.want) || entries == nil {
				t.Fatalf("got %v, want %v; output=%s", got, test.want, result.Stdout)
			}
			if again := Run(args); again.Stdout != result.Stdout {
				t.Fatalf("unstable output: %s / %s", result.Stdout, again.Stdout)
			}
		})
	}
}

func TestSchemaListViewImportsAndErrors(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.skel")
	shared := filepath.Join(t.TempDir(), "source.skel")
	writeCLIFile(t, shared, "domain shared\npub data Money { value: int }\n")
	writeCLIFile(t, source, "domain demo\nimport shared as external\npub data Value { money: external.Money }\n")
	base := []string{"schema", "list", "--skel-in", source}
	for _, mode := range []string{"--pub", "--api"} {
		result := Run(append(append([]string{}, base...), mode, "--skel-import", "shared="+shared))
		var entries []*output.SchemaEntry
		if result.ExitCode != ExitCodeSuccess || json.Unmarshal([]byte(result.Stdout), &entries) != nil || len(entries) != 1 || entries[0].SkelName != "demo.Value" {
			t.Fatalf("foreign declarations leaked into %s: %+v", mode, result)
		}
		failure := decodeCommandError(t, Run(append(append([]string{}, base...), mode)))
		if failure.Code != output.ErrorCodeCompilationFailed || !strings.Contains(failure.Message, "shared") {
			t.Fatalf("missing import hidden: %+v", failure)
		}
	}
	for _, flags := range [][]string{
		{"--pub", "--api"}, {"--actor", "demo.UserActor"}, {"--prune"}, {"--name", "demo.Value"},
		{"--pub", "--actor", "demo.UserActor"}, {"--api", "--prune"}, {"--api", "--name", "demo.Value"},
		{"--api", "--prune", "--name", "Value"}, {"--skel-import", "shared=" + shared},
		{"--api", "--skel-import", "broken"}, {"--api", "--prune", "--type", "demo.Value"},
	} {
		result := Run(append(append([]string{}, base...), flags...))
		failure := decodeCommandError(t, result)
		if failure.Code != output.ErrorCodeInvalidArgument {
			t.Fatalf("wrong error for %v: %+v", flags, result)
		}
	}
	writeCLIFile(t, source, "domain demo\nservice LegacyService { method ping {} }\n")
	for _, mode := range []string{"--pub", "--api"} {
		result := Run([]string{"schema", "list", mode, "--skel-in", source})
		if result.ExitCode != ExitCodeSuccess || result.Stdout != "[]\n" || !strings.Contains(result.Stderr, `"severity":"warning"`) {
			t.Fatalf("warnings mixed with empty result: %+v", result)
		}
		failure := decodeCommandError(t, Run([]string{"--strict", "schema", "list", mode, "--skel-in", source}))
		if failure.Code != output.ErrorCodeCompilationFailed {
			t.Fatalf("strict ignored: %+v", failure)
		}
	}
	writeCLIFile(t, source, "domain demo\nactor UserActor { via client {} }\npub service ReadService { for UserActor method ping {} }\n")
	failure := decodeCommandError(t, Run([]string{"schema", "list", "--pub", "--skel-in", source}))
	if failure.Code != output.ErrorCodeCompilationFailed || !strings.Contains(failure.Message, "non-pub actor") {
		t.Fatalf("public view validation skipped: %+v", failure)
	}
}
