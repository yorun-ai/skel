package skelc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/cmd/skelc/output"
	"go.yorun.ai/skel/internal/codegen/codegentest"
)

func decodeCommandError(t *testing.T, result Result) *output.Error {
	t.Helper()
	commandError := new(output.Error)
	if err := json.Unmarshal([]byte(result.Stdout), commandError); err != nil {
		t.Fatalf("decode command error: %v\n%s", err, result.Stdout)
	}
	return commandError
}

func TestRunUnknownCommandReturnsJSONError(t *testing.T) {
	result := Run([]string{"unknown"})

	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || result.Stderr != "" ||
		commandError.Code != output.ErrorCodeInvalidArgument ||
		!strings.Contains(commandError.Message, "No help topic for 'unknown'") {
		t.Fatalf("unexpected command error: %+v", result)
	}
}

func assertCommandErrorMessage(t *testing.T, result Result, message string) {
	t.Helper()
	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || result.Stderr != "" ||
		commandError.Code != output.ErrorCodeInvalidArgument || commandError.Message != message {
		t.Fatalf("unexpected command error: %+v", result)
	}
}

func assertGenerationResult(t *testing.T, result Result) {
	t.Helper()
	generated := new(output.GenerationResult)
	if err := json.Unmarshal([]byte(result.Stdout), generated); err != nil {
		t.Fatalf("decode generation result: %v\n%s", err, result.Stdout)
	}
	if result.ExitCode != ExitCodeSuccess || !generated.Generated || result.Stderr != "" {
		t.Fatalf("unexpected generation result: %+v", result)
	}
}

func writeCLIFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// newGenFixture returns a source directory holding a minimal valid domain and a
// sibling output directory, which is the setup almost every gen test needs.
func newGenFixture(t *testing.T) (skelIn string, out string) {
	t.Helper()
	skelIn = t.TempDir()
	out = filepath.Join(t.TempDir(), "skeled")
	writeCLIFile(t, filepath.Join(skelIn, "domain.skel"), "domain demo.user")
	return skelIn, out
}

func assertFileContains(t *testing.T, path string, expected ...string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(content)
	for _, item := range expected {
		if !strings.Contains(text, item) {
			t.Fatalf("expected %s to contain %q:\n%s", path, item, text)
		}
	}
}

// assertFileGoSourceContains asserts a generated Go file contains every
// expected fragment as a Go token sequence, so gofmt line breaks and column
// alignment do not decide the outcome of the assertion.
func assertFileGoSourceContains(t *testing.T, path string, expected ...string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, item := range expected {
		codegentest.AssertGoSourceContains(t, string(content), item)
	}
}
func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be missing, err=%v", path, err)
	}
}
