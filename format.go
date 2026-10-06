package skelc

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/formatter"
	"go.yorun.ai/skelc/internal/loader"
	"go.yorun.ai/skelc/internal/parser"
)

// ErrFormatCompilation identifies source loading or validation failures.
var ErrFormatCompilation = errors.New("format source compilation failed")

type _FormatCompilationError struct{ cause error }

func (e *_FormatCompilationError) Error() string   { return e.cause.Error() }
func (e *_FormatCompilationError) Unwrap() []error { return []error{ErrFormatCompilation, e.cause} }

// FormatSource returns canonical Skel syntax without writing any files.
// Invalid source returns an error; formatting preserves declaration order.
func FormatSource(content []byte) ([]byte, error) {
	if err := parser.ValidateSource("<source>", content); err != nil {
		return nil, err
	}
	formatted, err := formatter.Source(content)
	if err != nil {
		return nil, err
	}
	if err := parser.ValidateSource("<source>", formatted); err != nil {
		return nil, err
	}
	return formatted, nil
}

// FormatOption selects sources for a read-only formatting plan.
type FormatOption = ScanOption

// FormattedFile is one source requiring a change. Bytes are owned by the result.
type FormattedFile struct {
	Path     string
	Original []byte
	Content  []byte
}

// FormatResult contains changed files after all inputs have been validated.
// No files are written; callers own any subsequent write transaction.
type FormatResult struct {
	Changed     bool
	Files       []FormattedFile
	Diagnostics []diagnostic.Diagnostic
}

// FormatFiles checks formatting and returns replacements without writing files.
func FormatFiles(option FormatOption) (FormatResult, error) {
	return FormatFilesContext(context.Background(), option)
}

// FormatFilesContext is FormatFiles with cancellation support.
func FormatFilesContext(ctx context.Context, option FormatOption) (FormatResult, error) {
	input := Input{SkelIn: option.SkelIn, Sources: option.Sources, Strict: option.Strict}
	normalized, err := normalizeInput(input)
	if err != nil {
		return FormatResult{}, err
	}
	provider, err := inputProvider(input)
	if err != nil {
		return FormatResult{}, err
	}
	loaded, err := loader.LoadFrom(ctx, provider, normalized.SkelIn)
	if err != nil {
		return FormatResult{}, &_FormatCompilationError{cause: err}
	}
	result := FormatResult{Files: []FormattedFile{}, Diagnostics: compiler.LoaderWarningDiagnostics(loaded.Warnings)}
	if option.Strict {
		checked, err := compiler.CheckLoaded(ctx, loaded, normalized)
		if err != nil {
			return FormatResult{}, &_FormatCompilationError{cause: err}
		}
		if checked.Diagnostics.HasErrors() {
			return FormatResult{}, &_FormatCompilationError{cause: checked.Diagnostics}
		}
		result.Diagnostics = checked.Diagnostics
	}
	for _, file := range loaded.Files {
		if err := ctx.Err(); err != nil {
			return FormatResult{}, err
		}
		if err := parser.ValidateSource(file.FilePath, file.Content); err != nil {
			return FormatResult{}, &_FormatCompilationError{cause: err}
		}
		formatted, err := formatter.Source(file.Content)
		if err != nil {
			return FormatResult{}, fmt.Errorf("format %s: %w", file.FilePath, err)
		}
		if err := parser.ValidateSource(file.FilePath, formatted); err != nil {
			return FormatResult{}, err
		}
		if !bytes.Equal(file.Content, formatted) {
			result.Files = append(result.Files, FormattedFile{Path: file.FilePath, Original: bytes.Clone(file.Content), Content: formatted})
		}
	}
	result.Changed = len(result.Files) != 0
	return result, nil
}
