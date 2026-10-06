package skelc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/cmd/skelc/output"
	codegenoutput "go.yorun.ai/skel/internal/codegen/output"
)

func TestRunSkelcStrictGenerationPreservesOutputs(t *testing.T) {
	for _, kind := range []string{"go", "go-module", "ts", "skel"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			entry := filepath.Join(root, "order.skel")
			out := filepath.Join(root, "out")
			pubOut := filepath.Join(root, "pub")
			writeCLIFile(t, entry, "domain demo.order\npub service LegacyService { noauth method ping {} }\n")
			args := []string{"gen", kind, "--skel-in", entry}
			switch kind {
			case "go":
				args = append(args, "--go-out", out)
			case "go-module":
				args = append(args, "--go-out", out, "--go-module", "example.com/order", "--go-pub-out", pubOut, "--go-pub-module", "example.com/orderpub")
			case "ts":
				args = append(args, "--api", "--ts-out", out)
			case "skel":
				args = append(args, "--pub", "--skel-out", out)
			}
			if result := Run(args); result.ExitCode != ExitCodeSuccess {
				t.Fatalf("compatible generation failed: %+v", result)
			}
			before := map[string]string{}
			err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					return nil
				}
				data, err := os.ReadFile(path)
				before[path] = string(data)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			result := Run(append([]string{"--strict"}, args...))
			failure := decodeCommandError(t, result)
			if result.ExitCode != ExitCodeError || failure.Code != output.ErrorCodeCompilationFailed || !strings.Contains(result.Stderr, `"severity":"error"`) {
				t.Fatalf("expected strict generation failure: %+v", result)
			}
			for path, want := range before {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != want {
					t.Fatalf("strict failure changed %s: %v", path, err)
				}
			}
		})
	}
}

func TestRunSkelcGenGo(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})
	assertGenerationResult(t, result)
}

func TestRunSkelcGenGoWritesWarningsToDefaultJSONLLogs(t *testing.T) {
	dir, goOut := newGenFixture(t)
	writeCLIFile(t, dir+"/.hidden.skel", `domain demo.user`)

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})
	generated := new(output.GenerationResult)
	if err := json.Unmarshal([]byte(result.Stdout), generated); err != nil {
		t.Fatal(err)
	}
	var warning struct {
		Level    string `json:"level"`
		Code     string `json:"code"`
		Severity string `json:"severity"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(result.Stderr)), &warning); err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != ExitCodeSuccess || !generated.Generated ||
		warning.Level != "warn" || warning.Code != "loader.ignored-hidden-file" || warning.Severity != "warning" {
		t.Fatalf("unexpected generation result: %+v", result)
	}
}

func TestRunSkelcGenGoSupportsTextLogs(t *testing.T) {
	dir, goOut := newGenFixture(t)
	writeCLIFile(t, dir+"/.hidden.skel", `domain demo.user`)

	result := Run([]string{"--log-format", "text", "gen", "go", "--skel-in", dir, "--go-out", goOut})
	if result.ExitCode != ExitCodeSuccess || !strings.HasPrefix(result.Stderr, "[W] ") ||
		!strings.Contains(result.Stderr, ".hidden.skel ignored (HIDDEN_FILE)") {
		t.Fatalf("unexpected text generation log: %+v", result)
	}
}

func TestRunSkelcGenGoClassifiesOutputFailures(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), `domain demo.user`)
	blockedParent := filepath.Join(t.TempDir(), "blocked")
	writeCLIFile(t, blockedParent, "not a directory")

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", filepath.Join(blockedParent, "skeled")})
	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || commandError.Code != output.ErrorCodeCommandFailed {
		t.Fatalf("expected generation output failure: %+v", result)
	}
}

func TestRunSkelcGenGoOmitsCloneCode(t *testing.T) {
	dir, goOut := newGenFixture(t)
	writeCLIFile(t, filepath.Join(dir, "types.skel"), `domain demo.user

data User {
    name: string
    roles: list<string>
}

data Page<TItem> {
    items: list<TItem>
}

data Users {
    page: Page<User>
}

service UserService {
    method listUsers {
        output Page<User>
    }
}
`)

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}

	// Generated data and services no longer carry clone code: the Vine runtime
	// clones in-process Rpc values from their Go types.
	for _, path := range []string{filepath.Join(goOut, "data.go"), filepath.Join(goOut, "service.go")} {
		generated, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range []string{"Clone", "cloneTItem"} {
			if strings.Contains(string(generated), fragment) {
				t.Fatalf("%s must not contain %q:\n%s", path, fragment, generated)
			}
		}
	}
	assertFileContains(t, filepath.Join(goOut, "data.go"), "type User struct {", "type Page[TItem any] struct {")
}

func TestRunSkelcGenGoPreservesUnmanagedOutput(t *testing.T) {
	dir, goOut := newGenFixture(t)
	writeCLIFile(t, filepath.Join(goOut, ".hidden"), "hidden")
	writeCLIFile(t, filepath.Join(goOut, ".hidden-dir", "old.go"), "old")
	writeCLIFile(t, filepath.Join(goOut, "old.go"), "old")

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	assertFileContains(t, filepath.Join(goOut, ".hidden"), "hidden")
	assertFileContains(t, filepath.Join(goOut, ".hidden-dir", "old.go"), "old")
	assertFileContains(t, filepath.Join(goOut, "old.go"), "old")
	assertFileContains(t, filepath.Join(goOut, "doc.go"), "package skeled")
	assertFileContains(t, filepath.Join(goOut, "doc.go"), codegenoutput.GeneratedFileMarker)
}

func TestRunSkelcGenGoDoesNotChangeOutputWhenCompilationFails(t *testing.T) {
	dir := t.TempDir()
	goOut := filepath.Join(t.TempDir(), "skeled")
	writeCLIFile(t, filepath.Join(dir, "user.skel"), `data User { id: string }`)
	writeCLIFile(t, filepath.Join(goOut, "old.go"), "old")

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})

	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || commandError.Code != output.ErrorCodeCompilationFailed {
		t.Fatalf("expected compilation failure: %+v", result)
	}
	assertFileContains(t, filepath.Join(goOut, "old.go"), "old")
}

func TestRunSkelcGenGoModule(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go-module", "--skel-in", dir, "--go-out", goOut, "--go-module-prefix", "github.com/acme/skel"})
	assertGenerationResult(t, result)
	assertFileContains(t, filepath.Join(goOut, "go.mod"), "go.yorun.ai/vine v0.27.0")
}

func TestRunSkelcGenGoModuleWithGoVineVersion(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go-module", "--skel-in", dir, "--go-out", goOut, "--go-module-prefix", "github.com/acme/skel", "--go-vine-version", "v1.2.3"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	assertFileContains(t, filepath.Join(goOut, "go.mod"), "go.yorun.ai/vine v1.2.3")
}

func TestRunSkelcGenGoModuleRejectsLowGoVineVersion(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go-module", "--skel-in", dir, "--go-out", goOut, "--go-module-prefix", "github.com/acme/skel", "--go-vine-version", "v0.8.0"})

	assertCommandErrorMessage(t, result, "go-vine-version v0.8.0 is lower than minimum v0.27.0")
}

func TestRunSkelcGenGoModuleRejectsGoVineVersionWithoutVPrefix(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go-module", "--skel-in", dir, "--go-out", goOut, "--go-module-prefix", "github.com/acme/skel", "--go-vine-version", "1.2.3"})

	assertCommandErrorMessage(t, result, "go-vine-version 1.2.3 must be v-prefixed semantic version")
}

func TestRunSkelcGenGoModuleAcceptsPubFlag(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go-module", "--pub", "--skel-in", dir, "--go-out", goOut, "--go-module-prefix", "github.com/acme/skel"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
}

func TestRunSkelcGenGoRendersDescriptor(t *testing.T) {
	dir, goOut := newGenFixture(t)
	writeCLIFile(t, dir+"/types.skel", `domain demo.user

pub data User {
    id: string
}
`)

	result := Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	original := ldModuleVersion
	t.Cleanup(func() { ldModuleVersion = original })
	ldModuleVersion = "v9.9.9"
	result = Run([]string{"gen", "go", "--skel-in", dir, "--go-out", goOut})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	assertFileGoSourceContains(t, filepath.Join(goOut, "descriptor.go"),
		`Name: "demo.user"`,
		`CompilerVersion: "v9.9.9"`)
}

func TestRunSkelcGenGoRejectsModuleFlags(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go", "--go-module", "example.com/invalid", "--skel-in", dir, "--go-out", goOut})

	assertCommandErrorMessage(t, result, "flag provided but not defined: -go-module")
}

func TestRunSkelcGenGoModuleRejectsMissingModulePrefix(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{"gen", "go-module", "--skel-in", dir, "--go-out", goOut})

	assertCommandErrorMessage(t, result, "missing flag go-module or go-module-prefix")
}

func TestRunSkelcGenGoModuleUsesFlagNameForSharedValidationError(t *testing.T) {
	dir, goOut := newGenFixture(t)

	result := Run([]string{
		"gen", "go-module", "--skel-in", dir, "--go-out", goOut,
		"--go-module-prefix", "github.com/acme/skel/",
	})

	assertCommandErrorMessage(t, result, "flag go-module-prefix must not end with /")
}

func TestRunSkelcGenGoModuleWithSkelImportAndGoImport(t *testing.T) {
	root := t.TempDir()
	userDir := filepath.Join(root, "user")
	bookerDir := filepath.Join(root, "booker")
	goOut := filepath.Join(root, "booker_gen")
	writeCLIFile(t, filepath.Join(userDir, "domain.skel"), `domain user`)
	writeCLIFile(t, filepath.Join(userDir, "types.skel"), `domain user

pub data UserSummary {
    userId: int
}
`)
	writeCLIFile(t, filepath.Join(bookerDir, "domain.skel"), `domain booker

import user as account
`)
	writeCLIFile(t, filepath.Join(bookerDir, "types.skel"), `domain booker

pub data LoanRecord {
    borrower: account.UserSummary
}
`)

	result := Run([]string{
		"gen", "go-module",
		"--skel-in", bookerDir,
		"--go-out", goOut,
		"--skel-import", "user=" + userDir,
		"--go-import", "user=go.yorun.ai/app/vine/demo/userpub",
		"--go-module", "go.yorun.ai/app/vine/demo/bookerpub",
	})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	dataContent, err := os.ReadFile(filepath.Join(goOut, "data.go"))
	if err != nil {
		t.Fatalf("read generated data.go: %v", err)
	}
	if !strings.Contains(string(dataContent), `account "go.yorun.ai/app/vine/demo/userpub"`) {
		t.Fatalf("expected user import in generated data.go: %s", string(dataContent))
	}
	if !strings.Contains(string(dataContent), "Borrower account.UserSummary") {
		t.Fatalf("expected external user type in generated data.go: %s", string(dataContent))
	}
	goModContent, err := os.ReadFile(filepath.Join(goOut, "go.mod"))
	if err != nil {
		t.Fatalf("read generated go.mod: %v", err)
	}
	if !strings.Contains(string(goModContent), "module go.yorun.ai/app/vine/demo/bookerpub") {
		t.Fatalf("unexpected go.mod content: %s", string(goModContent))
	}
}

func TestRunSkelcGenGoModuleWithDefaultSkelImportUsesPubPackageName(t *testing.T) {
	root := t.TempDir()
	userDir := filepath.Join(root, "user")
	bookerDir := filepath.Join(root, "booker")
	goOut := filepath.Join(root, "booker_gen")
	writeCLIFile(t, filepath.Join(userDir, "domain.skel"), `domain user`)
	writeCLIFile(t, filepath.Join(userDir, "types.skel"), `domain user

pub data UserSummary {
    userId: int
}
`)
	writeCLIFile(t, filepath.Join(bookerDir, "domain.skel"), `domain booker

import user
`)
	writeCLIFile(t, filepath.Join(bookerDir, "types.skel"), `domain booker

pub data LoanRecord {
    borrower: user.UserSummary
}
`)

	result := Run([]string{
		"gen", "go-module",
		"--skel-in", bookerDir,
		"--go-out", goOut,
		"--skel-import", "user=" + userDir,
		"--go-module-prefix", "go.yorun.ai/app/vine/demo",
	})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	dataContent, err := os.ReadFile(filepath.Join(goOut, "data.go"))
	if err != nil {
		t.Fatalf("read generated data.go: %v", err)
	}
	if !strings.Contains(string(dataContent), `"go.yorun.ai/app/vine/demo/userpub"`) {
		t.Fatalf("expected user import in generated data.go: %s", string(dataContent))
	}
	if strings.Contains(string(dataContent), `user "go.yorun.ai/app/vine/demo/userpub"`) {
		t.Fatalf("expected user import without alias in generated data.go: %s", string(dataContent))
	}
	if !strings.Contains(string(dataContent), "Borrower userpub.UserSummary") {
		t.Fatalf("expected external userpub type in generated data.go: %s", string(dataContent))
	}
}
