package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
)

// describedUserDomain is the domain fixture shared by the compiler tests.
const describedUserDomain = "@desc(\"User domain\")\ndomain demo.user\n"

// mustMkdirAll creates every directory, failing the test on error.
func mustMkdirAll(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}
}

func parseDomain(t *testing.T, files map[string]string) *model.Domain {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		if name != loader.DomainFileName && !strings.HasPrefix(strings.TrimLeft(content, "\n\t "), "domain ") {
			content = "domain demo.user\n" + content
		}
		writeFile(t, filepath.Join(dir, name), content)
	}

	result, err := Compile(Option{SkelIn: dir})
	if err != nil {
		t.Fatalf("compile domain: %v", err)
	}
	return result.Domain
}

func findDataByName(t *testing.T, domain *model.Domain, name string) *model.Data {
	t.Helper()

	for _, data := range domain.Data() {
		if data.Name == name {
			return data
		}
	}
	t.Fatalf("data %s not found", name)
	return nil
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func identForTest(value string) *grammar.Identifier {
	return &grammar.Identifier{Value: value}
}

func domainContentForTest(name string, description string) *grammar.DomainContent {
	parts := strings.Split(name, ".")
	identParts := make([]*grammar.Identifier, 0, len(parts))
	for _, part := range parts {
		identParts = append(identParts, identForTest(part))
	}

	return &grammar.DomainContent{
		Description: description,
		Name: &grammar.QualifiedName{
			Parts: identParts,
		},
	}
}
func expectErrorContains(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %v", expected, err)
	}
}
