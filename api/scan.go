package api

import (
	"context"

	internalapi "go.yorun.ai/skel/internal/api"
)

// ScanOption configures source inspection without resolving imported domains.
type ScanOption = internalapi.ScanOption

// ScanImport identifies a direct import declaration and its source position.
// Repeated declarations are retained; Alias is empty when no alias was written.
type ScanImport = internalapi.ScanImport

// ScanImportsResult contains direct source imports and non-fatal diagnostics.
type ScanImportsResult = internalapi.ScanImportsResult

// ScanImports validates the target input without loading imported domains.
// It does not collect transitive imports or require dependency mappings.
// Invalid input returns an error, as with Parse.
func ScanImports(option ScanOption) (ScanImportsResult, error) {
	return internalapi.ScanImports(option)
}

// ScanImportsContext is ScanImports with cancellation support.
func ScanImportsContext(ctx context.Context, option ScanOption) (ScanImportsResult, error) {
	return internalapi.ScanImportsContext(ctx, option)
}
