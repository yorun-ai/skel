// Package diagnostic defines skelc's public structured diagnostic contract.
package diagnostic

import "go.yorun.ai/skelc/model"

const (
	CodeAuthLegacy         = "auth.legacy-marker"
	CodeApiAuthMissing     = "service.missing-auth"
	CodeWebAuthMissing     = "web.missing-auth"
	CodeServiceModifier    = "service.legacy-modifier"
	CodeServiceClientRules = "service.legacy-client-rules"
	CodeSyntaxUnexpected   = "syntax.unexpected-token"
	CodeSyntaxEOF          = "syntax.unexpected-eof"
	CodeSyntaxFinalize     = "syntax.invalid-declaration"
	CodeSemanticValidation = "semantic.validation"
	CodeSemanticDuplicate  = "semantic.duplicate"
	CodeSemanticNaming     = "semantic.naming"
	CodeSemanticReference  = "semantic.reference"
	CodeSemanticWarning    = "semantic.warning"
	CodeImportMissing      = "import.missing"
	CodeImportCycle        = "import.cycle"
	CodeDomainMissing      = "domain.missing"
	CodeDomainMismatch     = "domain.mismatch"
	CodeDomainFileContent  = "domain.file-content"
	CodeDomainDecorator    = "domain.decorator-location"
	CodeLoaderDirectory    = "loader.ignored-directory"
	CodeLoaderHiddenFile   = "loader.ignored-hidden-file"
	CodeLoaderUnsupported  = "loader.ignored-file"
)

// Severity identifies the impact of a diagnostic.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// SourceRange identifies a source span using one-based Skel positions.
type SourceRange struct {
	Start model.Position `json:"start"`
	End   model.Position `json:"end"`
}

// RelatedInformation points to another source range relevant to a diagnostic.
type RelatedInformation struct {
	Range   SourceRange `json:"range"`
	Message string      `json:"message"`
}

// Suggestion describes an optional edit that can address a diagnostic.
type Suggestion struct {
	Message     string `json:"message"`
	Replacement string `json:"replacement"`
	Replace     bool   `json:"replace,omitempty"`
}

// Diagnostic is one structured compiler diagnostic.
type Diagnostic struct {
	Code       string               `json:"code"`
	Severity   Severity             `json:"severity"`
	Position   model.Position       `json:"-"`
	Range      SourceRange          `json:"range"`
	Message    string               `json:"message"`
	Related    []RelatedInformation `json:"related,omitempty"`
	Suggestion *Suggestion          `json:"suggestion,omitempty"`
}

func (d Diagnostic) Error() string {
	if d.Position.Line <= 0 {
		return d.Message
	}
	return d.Position.String() + " " + d.Message
}

// Diagnostics is an ordered set of independent diagnostics.
type Diagnostics []Diagnostic

func (d Diagnostics) Error() string {
	if len(d) == 0 {
		return ""
	}
	return d[0].Error()
}

// Errors returns all diagnostics whose severity is not warning.
func (d Diagnostics) Errors() []error {
	result := make([]error, 0, len(d))
	for index := range d {
		if d[index].Severity != SeverityWarning {
			result = append(result, d[index])
		}
	}
	return result
}

// HasErrors reports whether the set contains a non-warning diagnostic.
func (d Diagnostics) HasErrors() bool {
	for _, item := range d {
		if item.Severity != SeverityWarning {
			return true
		}
	}
	return false
}

// Failures returns the non-warning diagnostics.
func (d Diagnostics) Failures() Diagnostics {
	result := Diagnostics{}
	for _, item := range d {
		if item.Severity != SeverityWarning {
			result = append(result, item)
		}
	}
	return result
}

// DiagnosticEntries returns a copy suitable for generic CLI error handling.
func (d Diagnostics) DiagnosticEntries() Diagnostics {
	return append(Diagnostics{}, d...)
}

var _ error = Diagnostic{}
var _ error = Diagnostics{}
var _ interface{ Errors() []error } = Diagnostics{}
