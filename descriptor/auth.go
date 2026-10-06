package descriptor

// AuthMode describes a runtime authentication requirement.
type AuthMode string

const (
	// AuthModeInherit uses the enclosing service authentication mode on methods.
	AuthModeInherit AuthMode = "inherit"
	// AuthModeRequired requires valid credentials.
	AuthModeRequired AuthMode = "required"
	// AuthModeOptional permits anonymous calls and authenticates supplied credentials.
	AuthModeOptional AuthMode = "optional"
	// AuthModeAnonymous permits only anonymous calls.
	AuthModeAnonymous AuthMode = "anonymous"
	// AuthModeOff disables authentication for a web entry point.
	AuthModeOff AuthMode = "off"
)

// IsValid reports whether mode is a canonical authentication mode.
// Consumers enforce restrictions for each declaration kind.
func (mode AuthMode) IsValid() bool {
	switch mode {
	case AuthModeInherit, AuthModeRequired, AuthModeOptional, AuthModeAnonymous, AuthModeOff:
		return true
	default:
		return false
	}
}
