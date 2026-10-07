package analyzer

import (
	"context"
	"slices"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/schema"
)

// MaxDiagnosticsPerDomain bounds validation work and prevents a badly broken
// source from flooding editor and command-line output.
const MaxDiagnosticsPerDomain = 50

// Semantic diagnostic codes are assigned where analyzer failures originate so
// downstream integrations never need to classify human-readable messages.
const (
	DiagnosticCodeValidation = diagnostic.CodeSemanticValidation
	DiagnosticCodeDuplicate  = diagnostic.CodeSemanticDuplicate
	DiagnosticCodeNaming     = diagnostic.CodeSemanticNaming
	DiagnosticCodeReference  = diagnostic.CodeSemanticReference
)

type _DiagnosticReporter struct {
	ctx    context.Context
	errors []error
	seen   map[string]bool
}

func newDiagnosticReporter() *_DiagnosticReporter {
	return &_DiagnosticReporter{seen: map[string]bool{}}
}

func (r *_DiagnosticReporter) check(condition bool, message string, args ...any) bool {
	return r.checkCode(DiagnosticCodeValidation, condition, message, args...)
}

func (r *_DiagnosticReporter) checkDuplicate(condition bool, message string, args ...any) bool {
	return r.checkCode(DiagnosticCodeDuplicate, condition, message, args...)
}

func (r *_DiagnosticReporter) checkReference(condition bool, message string, args ...any) bool {
	return r.checkCode(DiagnosticCodeReference, condition, message, args...)
}

func (r *_DiagnosticReporter) checkCode(code string, condition bool, message string, args ...any) bool {
	if r.cancelled() {
		return false
	}
	if condition {
		return true
	}
	r.report(newDiagnosticFailure(code, message, args...))
	return false
}

func (r *_DiagnosticReporter) checkNot(condition bool, message string, args ...any) bool {
	return r.check(!condition, message, args...)
}

func (r *_DiagnosticReporter) checkNotDuplicate(condition bool, message string, args ...any) bool {
	return r.checkDuplicate(!condition, message, args...)
}

func (r *_DiagnosticReporter) reportf(message string, args ...any) {
	r.report(newDiagnosticFailure(DiagnosticCodeValidation, message, args...))
}

func (r *_DiagnosticReporter) reportDuplicatef(message string, args ...any) {
	failure := newDiagnosticFailure(DiagnosticCodeDuplicate, message, args...)
	positions := diagnosticArgumentPositions(args)
	if len(positions) > 1 {
		failure.Related = []RelatedLocation{{Position: positions[len(positions)-1], Message: "first declaration"}}
	}
	r.report(failure)
}

func (r *_DiagnosticReporter) reportReferencef(message string, args ...any) {
	r.report(newDiagnosticFailure(DiagnosticCodeReference, message, args...))
}

func (r *_DiagnosticReporter) reportNamingf(replacement string, message string, args ...any) {
	failure := newDiagnosticFailure(DiagnosticCodeNaming, message, args...)
	failure.Suggestion = &Suggestion{
		Message:     "replace with " + replacement,
		Replacement: replacement,
		Replace:     true,
	}
	r.report(failure)
}

func newDiagnosticFailure(code string, message string, args ...any) *Failure {
	failure := NewFailuref(message, args...)
	failure.Code = code
	return failure
}

func diagnosticArgumentPositions(args []any) []schema.Position {
	positions := []schema.Position{}
	for _, argument := range args {
		switch position := argument.(type) {
		case schema.Position:
			positions = append(positions, position)
		case lexer.Position:
			positions = append(positions, schema.Position{File: position.Filename, Line: position.Line, Column: position.Column})
		}
	}
	return positions
}

func (r *_DiagnosticReporter) report(err error) {
	if r.cancelled() || err == nil || len(r.errors) >= MaxDiagnosticsPerDomain {
		return
	}
	position, _ := Position(err)
	key := position.String() + "\x00" + err.Error()
	if r.seen[key] {
		return
	}
	r.seen[key] = true
	r.errors = append(r.errors, err)
}

func (r *_DiagnosticReporter) full() bool {
	return r.cancelled() || len(r.errors) >= MaxDiagnosticsPerDomain
}

func (r *_DiagnosticReporter) result() []error {
	result := append([]error{}, r.errors...)
	slices.SortFunc(result, func(left, right error) int {
		leftPosition, _ := Position(left)
		rightPosition, _ := Position(right)
		if compared := strings.Compare(leftPosition.File, rightPosition.File); compared != 0 {
			return compared
		}
		if leftPosition.Line != rightPosition.Line {
			return leftPosition.Line - rightPosition.Line
		}
		if leftPosition.Column != rightPosition.Column {
			return leftPosition.Column - rightPosition.Column
		}
		return strings.Compare(left.Error(), right.Error())
	})
	return result
}

func (r *_DiagnosticReporter) cancelled() bool { return r.ctx != nil && r.ctx.Err() != nil }
