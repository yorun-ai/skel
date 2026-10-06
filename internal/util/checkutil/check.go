package checkutil

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/model"
)

const CodeValidation = "validation"

// Failure is the structured validation error used by semantic diagnostics.
// Error preserves the existing human-readable message, while Position and
// Code let tool integrations avoid parsing that message.
type Failure struct {
	Code       string
	Position   model.Position
	Message    string
	Cause      error
	Related    []RelatedLocation
	Suggestion *Suggestion
}

// RelatedLocation points to another source location relevant to a failure.
type RelatedLocation struct {
	Position model.Position
	Message  string
}

// Suggestion describes an optional structured edit for a failure.
type Suggestion struct {
	Message     string
	Replacement string
	Replace     bool
}

func (f *Failure) Error() string { return f.Message }

func (f *Failure) Unwrap() error { return f.Cause }

// SourcePosition returns the source position associated with the failure.
func (f *Failure) SourcePosition() model.Position { return f.Position }

// NewFailuref constructs a structured validation failure.
func NewFailuref(message string, args ...any) *Failure {
	return &Failure{
		Code: CodeValidation, Position: positionFromArgs(args), Message: fmt.Sprintf(message, args...),
	}
}

// NewFailureWithCause constructs a structured validation failure that wraps a cause.
func NewFailureWithCause(cause error, message string, args ...any) *Failure {
	return &Failure{
		Code: CodeValidation, Position: positionFromArgs(args), Message: message, Cause: cause,
	}
}

// Position returns the structured source position carried by err.
func Position(err error) (model.Position, bool) {
	var positioned interface{ SourcePosition() model.Position }
	if !errors.As(err, &positioned) {
		return model.Position{}, false
	}
	position := positioned.SourcePosition()
	return position, position.Line > 0
}

func positionFromArgs(args []any) model.Position {
	for _, arg := range args {
		switch position := arg.(type) {
		case model.Position:
			return position
		case *model.Position:
			if position != nil {
				return *position
			}
		case lexer.Position:
			return model.Position{File: position.Filename, Line: position.Line, Column: position.Column}
		case *lexer.Position:
			if position != nil {
				return model.Position{File: position.Filename, Line: position.Line, Column: position.Column}
			}
		}
	}
	return model.Position{}
}
