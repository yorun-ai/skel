package skelc

import (
	"errors"
	"fmt"
	"path/filepath"

	"go.yorun.ai/skel/internal/cmd/skelc/output"
	codegenoutput "go.yorun.ai/skel/internal/codegen/output"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/optionvalidation"
)

type _OptionValidationKey struct {
	field optionvalidation.Field
	rule  optionvalidation.Rule
}

var optionValidationMessages = map[_OptionValidationKey]string{
	{optionvalidation.FieldApiPrune, optionvalidation.RuleRequiresApi}:                 "flag prune requires api",
	{optionvalidation.FieldApiType, optionvalidation.RuleRequiresApi}:                  "flag name requires api",
	{optionvalidation.FieldApiType, optionvalidation.RuleRequiresPrune}:                "flag name requires prune",
	{optionvalidation.FieldApiPrune, optionvalidation.RuleRequired}:                    "flag prune requires at least one actor or name",
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
	code := output.ErrorCodeCompilationFailed
	var validationError *optionvalidation.ValidationError
	switch {
	case errors.As(err, &validationError):
		code = output.ErrorCodeInvalidArgument
	case errors.Is(err, codegenoutput.ErrOutputOperation):
		code = output.ErrorCodeCommandFailed
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
