package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"go.yorun.ai/skelc/internal/codegen/output"
	"go.yorun.ai/skelc/internal/command"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/optionvalidation"
)

type _OptionValidationKey struct {
	field optionvalidation.Field
	rule  optionvalidation.Rule
}

var optionValidationMessages = map[_OptionValidationKey]string{
	{optionvalidation.FieldApiPrune, optionvalidation.RuleRequiresApi}:                 "flag prune requires api",
	{optionvalidation.FieldApiType, optionvalidation.RuleRequiresApi}:                  "flag type requires api",
	{optionvalidation.FieldApiType, optionvalidation.RuleRequiresPrune}:                "flag type requires prune",
	{optionvalidation.FieldApiPrune, optionvalidation.RuleRequired}:                    "flag prune requires at least one actor or type",
	{optionvalidation.FieldApiActor, optionvalidation.RuleRequiresApi}:                 "flag actor requires api",
	{optionvalidation.FieldSkelInput, optionvalidation.RuleRequired}:                   "missing flag skel-in",
	{optionvalidation.FieldGoOutput, optionvalidation.RuleRequired}:                    "missing output flag",
	{optionvalidation.FieldGoModuleIdentity, optionvalidation.RuleRequired}:            "missing flag go-module or go-module-prefix",
	{optionvalidation.FieldGoPublicOutput, optionvalidation.RuleRequiresModule}:        "flag go-pub-out requires go-module output",
	{optionvalidation.FieldGoPublicModule, optionvalidation.RuleRequiresModule}:        "flag go-pub-module requires go-module output",
	{optionvalidation.FieldGoModule, optionvalidation.RuleRequiresModule}:              "flag go-module requires go-module output",
	{optionvalidation.FieldGoModulePrefix, optionvalidation.RuleRequiresModule}:        "flag go-module-prefix requires go-module output",
	{optionvalidation.FieldGoPublicModule, optionvalidation.RuleRequiresPublicOutput}:  "flag go-pub-module requires go-pub-out",
	{optionvalidation.FieldGoModulePrefix, optionvalidation.RuleNoTrailingSlash}:       "flag go-module-prefix must not end with /",
	{optionvalidation.FieldGoModule, optionvalidation.RuleNoTrailingSlash}:             "flag go-module must not end with /",
	{optionvalidation.FieldGoPublicModule, optionvalidation.RuleNoTrailingSlash}:       "flag go-pub-module must not end with /",
	{optionvalidation.FieldGoImport, optionvalidation.RuleNoTrailingSlash}:             "flag go-import must not end with /",
	{optionvalidation.FieldTypeScriptOutput, optionvalidation.RuleRequired}:            "missing output flag",
	{optionvalidation.FieldTypeScriptModuleIdentity, optionvalidation.RuleRequired}:    "missing flag ts-module or ts-module-scope",
	{optionvalidation.FieldTypeScriptModule, optionvalidation.RuleRequiresModule}:      "flag ts-module requires ts-as-module",
	{optionvalidation.FieldTypeScriptModuleScope, optionvalidation.RuleRequiresModule}: "flag ts-module-scope requires ts-as-module",
	{optionvalidation.FieldSkeletonOutput, optionvalidation.RuleRequired}:              "missing output flag",
	{optionvalidation.FieldSkeletonPublicOnly, optionvalidation.RuleRequired}:          "flag pub is required for gen skel",
}

func formatGenerationError(err error) error {
	var validationError *optionvalidation.ValidationError
	if !errors.As(err, &validationError) {
		return err
	}
	message := optionValidationMessages[_OptionValidationKey{field: validationError.Field, rule: validationError.Rule}]
	if message == "" {
		return err
	}
	return errors.New(message)
}

func generationCommandFailure(err error) error {
	code := command.ErrorCodeCompilationFailed
	var validationError *optionvalidation.ValidationError
	switch {
	case errors.As(err, &validationError):
		code = command.ErrorCodeInvalidArgument
	case errors.Is(err, output.ErrOutputOperation):
		code = command.ErrorCodeCommandFailed
	}
	return commandFailure(code, formatGenerationError(err))
}

func normalizeCompilerOption(option *compiler.Option) error {
	if option.SkelIn == "" {
		return fmt.Errorf("missing flag skel-in")
	}
	path, err := filepath.Abs(option.SkelIn)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", option.SkelIn, err)
	}
	option.SkelIn = path
	return nil
}
