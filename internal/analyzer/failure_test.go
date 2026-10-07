package analyzer

import (
	"errors"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFailurefCarriesStructuredPosition(t *testing.T) {
	err := NewFailuref("%s %s has an invalid type", lexer.Position{Filename: "/workspace/user.skel", Line: 4, Column: 9}, "field")

	var failure *Failure
	require.ErrorAs(t, err, &failure)
	assert.Equal(t, CodeValidation, failure.Code)
	assert.Equal(t, "/workspace/user.skel", failure.Position.File)
	assert.Equal(t, 4, failure.Position.Line)
	assert.Equal(t, 9, failure.Position.Column)
	assert.Equal(t, "/workspace/user.skel:4:9 field has an invalid type", failure.Message)
}

func TestNewFailureWithCauseWrapsCause(t *testing.T) {
	cause := errors.New("cause")
	failure := NewFailureWithCause(cause, "operation failed")

	require.ErrorIs(t, failure, cause)
}

func TestNewFailurefRendersPositionArgumentIntoMessage(t *testing.T) {
	failure := NewFailuref("%s invalid field", lexer.Position{Filename: "user.skel", Line: 2, Column: 3})

	assert.Equal(t, CodeValidation, failure.Code)
	assert.Equal(t, "user.skel", failure.Position.File)
	assert.Equal(t, 2, failure.Position.Line)
	assert.Equal(t, 3, failure.Position.Column)
	assert.Equal(t, "user.skel:2:3 invalid field", failure.Message)
}
