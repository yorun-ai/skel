package skel

import (
	"context"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/compiler"
)

// CheckOption selects filesystem or frozen sources; imports stay unresolved.
type CheckOption = ScanOption

// CheckResult includes errors and warnings found in the source. Invalid source
// sets Valid to false; an error return indicates invalid options or loading failure.
type CheckResult struct {
	Valid       bool
	Diagnostics []diagnostic.Diagnostic
}

// Check validates source without loading imported domains.
func Check(option CheckOption) (CheckResult, error) {
	return CheckContext(context.Background(), option)
}

// CheckContext is Check with cancellation support.
func CheckContext(ctx context.Context, option CheckOption) (CheckResult, error) {
	input := Input{SkelIn: option.SkelIn, Sources: option.Sources, Strict: option.Strict}
	normalized, err := normalizeInput(input)
	if err != nil {
		return CheckResult{}, err
	}
	provider, err := inputProvider(input)
	if err != nil {
		return CheckResult{}, err
	}
	result, err := compiler.CheckFrom(ctx, provider, normalized)
	if err != nil {
		return CheckResult{}, err
	}
	return CheckResult{Valid: !result.Diagnostics.HasErrors(), Diagnostics: result.Diagnostics}, nil
}
