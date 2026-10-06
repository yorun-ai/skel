package golang_test

import (
	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratorAlwaysRendersGoDocFile(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")

	pkg := newModelDomainForTest(t, codegentest.DomainModel("demo.user"))

	if err := generateFixture(pkg, golang.Option{Out: goOutDir}); err != nil {
		t.Fatal(err)
	}

	goDocContent, err := os.ReadFile(filepath.Join(goOutDir, "doc.go"))
	if err != nil {
		t.Fatalf("read go doc file: %v", err)
	}
	if !strings.Contains(string(goDocContent), "// Package skeled") {
		t.Fatalf("expected fallback go doc comment, got:\n%s", string(goDocContent))
	}
	if !strings.Contains(string(goDocContent), "// This package is fully managed by skelc.") ||
		!strings.Contains(string(goDocContent), "// unmanaged files.") {
		t.Fatalf("expected generated package ownership warning, got:\n%s", string(goDocContent))
	}
	if !strings.Contains(string(goDocContent), "package skeled") {
		t.Fatalf("expected go doc package declaration, got:\n%s", string(goDocContent))
	}
	if strings.Contains(string(goDocContent), "\n\npackage skeled") {
		t.Fatalf("did not expect blank line between fallback doc comment and package declaration, got:\n%s", string(goDocContent))
	}

	assertFileMissing(t, filepath.Join(goOutDir, "go.mod"))
	assertFileMissing(t, filepath.Join(goOutDir, "go.sum"))
}
