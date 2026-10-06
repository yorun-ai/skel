package features

import (
	"context"
	"encoding/json"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	compiler "go.yorun.ai/skel/internal/compiler"
)

// CommandShowSchemaCompatibility is the client command used by schema
// diagnostics and CodeLens entries to display the complete diff report.
const CommandShowSchemaCompatibility = "skel.showSchemaCompatibility"

type _CompatibilityDiagnosticData struct {
	Impact string `json:"impact"`
}

func (s *Service) CodeAction(_ context.Context, params *protocol.CodeActionParams) ([]protocol.CommandOrCodeAction, error) {
	actions := []protocol.CommandOrCodeAction{}
	for _, diagnostic := range params.Context.Diagnostics {
		if len(diagnostic.Data) == 0 {
			continue
		}
		var compatibility _CompatibilityDiagnosticData
		if json.Unmarshal(diagnostic.Data, &compatibility) == nil && compatibility.Impact != "" {
			argument, _ := json.Marshal(string(params.TextDocument.URI))
			actions = append(actions, &protocol.CodeAction{
				Title: "Show full schema compatibility report", Diagnostics: []protocol.Diagnostic{diagnostic},
				Command: protocol.Command{
					Title: "Show full schema compatibility report", Command: CommandShowSchemaCompatibility,
					Arguments: []protocol.LSPAny{argument},
				},
			})
			continue
		}
		var suggestion compiler.DiagnosticSuggestion
		if err := json.Unmarshal(diagnostic.Data, &suggestion); err != nil || suggestion.Replacement == "" {
			continue
		}
		range_ := diagnostic.Range
		if !suggestion.Replace {
			range_.End = range_.Start
		}
		preferred := true
		kind := protocol.CodeActionKindQuickFix
		actions = append(actions, &protocol.CodeAction{
			Title: suggestion.Message, Kind: &kind, Diagnostics: []protocol.Diagnostic{diagnostic}, IsPreferred: &preferred,
			Edit: &protocol.WorkspaceEdit{Changes: map[uri.URI][]protocol.TextEdit{
				params.TextDocument.URI: {{Range: range_, NewText: suggestion.Replacement}},
			}},
		})
	}
	return actions, nil
}
