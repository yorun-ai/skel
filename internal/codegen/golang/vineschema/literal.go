package vineschema

import (
	"fmt"
	"strconv"

	"go.yorun.ai/skel/internal/model"
)

func renderScalarLiteral(scalar _Scalar) string {
	switch scalar {
	case scalarString:
		return "skel.ScalarString"
	case scalarBool:
		return "skel.ScalarBool"
	case scalarInt:
		return "skel.ScalarInt"
	case scalarLong:
		return "skel.ScalarLong"
	case scalarFloat:
		return "skel.ScalarFloat"
	case scalarDouble:
		return "skel.ScalarDouble"
	case scalarDecimal:
		return "skel.ScalarDecimal"
	case scalarJson:
		return "skel.ScalarJson"
	case scalarUuid:
		return "skel.ScalarUuid"
	case scalarTimestamp:
		return "skel.ScalarTimestamp"
	case scalarDuration:
		return "skel.ScalarDuration"
	case scalarLocalDate:
		return "skel.ScalarLocalDate"
	case scalarLocalTime:
		return "skel.ScalarLocalTime"
	case scalarLocalDateTime:
		return "skel.ScalarLocalDateTime"
	case scalarBinary:
		return "skel.ScalarBinary"
	default:
		return "skel.Scalar(" + quote(string(scalar)) + ")"
	}
}

func actorVia(name string) _ActorVia {
	switch model.ActorViaKind(name) {
	case model.ActorViaClient:
		return actorViaClient
	case model.ActorViaAgent:
		return actorViaAgent
	case model.ActorViaOpenAPI:
		return actorViaOpenAPI
	default:
		return ""
	}
}

func renderActorViaLiteral(method _ActorVia) (string, error) {
	switch method {
	case actorViaClient:
		return "skel.ActorViaClient", nil
	case actorViaAgent:
		return "skel.ActorViaAgent", nil
	case actorViaOpenAPI:
		return "skel.ActorViaOpenAPI", nil
	default:
		return "", fmt.Errorf("unsupported actor via %q", method)
	}
}

func renderAuthModeLiteral(mode _AuthMode) (string, error) {
	switch mode {
	case "required":
		return "skel.AuthModeRequired", nil
	case "optional":
		return "skel.AuthModeOptional", nil
	case "anonymous":
		return "skel.AuthModeAnonymous", nil
	case "off":
		return "skel.AuthModeOff", nil
	case "inherit":
		return "skel.AuthModeInherit", nil
	default:
		return "", fmt.Errorf("unsupported auth mode %q", mode)
	}
}

func renderPermRequireModeLiteral(mode _PermRequireMode) (string, error) {
	switch mode {
	case permRequireModeCode:
		return "skel.PermRequireModeCode", nil
	case permRequireModeCheck:
		return "skel.PermRequireModeCheck", nil
	case permRequireModeAll:
		return "skel.PermRequireModeAll", nil
	case permRequireModeAny:
		return "skel.PermRequireModeAny", nil
	default:
		return "", fmt.Errorf("unsupported permission require mode %q", mode)
	}
}

func quote(value string) string {
	return strconv.Quote(value)
}
