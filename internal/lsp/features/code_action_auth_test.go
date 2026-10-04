package features

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/compiler"
	lspdiagnostic "go.yorun.ai/skelc/internal/lsp/diagnostic"
	lsource "go.yorun.ai/skelc/internal/lsp/source"
)

func TestAuthMigrationQuickFix(t *testing.T) {
	for _, test := range []struct{ declaration, marker, replacement string }{
		{"api service TestApiService { for ClientActor via client\n /* 中文🙂 */ auth // keep\n method ping {} }", "auth", "auth required"},
		{"api service TestApiService { for ClientActor via client\n noauth // keep\n method ping {} }", "noauth", "auth optional"},
		{"api service TestApiService { for ClientActor via client auth required\n method ping { auth /* keep */ } }", "auth", "auth required"},
		{"api service TestApiService { for ClientActor via client auth required\n method ping { noauth /* keep */ } }", "noauth", "auth optional"},
		{"web TestWeb { for ClientActor via client\n auth // keep\n }", "auth", "auth required"},
		{"web TestWeb { for ClientActor via client\n noauth // keep\n }", "noauth", "auth off"},
	} {
		t.Run(test.declaration, func(t *testing.T) {
			source := "domain demo.user\nactor ClientActor { via client {} }\n" + test.declaration + "\n"
			path := filepath.Join(t.TempDir(), "domain.skel")
			require.NoError(t, os.WriteFile(path, []byte(source), 0600))
			result, err := compiler.Compile(compiler.Option{SkelIn: path})
			require.NoError(t, err)
			require.Len(t, result.Diagnostics, 1)
			item := result.Diagnostics[0]
			require.Equal(t, diagnostic.CodeAuthLegacy, item.Code)
			for _, strict := range []bool{false, true} {
				diagnostics := compiler.Diagnostics{item}
				if strict {
					compiler.ApplyStrictMode(diagnostics)
				}
				converted := lspdiagnostic.ToProtocol(diagnostics[0], source, nil)
				documentURI := uri.File(path)
				actions, err := newFixture().service().CodeAction(t.Context(), &protocol.CodeActionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
					Context:      protocol.CodeActionContext{Diagnostics: []protocol.Diagnostic{converted}},
				})
				require.NoError(t, err)
				require.Len(t, actions, 1)
				edits := actions[0].(*protocol.CodeAction).Edit.Changes[documentURI]
				require.Len(t, edits, 1)
				edit := edits[0]
				buffer := lsource.New(source)
				start, end := buffer.Offset(edit.Range.Start), buffer.Offset(edit.Range.End)
				require.Equal(t, test.marker, source[start:end])
				require.Equal(t, test.replacement, edit.NewText)
				fixed := source[:start] + edit.NewText + source[end:]
				require.NoError(t, os.WriteFile(path, []byte(fixed), 0600))
				compiled, err := compiler.Compile(compiler.Option{SkelIn: path, Strict: true})
				require.NoError(t, err)
				require.Empty(t, compiled.Diagnostics)
			}
		})
	}
}
