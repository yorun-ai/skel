package api

import (
	"context"

	internalapi "go.yorun.ai/skel/internal/api"
)

// ErrFormatCompilation identifies source loading or validation failures,
// excluding cancellation and deadline expiration.
var ErrFormatCompilation = internalapi.ErrFormatCompilation

// FormatSource returns canonical Skel syntax without writing any files.
// Invalid source returns an error; formatting preserves declaration order.
func FormatSource(content []byte) ([]byte, error) {
	return internalapi.FormatSource(content)
}

// FormatOption selects sources for a read-only formatting plan.
type FormatOption = internalapi.FormatOption

// FormattedFile is one source requiring a change. Bytes are owned by the result.
type FormattedFile = internalapi.FormattedFile

// FormatResult contains changed files after all inputs have been validated.
// No files are written; callers own any subsequent write transaction.
type FormatResult = internalapi.FormatResult

// FormatFiles checks formatting and returns replacements without writing files.
func FormatFiles(option FormatOption) (FormatResult, error) {
	return internalapi.FormatFiles(option)
}

// FormatFilesContext is FormatFiles with cancellation support.
func FormatFilesContext(ctx context.Context, option FormatOption) (FormatResult, error) {
	return internalapi.FormatFilesContext(ctx, option)
}
