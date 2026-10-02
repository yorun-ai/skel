package analyzer

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/parser/grammar"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

type _CaseType string

const (
	caseTypeSnake          _CaseType = "snake_case"
	caseTypeScreamingSnake _CaseType = "SCREAMING_SNAKE_CASE"
	caseTypeCamel          _CaseType = "CamelCase"
	caseTypeLowerCamel     _CaseType = "lowerCamelCase"
)

var reservedKindSuffixes = []string{"Config", "Event", "Actor", "Service", "Web"}

func checkCase(reporter *_DiagnosticReporter, kindName string, expectedCase _CaseType, ident *grammar.Identifier) bool {
	return checkCaseAdvanced(reporter, kindName, "", "", expectedCase, ident)
}

func checkNotReservedKindSuffix(reporter *_DiagnosticReporter, kindName string, ident *grammar.Identifier) bool {
	valid := true
	for _, suffix := range reservedKindSuffixes {
		if reporter.cancelled() {
			break
		}
		valid = reporter.checkNot(strings.HasSuffix(ident.Value, suffix),
			"%s %s name must not end with %s", ident.Pos, kindName, suffix) && valid
	}
	return valid
}

func checkCaseAdvanced(
	reporter *_DiagnosticReporter,
	kindName string,
	prefix string,
	suffix string,
	expectedCase _CaseType,
	ident *grammar.Identifier,
) bool {
	name := ident.Value
	pos := ident.Pos
	valid := true

	valid = reporter.checkNot(strings.HasPrefix(name, "_"), "%s unexpected leading underscore for %s: %s ", pos, kindName, name) && valid

	expectedFormat := string(expectedCase)
	if prefix != "" {
		expectedFormat = fmt.Sprintf("[%s]%s", prefix, expectedFormat)
	}
	if suffix != "" {
		expectedFormat = fmt.Sprintf("%s[%s]", expectedFormat, suffix)
	}

	valid = reporter.check(prefix == "" || strings.HasPrefix(name, prefix),
		"%s missing prefix: found=%s, expected=%s... (%s -> %s)", pos, name, prefix, kindName, expectedFormat) && valid
	valid = reporter.check(suffix == "" || strings.HasSuffix(ident.Value, suffix),
		"%s missing suffix: found=%s, expected=...%s (%s -> %s)", pos, name, suffix, kindName, expectedFormat) && valid

	body := strings.TrimPrefix(name, prefix)
	body = strings.TrimSuffix(body, suffix)
	valid = reporter.check(body != "", "%s missing body after trimming prefix & suffix: found=%s", pos, name) && valid
	if !matchesCase(body, expectedCase) {
		expectedName := fmt.Sprintf("%s%s%s", prefix, convertCase(body, expectedCase), suffix)
		reporter.reportNamingf(expectedName, "%s incorrect case: found=%s, expected=%s (%s -> %s)", pos, name, expectedName, kindName, expectedFormat)
		valid = false
	}
	return valid
}

func matchesCase(value string, expected _CaseType) bool {
	switch expected {
	case caseTypeSnake:
		return nameutil.IsSnakeCase(value)
	case caseTypeScreamingSnake:
		return nameutil.IsScreamingSnakeCase(value)
	case caseTypeCamel:
		return nameutil.IsCamelCase(value)
	case caseTypeLowerCamel:
		return nameutil.IsLowerCamelCase(value)
	default:
		return false
	}
}

func convertCase(value string, expected _CaseType) string {
	switch expected {
	case caseTypeSnake:
		return nameutil.ToSnake(value)
	case caseTypeScreamingSnake:
		return nameutil.ToScreamingSnake(value)
	case caseTypeCamel:
		return nameutil.ToCamel(value)
	case caseTypeLowerCamel:
		return nameutil.ToLowerCamel(value)
	default:
		return value
	}
}
