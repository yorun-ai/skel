package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/golang"
	"go.yorun.ai/skel/internal/compiler"
)

func TestGenerateSensitiveTaskInputEndToEnd(t *testing.T) {
	inputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(inputDir, "domain.skel"), []byte("domain demo.task\n"), 0o644); err != nil {
		t.Fatalf("write domain source: %v", err)
	}
	source := `domain demo.task

task RebuildIndexTask {
    trigger atTime {
        @sensitive
        input {
            @sensitive
            token: string
        }
    }
}
`
	if err := os.WriteFile(filepath.Join(inputDir, "task.skel"), []byte(source), 0o644); err != nil {
		t.Fatalf("write task source: %v", err)
	}

	parsed, err := compiler.Compile(compiler.Option{SkelIn: inputDir})
	if err != nil {
		t.Fatalf("parse Skel source: %v", err)
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", parsed.Diagnostics)
	}
	trigger := parsed.Domain.Tasks()[0].Triggers[0]
	if !trigger.ArgumentsSensitive || !trigger.Arguments[0].Sensitive {
		t.Fatalf("sensitive metadata was not preserved by analysis: %+v", trigger)
	}
	if trigger.Hash == "" {
		t.Fatal("expected trigger compatibility hash")
	}

	outputDir := filepath.Join(t.TempDir(), "skeled")
	if err := generateFixture(parsed.Domain, golang.Option{Out: outputDir}); err != nil {
		t.Fatalf("generate Go: %v", err)
	}

	taskContent := readFileForTest(t, filepath.Join(outputDir, "task.go"))
	if !strings.Contains(taskContent, "ArgumentsSensitive: true,") {
		t.Fatalf("expected sensitive trigger input in TriggerSpec, got:\n%s", taskContent)
	}
	if !strings.Contains(taskContent, `json:"token" skel:"sensitive"`) {
		t.Fatalf("expected sensitive task argument tag, got:\n%s", taskContent)
	}

	schemaContent := readFileForTest(t, filepath.Join(outputDir, "schema.go"))
	if !strings.Contains(schemaContent, "ArgumentsSensitive: true,") ||
		!strings.Contains(schemaContent, "Sensitive: true,") {
		t.Fatalf("expected sensitive task metadata in DomainSchema, got:\n%s", schemaContent)
	}
}
