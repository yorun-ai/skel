package descriptor

import (
	"fmt"
	"strconv"

	"go.yorun.ai/skel/descriptor"
)

func renderScalarLiteral(scalar descriptor.Scalar) string {
	switch scalar {
	case descriptor.ScalarString:
		return "descriptor.ScalarString"
	case descriptor.ScalarBoolean:
		return "descriptor.ScalarBoolean"
	case descriptor.ScalarInt:
		return "descriptor.ScalarInt"
	case descriptor.ScalarFloat:
		return "descriptor.ScalarFloat"
	case descriptor.ScalarDecimal:
		return "descriptor.ScalarDecimal"
	case descriptor.ScalarJSON:
		return "descriptor.ScalarJSON"
	case descriptor.ScalarUUID:
		return "descriptor.ScalarUUID"
	case descriptor.ScalarTimestamp:
		return "descriptor.ScalarTimestamp"
	case descriptor.ScalarDuration:
		return "descriptor.ScalarDuration"
	case descriptor.ScalarLocalDate:
		return "descriptor.ScalarLocalDate"
	case descriptor.ScalarLocalTime:
		return "descriptor.ScalarLocalTime"
	case descriptor.ScalarLocalDateTime:
		return "descriptor.ScalarLocalDateTime"
	case descriptor.ScalarBinary:
		return "descriptor.ScalarBinary"
	default:
		return "descriptor.Scalar(" + quote(string(scalar)) + ")"
	}
}

func renderAuthModeLiteral(mode descriptor.AuthMode) (string, error) {
	switch mode {
	case "required":
		return "descriptor.AuthModeRequired", nil
	case "optional":
		return "descriptor.AuthModeOptional", nil
	case "anonymous":
		return "descriptor.AuthModeAnonymous", nil
	case "off":
		return "descriptor.AuthModeOff", nil
	case "inherit":
		return "descriptor.AuthModeInherit", nil
	default:
		return "", fmt.Errorf("unsupported auth mode %q", mode)
	}
}

func renderPermissionRequireModeLiteral(mode descriptor.PermissionRequireMode) (string, error) {
	switch mode {
	case descriptor.PermissionRequireModeCode:
		return "descriptor.PermissionRequireModeCode", nil
	case descriptor.PermissionRequireModeCheck:
		return "descriptor.PermissionRequireModeCheck", nil
	case descriptor.PermissionRequireModeAll:
		return "descriptor.PermissionRequireModeAll", nil
	case descriptor.PermissionRequireModeAny:
		return "descriptor.PermissionRequireModeAny", nil
	default:
		return "", fmt.Errorf("unsupported permission require mode %q", mode)
	}
}

func renderConfigLifecycleLiteral(lifecycle descriptor.ConfigLifecycle) (string, error) {
	switch lifecycle {
	case descriptor.ConfigLifecycleEternal:
		return "descriptor.ConfigLifecycleEternal", nil
	case descriptor.ConfigLifecycleInstant:
		return "descriptor.ConfigLifecycleInstant", nil
	default:
		return "", fmt.Errorf("unsupported config lifecycle %q", lifecycle)
	}
}

func quote(value string) string {
	return strconv.Quote(value)
}
