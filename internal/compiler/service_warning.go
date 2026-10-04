package compiler

import (
	"fmt"

	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/model"
)

// MigrationDiagnostics reports declarations accepted only for compatibility.
func MigrationDiagnostics(domain *model.Domain) Diagnostics {
	var result Diagnostics
	for _, service := range domain.Services() {
		end := service.Pos
		end.Column += len(service.Name)
		span := diagnostic.SourceRange{Start: service.Pos, End: end}
		if service.Api {
			continue
		}
		if !service.Public() {
			result = append(result, Diagnostic{
				Code: diagnostic.CodeServiceModifier, Severity: DiagnosticSeverityWarning,
				Position: service.Pos, Range: span,
				Message: fmt.Sprintf("service %s has no modifier; declare pub for backend calls, ext for reusable server contracts, or api for Portal clients", service.Name),
			})
		}
		if service.HasClientRules() {
			result = append(result, Diagnostic{
				Code: diagnostic.CodeServiceClientRules, Severity: DiagnosticSeverityWarning,
				Position: service.Pos, Range: span,
				Message: fmt.Sprintf("service %s declares client admission rules without api; declare api for Portal clients", service.Name),
			})
		}
	}
	return result
}
