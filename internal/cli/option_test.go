package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/optionvalidation"
)

func TestNormalizeCheckOption(t *testing.T) {
	compilerOption := compiler.Option{SkelIn: "./demo"}

	if err := normalizeCompilerOption(&compilerOption); err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs("./demo")
	if err != nil {
		t.Fatal(err)
	}
	if compilerOption.SkelIn != want {
		t.Fatalf("skel-in = %q, want %q", compilerOption.SkelIn, want)
	}
}

func TestNormalizeCheckOptionRequiresInput(t *testing.T) {
	compilerOption := compiler.Option{}
	expectOptionError(t, normalizeCompilerOption(&compilerOption), "missing flag skel-in")
}

func TestFormatGenerationErrorUsesTypedValidationContract(t *testing.T) {
	err := fmt.Errorf("normalize options: %w", optionvalidation.NewValidationError(
		optionvalidation.FieldGoModule,
		optionvalidation.RuleRequiresModule,
		"API message",
	))
	formatted := formatGenerationError(err)
	if formatted.Error() != "flag go-module requires go-module output" {
		t.Fatalf("unexpected CLI validation message: %v", formatted)
	}
}

func expectOptionError(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %v", expected, err)
	}
}
