package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/schema"
)

func TestDiagnosticsExposeOnlyFailuresAsErrors(t *testing.T) {
	diagnostics := Diagnostics{
		{Severity: SeverityWarning, Message: "warning"},
		{Severity: SeverityError, Position: schema.Position{File: "demo.skel", Line: 2, Column: 3}, Message: "failure"},
	}

	require.True(t, diagnostics.HasErrors())
	assert.Equal(t, "demo.skel:2:3 failure", diagnostics.Error())
	require.Len(t, diagnostics.Errors(), 1)
	assert.Equal(t, "demo.skel:2:3 failure", diagnostics.Errors()[0].Error())
	assert.Equal(t, diagnostics[1:], diagnostics.Failures())
}
