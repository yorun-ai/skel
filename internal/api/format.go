package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"go.yorun.ai/skel/diagnostic"
	internalcompiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/formatter"
	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/parser"
)

// ErrFormatCompilation identifies source loading or validation failures,
// excluding cancellation and deadline expiration.
var ErrFormatCompilation = errors.New("format source compilation failed")

type _FormatCompilationError struct{ cause error }

func (e *_FormatCompilationError) Error() string   { return e.cause.Error() }
func (e *_FormatCompilationError) Unwrap() []error { return []error{ErrFormatCompilation, e.cause} }

func formatCompilationError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return &_FormatCompilationError{cause: err}
}

// FormatSource returns canonical Skel syntax without writing any files.
// Invalid source returns an error; formatting preserves declaration order.
func FormatSource(content []byte) ([]byte, error) {
	return formatter.ValidatedSource(content)
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
		return FormatResult{}, formatCompilationError(err)
	}
	result := FormatResult{Files: []FormattedFile{}, Diagnostics: internalcompiler.LoaderWarningDiagnostics(loaded.Warnings)}
	if option.Strict {
		checked, err := internalcompiler.CheckLoaded(ctx, loaded, normalized)
		if err != nil {
			return FormatResult{}, formatCompilationError(err)
		}
		if checked.Diagnostics.HasErrors() {
			return FormatResult{}, formatCompilationError(checked.Diagnostics)
		}
		result.Diagnostics = checked.Diagnostics
	}
	for _, file := range loaded.Files {
		if err := ctx.Err(); err != nil {
			return FormatResult{}, err
		}
		if err := parser.ValidateSource(file.FilePath, file.Content); err != nil {
			return FormatResult{}, formatCompilationError(err)
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
	if err := ctx.Err(); err != nil {
		return FormatResult{}, err
	}
	result.Changed = len(result.Files) != 0
	return result, nil
}
