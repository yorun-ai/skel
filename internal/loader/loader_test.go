package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "domain.skel"), "domain demo.user\n")
	writeFile(t, filepath.Join(dir, "a.skel"), "domain demo.user\ndata User { id: string }\n")
	writeFile(t, filepath.Join(dir, ".hidden.skel"), "domain hidden\n")
	writeFile(t, filepath.Join(dir, "README.md"), "ignored\n")
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	result := mustLoad(t, dir)

	if !result.IsDir {
		t.Fatal("expected directory result")
	}
	if len(result.Files) != 2 {
		t.Fatalf("unexpected source files: %+v", result.Files)
	}
	if filepath.Base(result.Files[0].FilePath) != "a.skel" || filepath.Base(result.Files[1].FilePath) != "domain.skel" {
		t.Fatalf("source files are not sorted: %+v", result.Files)
	}
	if len(result.Warnings) != 3 {
		t.Fatalf("unexpected warnings: %+v", result.Warnings)
	}
	if result.Warnings[0].Code != WarningCodeHiddenFile || result.Warnings[1].Code != WarningCodeUnsupported || result.Warnings[2].Code != WarningCodeDirectory {
		t.Fatalf("unexpected warning codes: %+v", result.Warnings)
	}
}

func TestLoadReturnsErrorWhenDomainFileMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "service.skel"), "domain demo.user\n")

	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "domain.skel not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsNonSkelFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "README.md")
	writeFile(t, filePath, "ignored\n")

	_, err := Load(filePath)
	if err == nil || !strings.Contains(err.Error(), "is not a .skel file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustLoad(t *testing.T, path string) Result {
	t.Helper()
	result, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
