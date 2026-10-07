package compiler

import (
	"errors"
	"path/filepath"
	"strings"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/analyzer"
	textsource "go.yorun.ai/skel/internal/source"
	"go.yorun.ai/skel/schema"
)

const (
	DiagnosticCodeSyntaxUnexpected   = diagnostic.CodeSyntaxUnexpected
	DiagnosticCodeSyntaxEOF          = diagnostic.CodeSyntaxEOF
	DiagnosticCodeSyntaxFinalize     = diagnostic.CodeSyntaxFinalize
	DiagnosticCodeSemanticValidation = diagnostic.CodeSemanticValidation
	DiagnosticCodeSemanticDuplicate  = diagnostic.CodeSemanticDuplicate
	DiagnosticCodeSemanticNaming     = diagnostic.CodeSemanticNaming
	DiagnosticCodeSemanticReference  = diagnostic.CodeSemanticReference
	DiagnosticCodeSemanticWarning    = diagnostic.CodeSemanticWarning
	DiagnosticCodeImportMissing      = diagnostic.CodeImportMissing
	DiagnosticCodeImportCycle        = diagnostic.CodeImportCycle
	DiagnosticCodeDomainMissing      = diagnostic.CodeDomainMissing
	DiagnosticCodeDomainMismatch     = diagnostic.CodeDomainMismatch
	DiagnosticCodeDomainFileContent  = diagnostic.CodeDomainFileContent
	DiagnosticCodeDomainDecorator    = diagnostic.CodeDomainDecorator
	DiagnosticCodeLoaderDirectory    = diagnostic.CodeLoaderDirectory
	DiagnosticCodeLoaderHiddenFile   = diagnostic.CodeLoaderHiddenFile
	DiagnosticCodeLoaderUnsupported  = diagnostic.CodeLoaderUnsupported
)

type DiagnosticSeverity = diagnostic.Severity

const (
	DiagnosticSeverityError   = diagnostic.SeverityError
	DiagnosticSeverityWarning = diagnostic.SeverityWarning
)

type SourceRange = diagnostic.SourceRange
type DiagnosticRelatedInformation = diagnostic.RelatedInformation
type DiagnosticSuggestion = diagnostic.Suggestion
type Diagnostic = diagnostic.Diagnostic
type Diagnostics = diagnostic.Diagnostics

func diagnosticFromError(path, fallbackCode string, err error) Diagnostic {
	diagnostic := Diagnostic{
		Code: fallbackCode, Severity: DiagnosticSeverityError,
		Position: schema.Position{File: path, Line: 1, Column: 1}, Message: err.Error(),
	}
	if sourcePosition, ok := analyzer.Position(err); ok {
		diagnostic.Position = sourcePosition
		diagnostic.Message = strings.TrimPrefix(err.Error(), sourcePosition.String()+" ")
	}
	var failure *analyzer.Failure
	if errors.As(err, &failure) {
		if failure.Code != "" && failure.Code != analyzer.CodeValidation {
			diagnostic.Code = failure.Code
		}
		for _, related := range failure.Related {
			diagnostic.Related = append(diagnostic.Related, DiagnosticRelatedInformation{
				Range: SourceRange{Start: related.Position, End: related.Position}, Message: related.Message,
			})
		}
		if failure.Suggestion != nil {
			diagnostic.Suggestion = &DiagnosticSuggestion{
				Message: failure.Suggestion.Message, Replacement: failure.Suggestion.Replacement, Replace: failure.Suggestion.Replace,
			}
		}
	}
	var missingImport *analyzer.MissingImportError
	if errors.As(err, &missingImport) {
		diagnostic.Code = DiagnosticCodeImportMissing
	}
	diagnostic.Range = SourceRange{Start: diagnostic.Position, End: diagnostic.Position}
	return diagnostic
}

func completeDiagnostic(diagnostic *Diagnostic, contents map[string]*textsource.Document) {
	if diagnostic.Severity == "" {
		diagnostic.Severity = DiagnosticSeverityError
	}
	if diagnostic.Range.Start.Line <= 0 || diagnostic.Range.End.Line <= 0 || diagnostic.Range.End == diagnostic.Range.Start {
		diagnostic.Range = sourceRangeAtDocument(diagnostic.Position, contents[filepath.Clean(diagnostic.Position.File)])
	}
	for index := range diagnostic.Related {
		position := diagnostic.Related[index].Range.Start
		if diagnostic.Related[index].Range.End.Line <= 0 || diagnostic.Related[index].Range.End == position {
			diagnostic.Related[index].Range = sourceRangeAtDocument(position, contents[filepath.Clean(position.File)])
		}
	}
}

func compareDiagnostics(left, right Diagnostic) int {
	if compared := strings.Compare(left.Position.File, right.Position.File); compared != 0 {
		return compared
	}
	if left.Position.Line != right.Position.Line {
		return left.Position.Line - right.Position.Line
	}
	if left.Position.Column != right.Position.Column {
		return left.Position.Column - right.Position.Column
	}
	return strings.Compare(left.Message, right.Message)
}
