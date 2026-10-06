package api

import (
	"context"

	internalapi "go.yorun.ai/skel/internal/api"
)

// Input identifies the primary Skel source and any imported domains.
type Input = internalapi.Input

// ParseResult contains a validated semantic model and non-fatal diagnostics.
type ParseResult = internalapi.ParseResult

// Parse loads and validates a Skel contract for use by custom generators and
// tools. All direct and transitive imported domains must be declared in
// Input.SkelImports.
func Parse(input Input) (ParseResult, error) {
	return internalapi.Parse(input)
}

// ParseContext is Parse with cancellation support, including frozen inputs.
func ParseContext(ctx context.Context, input Input) (ParseResult, error) {
	return internalapi.ParseContext(ctx, input)
}
