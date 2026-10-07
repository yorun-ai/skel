package schema

import "go.yorun.ai/skel/internal/policy"

// AuthModeInherit represents a method that uses its enclosing service's policy.
const AuthModeInherit AuthMode = "inherit"

// NormalizedAuth returns the effective service default.
func (s *Service) NormalizedAuth() AuthMode {
	return AuthMode(policy.NormalizeAuth(string(s.AuthMode), "service"))
}

// NormalizedAuth returns the method policy, retaining inheritance explicitly.
func (m *Method) NormalizedAuth() AuthMode {
	return AuthMode(policy.NormalizeAuth(string(m.AuthMode), "method"))
}

// NormalizedAuth returns the web policy.
func (w *Web) NormalizedAuth() AuthMode {
	return AuthMode(policy.NormalizeAuth(string(w.AuthMode), "web"))
}
