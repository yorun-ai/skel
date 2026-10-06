package descriptor

import (
	"fmt"
	"strconv"

	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func renderScalarLiteral(scalar descriptor.Scalar) string {
	switch scalar {
	case descriptor.ScalarString:
		return "descriptor.ScalarString"
	case descriptor.ScalarBool:
		return "descriptor.ScalarBool"
	case descriptor.ScalarInt:
		return "descriptor.ScalarInt"
	case descriptor.ScalarLong:
		return "descriptor.ScalarLong"
	case descriptor.ScalarFloat:
		return "descriptor.ScalarFloat"
	case descriptor.ScalarDouble:
		return "descriptor.ScalarDouble"
	case descriptor.ScalarDecimal:
		return "descriptor.ScalarDecimal"
	case descriptor.ScalarJson:
		return "descriptor.ScalarJson"
	case descriptor.ScalarUuid:
		return "descriptor.ScalarUuid"
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

func actorVia(name string) descriptor.ActorVia {
	switch schema.ActorViaKind(name) {
	case schema.ActorViaClient:
		return descriptor.ActorViaClient
	case schema.ActorViaAgent:
		return descriptor.ActorViaAgent
	case schema.ActorViaOpenAPI:
		return descriptor.ActorViaOpenAPI
	default:
		return ""
	}
}

func renderActorViaLiteral(method descriptor.ActorVia) (string, error) {
	switch method {
	case descriptor.ActorViaClient:
		return "descriptor.ActorViaClient", nil
	case descriptor.ActorViaAgent:
		return "descriptor.ActorViaAgent", nil
	case descriptor.ActorViaOpenAPI:
		return "descriptor.ActorViaOpenAPI", nil
	default:
		return "", fmt.Errorf("unsupported actor via %q", method)
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

func quote(value string) string {
	return strconv.Quote(value)
}
