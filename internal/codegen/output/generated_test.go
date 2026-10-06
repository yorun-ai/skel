package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarkGeneratedFileUsesFormatCompatibleMarkers(t *testing.T) {
	lineMarkedFiles := []string{"data.go", "service.ts", "domain.skel", "go.mod"}
	for _, path := range lineMarkedFiles {
		t.Run(path, func(t *testing.T) {
			content, err := MarkGeneratedFile(path, "content\n")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(content, "// "+GeneratedFileMarker+"\n\n") {
				t.Fatalf("missing line marker in %q", content)
			}
			if !HasGeneratedFileMarker([]byte(content)) {
				t.Fatal("expected generated file marker to be detected")
			}
		})
	}

	t.Run("package.json", func(t *testing.T) {
		content, err := MarkGeneratedFile("package.json", "{\n  \"name\": \"demo\"\n}\n")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(content, "{\n  \"_generated\": \""+GeneratedFileMarker+"\",") {
			t.Fatalf("missing JSON marker in %q", content)
		}
		if !json.Valid([]byte(content)) {
			t.Fatalf("marked package.json is invalid: %s", content)
		}
		if !HasGeneratedFileMarker([]byte(content)) {
			t.Fatal("expected generated JSON marker to be detected")
		}
	})
}

func TestMarkGeneratedFileRejectsUnsupportedFormat(t *testing.T) {
	if _, err := MarkGeneratedFile("generated.bin", "content"); err == nil {
		t.Fatal("expected unsupported generated output format error")
	}
}

func TestMarkGeneratedFileDoesNotDuplicateMarker(t *testing.T) {
	marked, err := MarkGeneratedFile("data.go", "package demo\n")
	if err != nil {
		t.Fatal(err)
	}
	again, err := MarkGeneratedFile("data.go", marked)
	if err != nil {
		t.Fatal(err)
	}
	if again != marked {
		t.Fatalf("marker changed on second application:\n%s", again)
	}
}
