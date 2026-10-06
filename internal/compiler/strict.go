package compiler

import "go.yorun.ai/skel/diagnostic"

// ApplyStrictMode promotes migration warnings to errors in place.
func ApplyStrictMode(diagnostics Diagnostics) {
	for index := range diagnostics {
		item := &diagnostics[index]
		if item.Severity != DiagnosticSeverityWarning {
			continue
		}
		switch item.Code {
		case diagnostic.CodeServiceModifier, diagnostic.CodeServiceClientRules, diagnostic.CodeAuthLegacy, diagnostic.CodeApiAuthMissing, diagnostic.CodeWebAuthMissing:
			item.Severity = DiagnosticSeverityError
		}
	}
}
