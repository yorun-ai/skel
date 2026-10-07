package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/schema"
)

func TestPublicDiagnosticCodesRemainStable(t *testing.T) {
	expected := map[string]string{
		"syntax unexpected":   CodeSyntaxUnexpected,
		"syntax eof":          CodeSyntaxEOF,
		"syntax finalize":     CodeSyntaxFinalize,
		"semantic validation": CodeSemanticValidation,
		"semantic duplicate":  CodeSemanticDuplicate,
		"semantic naming":     CodeSemanticNaming,
		"semantic reference":  CodeSemanticReference,
		"semantic warning":    CodeSemanticWarning,
		"import missing":      CodeImportMissing,
		"import cycle":        CodeImportCycle,
		"domain missing":      CodeDomainMissing,
		"domain mismatch":     CodeDomainMismatch,
		"domain file content": CodeDomainFileContent,
		"domain decorator":    CodeDomainDecorator,
		"loader directory":    CodeLoaderDirectory,
		"loader hidden file":  CodeLoaderHiddenFile,
		"loader unsupported":  CodeLoaderUnsupported,
	}
	assert.Equal(t, map[string]string{
		"syntax unexpected":   "syntax.unexpected-token",
		"syntax eof":          "syntax.unexpected-eof",
		"syntax finalize":     "syntax.invalid-declaration",
		"semantic validation": "semantic.validation",
		"semantic duplicate":  "semantic.duplicate",
		"semantic naming":     "semantic.naming",
		"semantic reference":  "semantic.reference",
		"semantic warning":    "semantic.warning",
		"import missing":      "import.missing",
		"import cycle":        "import.cycle",
		"domain missing":      "domain.missing",
		"domain mismatch":     "domain.mismatch",
		"domain file content": "domain.file-content",
		"domain decorator":    "domain.decorator-location",
		"loader directory":    "loader.ignored-directory",
		"loader hidden file":  "loader.ignored-hidden-file",
		"loader unsupported":  "loader.ignored-file",
	}, expected)
}

func TestDiagnosticsExposeOnlyFailuresAsErrors(t *testing.T) {
	diagnostics := Diagnostics{
		{Severity: SeverityWarning, Message: "warning"},
		{Severity: SeverityError, Position: schema.Position{File: "demo.skel", Line: 2, Column: 3}, Message: "failure"},
	}

	require.True(t, diagnostics.HasErrors())
	require.Len(t, diagnostics.Errors(), 1)
	assert.Equal(t, "demo.skel:2:3 failure", diagnostics.Errors()[0].Error())
	assert.Equal(t, diagnostics[1:], diagnostics.Failures())
}
