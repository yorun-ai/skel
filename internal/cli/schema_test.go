package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/command"
	schemas "go.yorun.ai/skelc/internal/schema"
	"go.yorun.ai/skelc/internal/testutil"
)

func TestRunSkelcStrictSchemaCommands(t *testing.T) {
	dir := t.TempDir()
	entry, baseline := filepath.Join(dir, "order.skel"), filepath.Join(dir, "baseline.skel")
	writeCLIFile(t, entry, "domain demo.order\nservice OrderService { method ping {} }\n")
	writeCLIFile(t, baseline, "domain demo.order\nservice OrderService { method ping {} }\n")
	for _, args := range [][]string{
		{"schema", "list"},
		{"schema", "import"},
		{"schema", "get", "service", "demo.order.OrderService"},
		{"schema", "snapshot"},
		{"schema", "diff", "--baseline-skel-in", baseline},
	} {
		args = append(args, "--skel-in", entry)
		if result := Run(args); result.ExitCode != ExitCodeSuccess {
			t.Fatalf("compatible schema failed: %+v", result)
		}
		result := Run(append([]string{"--strict"}, args...))
		failure := decodeCommandError(t, result)
		if result.ExitCode != ExitCodeError || failure.Code != command.ErrorCodeCompilationFailed || !strings.Contains(result.Stderr, `"severity":"error"`) {
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
	var entries []*schemas.Entry
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
		missingTypeError.Code != command.ErrorCodeInvalidArgument || !strings.Contains(missingTypeError.Message, "expected TYPE SKEL_NAME") {
		t.Fatalf("expected missing type error: %+v", missingTypeResult)
	}

	getResult := Run([]string{"schema", "get", "data", "demo.user.User", "--skel-in", dir})
	if getResult.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected get result: %+v", getResult)
	}
	var declaration schemas.Declaration
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
			commandError.Code != command.ErrorCodeInvalidArgument || !strings.Contains(commandError.Message, "invalid schema declaration type") {
			t.Fatalf("expected invalid type error for %v: %+v", args, result)
		}
	}
}

func TestRunSkelcSchemaKeepsWarningsOnStderr(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(dir, ".hidden.skel"), `domain demo.user`)

	result := Run([]string{"schema", "list", "--skel-in", dir})
	var entries []*schemas.Entry
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
		code command.ErrorCode
	}{
		{name: "snapshot argument", args: []string{"schema", "snapshot"}, code: command.ErrorCodeInvalidArgument},
		{name: "diff argument", args: []string{"schema", "diff"}, code: command.ErrorCodeInvalidArgument},
		{name: "compilation", args: []string{"schema", "list", "--skel-in", invalidSource}, code: command.ErrorCodeCompilationFailed},
		{name: "diff candidate compilation", args: []string{"schema", "diff", "--baseline-skel-in", validSource, "--skel-in", invalidSource}, code: command.ErrorCodeCompilationFailed},
		{name: "diff baseline compilation", args: []string{"schema", "diff", "--baseline-skel-in", invalidSource, "--skel-in", validSource}, code: command.ErrorCodeCompilationFailed},
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
		{"schema", "list", "--unknown"},
		{"--log-format", "unknown", "schema", "list"},
	} {
		result := Run(args)
		commandError := decodeCommandError(t, result)
		if result.ExitCode != ExitCodeError || result.Stderr != "" ||
			commandError.Code != command.ErrorCodeInvalidArgument || commandError.Message == "" {
			t.Fatalf("expected structured parser error for %v: %+v", args, result)
		}
	}
}

func TestRunSkelcSchemaSnapshotFullDocument(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), `domain demo.user`)
	writeCLIFile(t, filepath.Join(dir, "data.skel"), `domain demo.user

pub data User {
    id: string
}

data InternalUser {
    id: string
}
`)

	result := Run([]string{"schema", "snapshot", "--skel-in", dir})
	if result.ExitCode != ExitCodeSuccess || result.Stdout == "" || result.Stderr != "" {
		t.Fatalf("unexpected snapshot result: %+v", result)
	}
	document, err := schemas.Decode(strings.NewReader(result.Stdout))
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Declarations) != 2 || schemas.Find(document, "data", "demo.user.User") == nil {
		t.Fatalf("unexpected schema: %+v", document)
	}
	internal := schemas.Find(document, "data", "demo.user.InternalUser")
	if internal == nil || internal.Pub {
		t.Fatalf("expected private declaration in full schema: %+v", internal)
	}
}

func TestRunSkelcSchemaSnapshotKeepsImportsOpaque(t *testing.T) {
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

	shallow := Run([]string{"schema", "snapshot", "--skel-in", source})
	if shallow.ExitCode != ExitCodeSuccess {
		t.Fatalf("export with opaque imports failed: %+v", shallow)
	}
	aliasSnapshot := Run([]string{"schema", "snapshot", "--skel-in", aliasSource})
	if aliasSnapshot.ExitCode != ExitCodeSuccess {
		t.Fatalf("snapshot with alternate import alias failed: %+v", aliasSnapshot)
	}
	if shallow.Stdout != aliasSnapshot.Stdout {
		t.Fatalf("import alias changed schema output:\nidentity alias:\n%s\naccount alias:\n%s", shallow.Stdout, aliasSnapshot.Stdout)
	}
	document, err := schemas.Decode(strings.NewReader(shallow.Stdout))
	if err != nil {
		t.Fatal(err)
	}
	order := schemas.Find(document, "data", "demo.order.Order")
	if order == nil || order.Data.Members[0].Type.Kind != "importedReference" || order.Data.Members[0].Type.Name != "demo.user.User" {
		t.Fatalf("unexpected imported data reference: %+v", order)
	}
	service := schemas.Find(document, "service", "demo.order.OrderService")
	if service == nil || service.Service.Require == nil || service.Service.Require.Mode != "reference" ||
		service.Service.Require.Check.Resource != "demo.user.User" {
		t.Fatalf("unexpected imported permission reference: %+v", service)
	}
	diff := Run([]string{
		"schema", "diff", "--baseline-skel-in", source, "--skel-in", aliasSource,
	})
	var diffReport schemas.Report
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
	var report schemas.Report
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
	var report schemas.Report
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
			var report schemas.Report
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
		commandError.Code != command.ErrorCodeGitHistoryNotFound ||
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
	if result.ExitCode != ExitCodeError || commandError.Code != command.ErrorCodeCompilationFailed ||
		!strings.Contains(commandError.Message, "HEAD:skel/data.skel") ||
		strings.Contains(commandError.Message, "skelc-schema-baseline-") {
		t.Fatalf("expected stable Git baseline error path: %+v", result)
	}
}

func TestSchemaImportListsDirectDeclarationsWithoutLoadingDependencies(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), "domain demo.main\n")
	first := filepath.Join(dir, "a.skel")
	second := filepath.Join(dir, "b.skel")
	writeCLIFile(t, first, "domain demo.main\n// import ignored.comment\nimport external.shared\n\npub data Payload { value: external.shared.Value }\n")
	writeCLIFile(t, second, "domain demo.main\n/* import ignored.block */\nimport external.shared as shared\n@desc(\"\"\"\nimport ignored.description\n\"\"\")\npub data Other { value: shared.Value }\n")
	// The imported domain is not present on disk and needs no mapping.
	args := []string{"schema", "import", "--skel-in", dir}
	result := Run(args)
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("import query failed: %+v", result)
	}
	var imports []command.SchemaImport
	if err := json.Unmarshal([]byte(result.Stdout), &imports); err != nil {
		t.Fatal(err)
	}
	if len(imports) != 2 {
		t.Fatalf("expected both source declarations, got %+v", imports)
	}
	for i, want := range []command.SchemaImport{
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

func TestSchemaImportEmptyAndInvalidInputs(t *testing.T) {
	root := t.TempDir()
	empty, broken := filepath.Join(root, "empty.skel"), filepath.Join(root, "broken.skel")
	writeCLIFile(t, empty, "domain demo.empty\npub data Value { id: string }\n")
	writeCLIFile(t, broken, "domain demo.broken\nimport\n")
	result := Run([]string{"schema", "import", "--skel-in", empty})
	if result.ExitCode != ExitCodeSuccess || strings.TrimSpace(result.Stdout) != "[]" {
		t.Fatalf("expected empty array, got %+v", result)
	}
	for _, tc := range []struct {
		args []string
		code command.ErrorCode
	}{
		{[]string{"schema", "import"}, command.ErrorCodeInvalidArgument},
		{[]string{"schema", "import", "extra", "--skel-in", empty}, command.ErrorCodeInvalidArgument},
		{[]string{"schema", "import", "--skel-in", broken}, command.ErrorCodeCompilationFailed},
		{[]string{"schema", "import", "--skel-in", filepath.Join(root, "missing.skel")}, command.ErrorCodeCompilationFailed},
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
