package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/internal/analyzer"
)

func TestDiagnosticFromErrorUsesStructuredMetadata(t *testing.T) {
	failure := analyzer.NewFailuref("wording does not identify this diagnostic")
	failure.Code = analyzer.DiagnosticCodeNaming
	failure.Suggestion = &analyzer.Suggestion{
		Message: "replace with UserProfile", Replacement: "UserProfile", Replace: true,
	}

	diagnostic := diagnosticFromError("user.skel", DiagnosticCodeSemanticValidation, failure)

	assert.Equal(t, DiagnosticCodeSemanticNaming, diagnostic.Code)
	require.NotNil(t, diagnostic.Suggestion)
	assert.Equal(t, "replace with UserProfile", diagnostic.Suggestion.Message)
	assert.Equal(t, "UserProfile", diagnostic.Suggestion.Replacement)
	assert.True(t, diagnostic.Suggestion.Replace)
}

func TestDiagnosticFromErrorDoesNotInferMetadataFromMessage(t *testing.T) {
	failure := analyzer.NewFailuref("duplicated unknown value not found; expected=WrongName")

	diagnostic := diagnosticFromError("user.skel", DiagnosticCodeSemanticValidation, failure)

	assert.Equal(t, DiagnosticCodeSemanticValidation, diagnostic.Code)
	assert.Nil(t, diagnostic.Suggestion)
}
