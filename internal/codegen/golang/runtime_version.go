package golang

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"go.yorun.ai/skel/internal/optionvalidation"
)

// ResolveOption validates compiler metadata and resolves only the selected runtime.
func ResolveOption(option Option) (ResolvedOption, error) {
	option.VineVersion = strings.TrimSpace(option.VineVersion)
	option.VrpcVersion = strings.TrimSpace(option.VrpcVersion)
	if option.ApiOnly {
		if option.VineVersion != "" {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoVineVersion, optionvalidation.RuleInvalid, "go-vine-version is not used by api clients")
		}
		version, err := resolveVrpcVersion(option.VrpcVersion)
		if err != nil {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoVrpcVersion, optionvalidation.RuleInvalid, err.Error())
		}
		option.VrpcVersion = version
	} else {
		if option.VrpcVersion != "" {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoVrpcVersion, optionvalidation.RuleInvalid, "go-vrpc-version requires api")
		}
		version, err := resolveVineVersion(option.VineVersion)
		if err != nil {
			return ResolvedOption{}, optionvalidation.NewValidationError(optionvalidation.FieldGoVineVersion, optionvalidation.RuleInvalid, err.Error())
		}
		option.VineVersion = version
	}
	option.CompilerVersion = strings.TrimSpace(option.CompilerVersion)
	if !option.ApiOnly {
		if err := validateCompilerVersion(option.CompilerVersion); err != nil {
			return ResolvedOption{}, err
		}
	}
	return ResolvedOption{option: option}, nil
}

const minimumCompilerVersion = "v0.17.1"
const developmentCompilerVersion = "v0.0.0-dev"

func validateCompilerVersion(version string) error {
	field := optionvalidation.FieldGoCompilerVersion
	if version == "" {
		return optionvalidation.NewValidationError(field, optionvalidation.RuleRequired, "CompilerVersion is required for backend Go output; specify the actual skelc version")
	}
	if version == developmentCompilerVersion {
		return nil
	}
	parsed, err := semver.StrictNewVersion(strings.TrimPrefix(version, "v"))
	if !strings.HasPrefix(version, "v") || err != nil {
		return optionvalidation.NewValidationError(field, optionvalidation.RuleInvalid, fmt.Sprintf("invalid CompilerVersion %q: must be a v-prefixed semantic version", version))
	}
	if parsed.Compare(semver.MustParse(minimumCompilerVersion)) < 0 {
		return optionvalidation.NewValidationError(field, optionvalidation.RuleInvalid, fmt.Sprintf("CompilerVersion %s is lower than minimum %s", version, minimumCompilerVersion))
	}
	return nil
}
