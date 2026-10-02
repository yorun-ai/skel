package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
)

func FuzzDocumentChanges(f *testing.F) {
	for _, input := range []string{
		`null`, `{}`, `{"textDocument":{"uri":"untitled:audit","version":1},"contentChanges":[]}`,
		`{"textDocument":{"uri":"untitled:audit","version":1},"contentChanges":[null]}`,
		`{"textDocument":{"uri":"untitled:audit","version":1},"contentChanges":[{"text":"domain demo\ndata Value {}"}]}`,
		`{"textDocument":{"uri":"untitled:audit","version":1},"contentChanges":[{"text":"x","range":{"start":{"line":4294967295,"character":4294967295},"end":{"line":0,"character":0}}}]}`,
	} {
		f.Add([]byte(input))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 4096 {
			t.Skip()
		}
		params := new(protocol.DidChangeTextDocumentParams)
		if err := protocol.Unmarshal(input, params); err != nil {
			return
		}
		server := newServer()
		defer server.stopSemanticAnalysis()
		// Invalid synchronization variants may return an error, but must not panic.
		_ = server.DidChange(t.Context(), params)
	})
}
