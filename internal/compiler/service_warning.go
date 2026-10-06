package compiler

import (
	"fmt"

	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/model"
)

// MigrationDiagnostics reports declarations accepted only for compatibility.
func MigrationDiagnostics(domain *model.Domain) Diagnostics {
	var result Diagnostics
	for _, service := range domain.Services() {
		end := service.Pos
		end.Column += len(service.Name)
		span := diagnostic.SourceRange{Start: service.Pos, End: end}
		result = append(result, authMigrationDiagnostic(service.Auth, service.AuthPos, service.Name, false)...)
		for _, method := range service.Methods {
			result = append(result, authMigrationDiagnostic(method.Auth, method.AuthPos, service.Name+"/"+method.Name, false)...)
		}
		if service.Api && (service.Auth == "" || service.Auth == model.AuthModeUnset) {
			result = append(result, Diagnostic{Code: diagnostic.CodeApiAuthMissing, Severity: DiagnosticSeverityWarning, Position: service.Pos, Range: span,
				Message: fmt.Sprintf("API service %s must explicitly declare auth required, auth optional, or auth anonymous; defaults to required", service.Name)})
		}
		if service.Api {
			continue
		}
		if !service.Public() {
			result = append(result, Diagnostic{
				Code: diagnostic.CodeServiceModifier, Severity: DiagnosticSeverityWarning,
				Position: service.Pos, Range: span,
				Message: fmt.Sprintf("service %s has no modifier; declare pub for backend calls, ext for reusable server contracts, or api for portal clients", service.Name),
			})
		}
		if service.HasClientRules() {
			result = append(result, Diagnostic{
				Code: diagnostic.CodeServiceClientRules, Severity: DiagnosticSeverityWarning,
				Position: service.Pos, Range: span,
				Message: fmt.Sprintf("service %s declares client admission rules without api; declare api for portal clients", service.Name),
			})
		}
	}
	for _, web := range domain.Webs() {
		result = append(result, authMigrationDiagnostic(web.Auth, web.AuthPos, web.Name, true)...)
		if web.Auth == "" || web.Auth == model.AuthModeUnset {
			end := web.Pos
			end.Column += len(web.Name)
			result = append(result, Diagnostic{Code: diagnostic.CodeWebAuthMissing, Severity: DiagnosticSeverityWarning, Position: web.Pos,
				Range:   diagnostic.SourceRange{Start: web.Pos, End: end},
				Message: fmt.Sprintf("web %s must explicitly declare auth required, auth optional, auth anonymous, or auth off; defaults to required", web.Name)})
		}
	}
	return result
}

func authMigrationDiagnostic(mode model.AuthMode, pos model.Position, name string, web bool) Diagnostics {
	if mode != model.AuthModeAuth && mode != model.AuthModeNoAuth {
		return nil
	}
	replacement := "auth required"
	if mode == model.AuthModeNoAuth {
		replacement = "auth optional"
		if web {
			replacement = "auth off"
		}
	}
	end := pos
	end.Column += len(mode)
	var suggestion *diagnostic.Suggestion
	if pos.Line > 0 && pos.Column > 0 {
		suggestion = &diagnostic.Suggestion{Message: "Replace with " + replacement, Replacement: replacement, Replace: true}
	}
	return Diagnostics{{Suggestion: suggestion, Code: diagnostic.CodeAuthLegacy, Severity: DiagnosticSeverityWarning, Position: pos,
		Range:   diagnostic.SourceRange{Start: pos, End: end},
		Message: fmt.Sprintf("%s uses deprecated %s; replace it with %s", name, mode, replacement)}}
}
