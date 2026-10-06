package typescript

import (
	"fmt"
	"strings"
)

// ValidatePackageName checks npm package names used in generated metadata.
// Dependency version specifiers are not package names.
func validatePackageName(name string) error {
	invalid := func() error { return fmt.Errorf("invalid npm package name %q", name) }
	if name == "" || len(name) > 214 || name != strings.ToLower(name) || name == "node_modules" || name == "favicon.ico" {
		return invalid()
	}
	pkg := name
	if strings.HasPrefix(name, "@") {
		scope, rest, ok := strings.Cut(name[1:], "/")
		if !ok || scope == "" || !npmNameCharacters(scope, true) {
			return invalid()
		}
		pkg = rest
	} else if strings.HasPrefix(name, "_") || strings.HasPrefix(name, "-") {
		return invalid()
	}
	if pkg == "" || strings.HasPrefix(pkg, ".") || !npmNameCharacters(pkg, false) {
		return invalid()
	}
	return nil
}

func npmNameCharacters(part string, scope bool) bool {
	for _, ch := range part {
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_' || ch == '.' {
			continue
		}
		if scope && strings.ContainsRune("~!'()*", ch) {
			continue
		}
		return false
	}
	return true
}
