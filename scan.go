package skelc

import (
	"cmp"
	"context"
	"slices"

	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/command"
)

// ScanOption configures source inspection without resolving imported domains.
type ScanOption struct {
	// SkelIn is a Skel source file or domain directory.
	SkelIn string
	// Sources optionally supplies a complete in-memory snapshot; see Input.Sources.
	Sources map[string][]byte
	// Strict rejects legacy declarations accepted with migration warnings.
	Strict bool
}

// ScanImport identifies a direct import declaration and its source position.
// Repeated declarations are retained; Alias is empty when no alias was written.
type ScanImport = command.ScanImport

// ScanImportsResult contains direct source imports and non-fatal diagnostics.
type ScanImportsResult struct {
	// Imports is sorted by file, line and column and is empty rather than nil
	// when the source declares no imports.
	Imports     []ScanImport
	Diagnostics []diagnostic.Diagnostic
}

// ScanImports validates the target input without loading imported domains.
// It does not collect transitive imports or require dependency mappings.
// Invalid input returns an error, as with Parse.
func ScanImports(option ScanOption) (ScanImportsResult, error) {
	return ScanImportsContext(context.Background(), option)
}

// ScanImportsContext is ScanImports with cancellation support.
func ScanImportsContext(ctx context.Context, option ScanOption) (ScanImportsResult, error) {
	input, err := normalizeInput(Input{SkelIn: option.SkelIn, Sources: option.Sources, Strict: option.Strict})
	if err != nil {
		return ScanImportsResult{}, err
	}
	result, err := compileInput(ctx, Input{Sources: option.Sources}, input, true)
	if err != nil {
		return ScanImportsResult{}, err
	}
	imports := make([]ScanImport, 0, len(result.Imports))
	for _, imported := range result.Imports {
		imports = append(imports, ScanImport{Domain: imported.Name, Alias: imported.Alias,
			File: imported.Pos.File, Line: imported.Pos.Line, Column: imported.Pos.Column})
	}
	slices.SortFunc(imports, func(a, b ScanImport) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column))
	})
	return ScanImportsResult{Imports: imports, Diagnostics: result.Diagnostics}, nil
}
