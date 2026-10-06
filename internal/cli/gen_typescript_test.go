package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/command"
)

func TestRunSkelcGenTS(t *testing.T) {
	dir := t.TempDir()
	tsOut := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)

	result := Run([]string{"gen", "ts", "--api", "--skel-in", dir, "--ts-out", tsOut})
	assertGenerationResult(t, result)
	assertFileMissing(t, filepath.Join(tsOut, "package.json"))
}

func TestRunSkelcGenTSWithModule(t *testing.T) {
	dir := t.TempDir()
	tsOut := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)

	result := Run([]string{"gen", "ts", "--api", "--skel-in", dir, "--ts-out", tsOut, "--ts-as-module", "--ts-module", "@acme/skeled-user"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	packageJSONContent, err := os.ReadFile(filepath.Join(tsOut, "package.json"))
	if err != nil {
		t.Fatalf("read generated package.json: %v", err)
	}
	if !strings.Contains(string(packageJSONContent), `"name": "@acme/skeled-user"`) {
		t.Fatalf("expected generated package.json to use ts module: %s", string(packageJSONContent))
	}
}

func TestRunSkelcGenTSUsesFlagNamesForSharedValidationErrors(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "missing module identity",
			args:     []string{"gen", "ts", "--api", "--skel-in", dir, "--ts-out", t.TempDir(), "--ts-as-module"},
			expected: "missing flag ts-module or ts-module-scope",
		},
		{
			name:     "module requires module output",
			args:     []string{"gen", "ts", "--api", "--skel-in", dir, "--ts-out", t.TempDir(), "--ts-module", "@acme/user"},
			expected: "flag ts-module requires ts-as-module",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Run(test.args)
			assertCommandErrorMessage(t, result, test.expected)
		})
	}
}

func TestRunSkelcGenTSApi(t *testing.T) {
	dir := t.TempDir()
	tsOut := t.TempDir()
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)
	writeCLIFile(t, dir+"/actor.skel", `domain demo.user

pub actor ClientActor {
    via client {}
}
`)
	writeCLIFile(t, dir+"/types.skel", `domain demo.user

pub enum UserStatus {
    ACTIVE
}

pub data User {
    status: UserStatus
}

pub data PublicUnused {
    id: int
}

data InternalUser {
    id: int
}
`)
	writeCLIFile(t, dir+"/service.skel", `domain demo.user

api service UserApiService {
    for ClientActor

    method getUser {
        output User
    }
}

pub service InternalService {

    method getUser {
        output User
    }
}
`)

	result := Run([]string{"gen", "ts", "--api", "--skel-in", dir, "--ts-out", tsOut})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	dataContent, err := os.ReadFile(filepath.Join(tsOut, "data.ts"))
	if err != nil {
		t.Fatalf("read generated data.ts: %v", err)
	}
	if !strings.Contains(string(dataContent), "export type PublicUnused = {") {
		t.Fatalf("expected API TypeScript to include unreachable pub data: %s", string(dataContent))
	}
	if strings.Contains(string(dataContent), "InternalUser") {
		t.Fatalf("did not expect non-pub data in generated data.ts: %s", string(dataContent))
	}
	serviceContent, err := os.ReadFile(filepath.Join(tsOut, "service.ts"))
	if err != nil {
		t.Fatalf("read generated service.ts: %v", err)
	}
	if !strings.Contains(string(serviceContent), "createUserApiService") {
		t.Fatalf("expected API service client in generated service.ts: %s", string(serviceContent))
	}
	if strings.Contains(string(serviceContent), "createInternalService") {
		t.Fatalf("did not expect backend service client in generated service.ts: %s", string(serviceContent))
	}
	specContent, err := os.ReadFile(filepath.Join(tsOut, "spec.ts"))
	if err != nil {
		t.Fatalf("read generated spec.ts: %v", err)
	}
	if !strings.Contains(string(specContent), "UserApiServiceSpec") {
		t.Fatalf("expected API service spec in generated spec.ts: %s", string(specContent))
	}
	if strings.Contains(string(specContent), "InternalServiceSpec") {
		t.Fatalf("did not expect backend service spec in generated spec.ts: %s", string(specContent))
	}
}

func TestRunSkelcGenTSRequiresApiAndRejectsPub(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), "domain demo.order")
	for _, test := range []struct {
		name    string
		flags   []string
		code    command.ErrorCode
		message string
	}{
		{name: "missing api", code: command.ErrorCodeCompilationFailed, message: "TypeScript generation requires api"},
		{name: "rejects pub", flags: []string{"--pub"}, code: command.ErrorCodeInvalidArgument, message: "flag provided but not defined: -pub"},
		{name: "rejects api with pub", flags: []string{"--api", "--pub"}, code: command.ErrorCodeInvalidArgument, message: "flag provided but not defined: -pub"},
	} {
		t.Run(test.name, func(t *testing.T) {
			args := append([]string{"gen", "ts", "--skel-in", dir, "--ts-out", t.TempDir()}, test.flags...)
			result := Run(args)
			commandError := decodeCommandError(t, result)
			if result.ExitCode != ExitCodeError || result.Stderr != "" ||
				commandError.Code != test.code || commandError.Message != test.message {
				t.Fatalf("unexpected command error: %+v", result)
			}
		})
	}
	help := Run([]string{"gen", "ts", "--help"})
	if help.ExitCode != ExitCodeSuccess || !strings.Contains(help.Stdout, "--api") || strings.Contains(help.Stdout, "--pub") {
		t.Fatalf("unexpected TS options: %s", help.Stdout)
	}
}

func TestRunSkelcGenTSEmptyPackageReplacesPreviousDeclarations(t *testing.T) {
	for _, asModule := range []bool{false, true} {
		t.Run(fmt.Sprintf("module=%v", asModule), func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(t.TempDir(), "typescript")
			input := filepath.Join(dir, "contracts.skel")
			args := []string{"gen", "ts", "--api", "--skel-in", input, "--ts-out", out}
			if asModule {
				args = append(args, "--ts-as-module", "--ts-module", "@acme/empty")
			}
			writeCLIFile(t, input, "domain demo.empty\n\npub data Visible {\n value: string\n}\n")
			assertGenerationResult(t, Run(args))
			for _, name := range []string{"data.ts", "service.ts", "spec.ts"} {
				if _, err := os.Stat(filepath.Join(out, name)); err != nil {
					t.Fatal(err)
				}
			}
			writeCLIFile(t, filepath.Join(out, "notes.txt"), "user-owned\n")
			// A domain can contain declarations but expose no TypeScript API.
			writeCLIFile(t, input, "domain demo.empty\n\ndata Internal {\n value: string\n}\n")
			assertGenerationResult(t, Run(args))
			index, err := os.ReadFile(filepath.Join(out, "index.ts"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(index), "export {};") || strings.Contains(string(index), "export *") {
				t.Fatalf("unexpected empty entry point: %s", index)
			}
			for _, name := range []string{"data.ts", "service.ts", "spec.ts"} {
				assertFileMissing(t, filepath.Join(out, name))
			}
			if content, err := os.ReadFile(filepath.Join(out, "notes.txt")); err != nil || string(content) != "user-owned\n" {
				t.Fatalf("user file changed: %q, %v", content, err)
			}
			if asModule {
				content, err := os.ReadFile(filepath.Join(out, "package.json"))
				if err != nil {
					t.Fatal(err)
				}
				var pkg struct {
					Name    string `json:"name"`
					Exports map[string]struct {
						Types   string `json:"types"`
						Default string `json:"default"`
					} `json:"exports"`
				}
				if err := json.Unmarshal(content, &pkg); err != nil {
					t.Fatal(err)
				}
				if pkg.Name != "@acme/empty" || pkg.Exports["."].Types != "./index.ts" || pkg.Exports["."].Default != "./index.ts" {
					t.Fatalf("unexpected empty package metadata: %s", content)
				}
			} else {
				assertFileMissing(t, filepath.Join(out, "package.json"))
			}
			assertGenerationResult(t, Run(args))
			repeated, err := os.ReadFile(filepath.Join(out, "index.ts"))
			if err != nil || string(repeated) != string(index) {
				t.Fatalf("empty output is not stable: %q, %v", repeated, err)
			}
		})
	}
}
