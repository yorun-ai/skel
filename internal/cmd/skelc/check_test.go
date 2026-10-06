package skelc

import (
	"encoding/json"
	"strings"
	"testing"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/cmd/skelc/output"
	compiler "go.yorun.ai/skel/internal/compiler"
)

func TestRunSkelcCheckStrictMigrationDiagnostics(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", "domain demo.order")
	writeCLIFile(t, dir+"/service.skel", "domain demo.order\nservice LegacyService { method ping {} }\npub service DualService { noauth method ping {} }\n")
	writeCLIFile(t, dir+"/.hidden.skel", "ignored")
	for _, test := range []struct {
		name   string
		args   []string
		strict bool
	}{
		{name: "default", args: []string{"check"}},
		{name: "disabled", args: []string{"--strict=false", "check"}},
		{name: "global", args: []string{"--strict", "check"}, strict: true},
		{name: "inherited", args: []string{"check", "--strict"}, strict: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := Run(append(test.args, "--skel-in", dir))
			checked := decodeCheckResult(t, result)
			exitCode := ExitCodeSuccess
			if test.strict {
				exitCode = ExitCodeUnsatisfied
			}
			if result.ExitCode != exitCode || checked.Valid == test.strict || len(checked.Diagnostics) != 4 {
				t.Fatalf("unexpected strict check: %+v", result)
			}
			for _, item := range checked.Diagnostics {
				severity := compiler.DiagnosticSeverityWarning
				if test.strict && item.Code != diagnostic.CodeLoaderHiddenFile {
					severity = compiler.DiagnosticSeverityError
				}
				if item.Severity != severity {
					t.Fatalf("unexpected severity: %+v", item)
				}
			}
		})
	}
}

func TestRunSkelcCheck(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeSuccess || !checked.Valid || len(checked.Diagnostics) != 0 || result.Stderr != "" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunSkelcCheckReturnsSyntaxDiagnostics(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)
	writeCLIFile(t, dir+"/service.skel", `domain demo.user

actor ClientActor {
    via client {}
}

service UserService {
    allow ClientActor

    method ping {}
}`)

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeUnsatisfied || checked.Valid ||
		!checkDiagnosticsContain(checked, diagnostic.CodeSyntaxUnexpected, `unexpected token "allow"`) || result.Stderr != "" {
		t.Fatalf("unexpected check result: %+v", result)
	}
}

func TestRunSkelcCheckIncludesLoaderWarnings(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)
	writeCLIFile(t, dir+"/.hidden.skel", `domain demo.user`)

	for _, args := range [][]string{
		{"check", "--skel-in", dir},
		{"check", "--log-format", "jsonl", "--skel-in", dir},
	} {
		result := Run(args)
		checked := decodeCheckResult(t, result)
		if result.ExitCode != ExitCodeSuccess || !checked.Valid || len(checked.Diagnostics) != 1 ||
			checked.Diagnostics[0].Severity != compiler.DiagnosticSeverityWarning ||
			!strings.Contains(checked.Diagnostics[0].Message, ".hidden.skel ignored (HIDDEN_FILE)") {
			t.Fatalf("unexpected loader warning result: %+v", result)
		}
	}
}

func TestRunSkelcCheckErrorsUseCommandResult(t *testing.T) {
	result := Run([]string{"check"})
	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || result.Stderr != "" ||
		commandError.Code != output.ErrorCodeInvalidArgument || commandError.Message != "missing flag skel-in" {
		t.Fatalf("unexpected command error: %+v", result)
	}
}

func TestRunSkelcCheckReturnsMultipleStructuredSyntaxDiagnostics(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", "domain demo.user")
	writeCLIFile(t, dir+"/types.skel", `domain demo.user
data User {
    first string
    second:
}
`)

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeUnsatisfied || checked.Valid || len(checked.Diagnostics) != 2 {
		t.Fatalf("expected two diagnostics: %+v", result)
	}
	for _, item := range checked.Diagnostics {
		if !strings.HasPrefix(item.Code, "syntax.") || item.Severity != compiler.DiagnosticSeverityError ||
			item.Range.End.Column <= item.Range.Start.Column {
			t.Fatalf("unexpected structured diagnostic: %+v", item)
		}
	}
}

func TestRunSkelcCheckDoesNotRequireSkelImport(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.booker`)
	writeCLIFile(t, dir+"/types.skel", `domain demo.booker
import demo.user as user

data Booking {
    owner: user.User
}`)

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeSuccess || !checked.Valid || len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunSkelcCheckStillValidatesLocalDeclarationsWithMissingImports(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.booker`)
	writeCLIFile(t, dir+"/types.skel", `domain demo.booker
import demo.user as user

data invalidName {
    owner: user.User
}
`)

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeUnsatisfied ||
		!checkDiagnosticsContain(checked, diagnostic.CodeSemanticNaming, "Data") ||
		checkDiagnosticsContainCode(checked, diagnostic.CodeImportMissing) {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunSkelcCheckReportsMultipleSemanticErrors(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)
	writeCLIFile(t, dir+"/types.skel", `domain demo.user
data User { missing: MissingUser }
data Order { missing: MissingOrder }
`)

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeUnsatisfied || len(checked.Diagnostics) != 2 ||
		!checkDiagnosticsContain(checked, diagnostic.CodeSemanticReference, "definition of MissingUser not found") ||
		!checkDiagnosticsContain(checked, diagnostic.CodeSemanticReference, "definition of MissingOrder not found") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunSkelcCheckReportsSyntaxErrorsFromMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)
	writeCLIFile(t, dir+"/user.skel", "domain demo.user\ndata User {")
	writeCLIFile(t, dir+"/order.skel", "domain demo.user\ndata Order {")

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeUnsatisfied || len(checked.Diagnostics) != 2 ||
		!strings.Contains(checked.Diagnostics[0].Range.Start.File+checked.Diagnostics[1].Range.Start.File, "user.skel") ||
		!strings.Contains(checked.Diagnostics[0].Range.Start.File+checked.Diagnostics[1].Range.Start.File, "order.skel") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunSkelcCheckPreservesDomainFileRestrictions(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", "domain demo.user\ndata User {}\n")

	result := Run([]string{"check", "--skel-in", dir})
	checked := decodeCheckResult(t, result)
	if result.ExitCode != ExitCodeUnsatisfied || !checkDiagnosticsContain(checked, diagnostic.CodeDomainFileContent, "can only contain domain declaration and @desc") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func decodeCheckResult(t *testing.T, result Result) *output.CheckResult {
	t.Helper()
	checked := new(output.CheckResult)
	if err := json.Unmarshal([]byte(result.Stdout), checked); err != nil {
		t.Fatalf("decode check result: %v\n%s", err, result.Stdout)
	}
	return checked
}

// checkDiagnosticsContain reports whether a diagnostic carries the stable
// public code and a message containing substring.
func checkDiagnosticsContain(result *output.CheckResult, code string, substring string) bool {
	for _, item := range result.Diagnostics {
		if item.Code == code && strings.Contains(item.Message, substring) {
			return true
		}
	}
	return false
}

// checkDiagnosticsContainCode reports whether any diagnostic carries code.
func checkDiagnosticsContainCode(result *output.CheckResult, code string) bool {
	for _, item := range result.Diagnostics {
		if item.Code == code {
			return true
		}
	}
	return false
}
