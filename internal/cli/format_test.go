package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/internal/command"
)

func TestRunSkelcStrictFormatPreservesLegacySource(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "order.skel")
	original := "domain demo.order\nservice OrderService{method ping{}}\n"
	writeCLIFile(t, entry, original)
	result := Run([]string{"--strict", "format", "--skel-in", entry})
	if result.ExitCode != ExitCodeError || decodeCommandError(t, result).Code != command.ErrorCodeCompilationFailed {
		t.Fatalf("expected strict format failure: %+v", result)
	}
	contents, err := os.ReadFile(entry)
	if err != nil || string(contents) != original {
		t.Fatalf("strict format changed source: %v", err)
	}
}

func TestRunSkelcFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "domain.skel")
	writeCLIFile(t, path, "  domain demo.user  \n")

	result := Run([]string{"format", "--skel-in", dir})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected result: %+v", result)
	}
	var output _FormatResult
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil || !output.Changed || len(output.Files) != 1 {
		t.Fatalf("unexpected format result: %+v, err=%v", output, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "domain demo.user\n" {
		t.Fatalf("unexpected formatted content: %q", content)
	}
}

func TestRunSkelcFormatChecksAllFilesBeforeWriting(t *testing.T) {
	dir := t.TempDir()
	domainPath := filepath.Join(dir, "domain.skel")
	writeCLIFile(t, domainPath, "  domain demo.user\n")
	writeCLIFile(t, filepath.Join(dir, "invalid.skel"), "domain demo.user\ndata Invalid { value: }\n")

	result := Run([]string{"format", "--skel-in", dir})
	commandError := decodeCommandError(t, result)
	if result.ExitCode != ExitCodeError || commandError.Code != command.ErrorCodeCompilationFailed {
		t.Fatalf("expected format compilation failure: %+v", result)
	}
	content, err := os.ReadFile(domainPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "  domain demo.user\n" {
		t.Fatalf("valid file changed before validation completed: %q", content)
	}
}

func TestRunSkelcFormatPreservesFileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "domain.skel")
	writeCLIFile(t, path, "  domain demo.user  \n")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}

	result := Run([]string{"format", "--skel-in", dir})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected result: %+v", result)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected formatted file mode: %o", info.Mode().Perm())
	}
}

func TestRunSkelcFormatPreservesInputSymlink(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.skel")
	linkPath := filepath.Join(dir, "domain.skel")
	writeCLIFile(t, targetPath, "  domain demo.user  \n")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Skipf("create symlink: %v", err)
	}

	result := Run([]string{"format", "--skel-in", linkPath})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected result: %+v", result)
	}
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("format replaced the input symlink")
	}
	assertFileExact(t, targetPath, "domain demo.user\n")
}

func TestRunSkelcFormatCheckReportsFilesWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "domain.skel")
	writeCLIFile(t, path, "  domain demo.user  \n")

	result := Run([]string{"format", "--check", "--skel-in", dir})
	if result.ExitCode != ExitCodeUnsatisfied {
		t.Fatalf("expected format check failure: %+v", result)
	}
	var output _FormatResult
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil || !output.Changed || len(output.Files) != 1 || output.Files[0] != path {
		t.Fatalf("unexpected format check output: %+v, err=%v", output, err)
	}
	assertFileExact(t, path, "  domain demo.user  \n")
}

func TestRunSkelcFormatJSONReportsWrittenFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "domain.skel")
	writeCLIFile(t, path, "  domain demo.user  \n")

	result := Run([]string{"format", "--skel-in", dir})
	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected format result: %+v", result)
	}
	var output _FormatResult
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		t.Fatalf("decode format result: %v, stdout=%q", err, result.Stdout)
	}
	if !output.Changed || len(output.Files) != 1 || output.Files[0] != path {
		t.Fatalf("unexpected format result: %+v", output)
	}
	assertFileExact(t, path, "domain demo.user\n")
}

func TestRunSkelcFormatCheckSucceedsWhenFilesAreFormatted(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "domain.skel"), "domain demo.user\n")

	result := Run([]string{"format", "--check", "--skel-in", dir})
	if result.ExitCode != ExitCodeSuccess || result.Stderr != "" {
		t.Fatalf("unexpected format check result: %+v", result)
	}
	var output _FormatResult
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		t.Fatal(err)
	}
	if output.Changed || len(output.Files) != 0 {
		t.Fatalf("unexpected format result: %+v", output)
	}
}

func assertFileExact(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expected {
		t.Fatalf("unexpected content for %s: %q", path, content)
	}
}
