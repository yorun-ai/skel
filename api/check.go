package api

import (
	"context"

	internalapi "go.yorun.ai/skel/internal/api"
)

// CheckOption selects filesystem or frozen sources; imports stay unresolved.
type CheckOption = internalapi.CheckOption

// CheckResult includes errors and warnings found in the source. Invalid source
// sets Valid to false; an error return indicates invalid options or loading failure.
type CheckResult = internalapi.CheckResult

// Check validates source without loading imported domains.
func Check(option CheckOption) (CheckResult, error) {
	return internalapi.Check(option)
}

// CheckContext is Check with cancellation support.
func CheckContext(ctx context.Context, option CheckOption) (CheckResult, error) {
	return internalapi.CheckContext(ctx, option)
}
