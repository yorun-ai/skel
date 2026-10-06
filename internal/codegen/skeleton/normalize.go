package skeleton

import (
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/optionvalidation"
	"strings"
)

// NormalizeOption validates and normalizes generation options before source compilation.
func NormalizeOption(option Option) (Option, error) {
	if strings.TrimSpace(option.Out) == "" {
		return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldSkeletonOutput, optionvalidation.RuleRequired, "Skel output is required")
	}
	if !option.PubOnly {
		return Option{}, optionvalidation.NewValidationError(optionvalidation.FieldSkeletonPublicOnly, optionvalidation.RuleRequired, "Skel generation requires PubOnly")
	}
	out, err := common.AbsolutePath(option.Out)
	if err != nil {
		return Option{}, err
	}
	return Option{PubOnly: option.PubOnly, Out: out}, nil
}
