package nameutil

import (
	"strings"
	"unicode"
)

func MaxLength(strs []string) int {
	maxLen := 0
	for _, str := range strs {
		maxLen = max(maxLen, len(str))
	}
	return maxLen
}

// CutKindSuffix splits a declaration name into its body and optional version
// after removing suffix. Versions use V followed by a positive decimal integer
// without leading zeroes. A failed match returns value unchanged.
func CutKindSuffix(value string, suffix string) (body string, version string, found bool) {
	if body, found := strings.CutSuffix(value, suffix); found {
		return body, "", true
	}
	i := strings.LastIndexByte(value, 'V')
	if i < 0 || i+1 == len(value) || value[i+1] < '1' || value[i+1] > '9' {
		return value, "", false
	}
	for j := i + 2; j < len(value); j++ {
		if value[j] < '0' || value[j] > '9' {
			return value, "", false
		}
	}
	if body, found := strings.CutSuffix(value[:i], suffix); found {
		return body, value[i:], true
	}
	return value, "", false
}

func PaddingSpaces(count int) string {
	return strings.Repeat(" ", count)
}

func ToCamel(value string) string {
	words := splitWords(value)
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(value))
	for _, word := range words {
		runes := []rune(strings.ToLower(word))
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}

func ToLowerCamel(value string) string {
	camel := ToCamel(value)
	if camel == "" {
		return ""
	}

	runes := []rune(camel)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func ToSnake(value string) string {
	return strings.ToLower(ToScreamingSnake(value))
}

func ToScreamingSnake(value string) string {
	words := splitWords(value)
	if len(words) == 0 {
		return ""
	}

	parts := make([]string, 0, len(words))
	for _, word := range words {
		parts = append(parts, strings.ToUpper(word))
	}
	return strings.Join(parts, "_")
}

func IsSnakeCase(value string) bool {
	return checkIdentifierCase(value, unicode.IsLower, func(r rune) bool {
		return unicode.IsLower(r)
	}, true)
}

func IsScreamingSnakeCase(value string) bool {
	return checkIdentifierCase(value, unicode.IsUpper, func(r rune) bool {
		return unicode.IsUpper(r)
	}, true)
}

func IsCamelCase(value string) bool {
	return checkIdentifierCase(value, unicode.IsUpper, func(r rune) bool {
		return unicode.IsLetter(r)
	}, false)
}

func IsLowerCamelCase(value string) bool {
	return checkIdentifierCase(value, unicode.IsLower, func(r rune) bool {
		return unicode.IsLetter(r)
	}, false)
}

func splitWords(value string) []string {
	var words []string
	var current []rune

	flush := func() {
		if len(current) == 0 {
			return
		}
		words = append(words, string(current))
		current = nil
	}

	runes := []rune(value)
	for i, r := range runes {
		if isSeparator(r) {
			flush()
			continue
		}

		if len(current) > 0 && shouldSplitWord(runes, i) {
			flush()
		}

		current = append(current, r)
	}
	flush()

	return words
}

func shouldSplitWord(runes []rune, index int) bool {
	if index <= 0 || index >= len(runes) {
		return false
	}

	curr := runes[index]
	prev := runes[index-1]

	if unicode.IsUpper(curr) {
		if unicode.IsLower(prev) || unicode.IsDigit(prev) {
			return true
		}
		if unicode.IsUpper(prev) && index+1 < len(runes) && unicode.IsLower(runes[index+1]) {
			return true
		}
	}

	if unicode.IsDigit(curr) && unicode.IsLetter(prev) {
		return true
	}

	if unicode.IsLetter(curr) && unicode.IsDigit(prev) {
		return true
	}

	return false
}

func isSeparator(r rune) bool {
	switch r {
	case '_', '-', ' ', '.', '/':
		return true
	default:
		return false
	}
}

func checkIdentifierCase(value string, firstCheck func(rune) bool, letterCheck func(rune) bool, allowUnderscore bool) bool {
	if value == "" {
		return false
	}

	for index, r := range value {
		switch {
		case index == 0:
			if !firstCheck(r) {
				return false
			}
		case r == '_':
			if !allowUnderscore {
				return false
			}
		case unicode.IsDigit(r):
			continue
		case unicode.IsLetter(r):
			if !letterCheck(r) {
				return false
			}
		default:
			return false
		}
	}

	if strings.Contains(value, "__") || strings.HasSuffix(value, "_") {
		return false
	}

	return true
}

// SplitQualified separates a declaration name from its domain or explicit alias.
func SplitQualified(value string) (qualifier, name string, qualified bool) {
	index := strings.LastIndexByte(value, '.')
	if index < 0 {
		return "", value, false
	}
	return value[:index], value[index+1:], true
}
