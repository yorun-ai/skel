package schema

// AuthModeInherit represents a method that uses its enclosing service's policy.
const AuthModeInherit AuthMode = "inherit"

// NormalizedAuth returns the effective service default, including legacy syntax.
func (s *Service) NormalizedAuth() AuthMode {
	mode := normalizeAuth(s.Auth)
	if mode == AuthModeInherit {
		return AuthModeRequired
	}
	return mode
}

// NormalizedAuth returns the method policy, retaining inheritance explicitly.
func (m *Method) NormalizedAuth() AuthMode { return normalizeAuth(m.Auth) }

// NormalizedAuth returns the web policy, where legacy noauth means off.
func (w *Web) NormalizedAuth() AuthMode {
	switch w.Auth {
	case "", AuthModeUnset, AuthModeAuth:
		return AuthModeRequired
	case AuthModeNoAuth:
		return AuthModeOff
	default:
		return w.Auth
	}
}

func normalizeAuth(mode AuthMode) AuthMode {
	switch mode {
	case "", AuthModeUnset:
		return AuthModeInherit
	case AuthModeAuth:
		return AuthModeRequired
	case AuthModeNoAuth:
		return AuthModeOptional
	default:
		return mode
	}
}
