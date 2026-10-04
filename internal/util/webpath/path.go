package webpath

import (
	"fmt"
	"strings"
	"unicode"
)

// Validate checks a literal absolute web mount path without normalizing it.
func Validate(value string) error {
	if !strings.HasPrefix(value, "/") {
		return fmt.Errorf("must start with /")
	}
	if strings.ContainsAny(value, "?#\\:*{}%\"<>") || strings.ContainsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) {
		return fmt.Errorf("must be a literal path without query, fragment, escapes, or route parameters")
	}
	if strings.Contains(value, "//") {
		return fmt.Errorf("must not contain empty path segments")
	}
	for segment := range strings.SplitSeq(value, "/") {
		if segment == "." || segment == ".." {
			return fmt.Errorf("must not contain . or .. path segments")
		}
	}
	return nil
}
