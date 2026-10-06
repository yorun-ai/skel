package lsp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skel/internal/lsp/analysis"
)

func receiveDiagnostics(t *testing.T, diagnostics <-chan *protocol.PublishDiagnosticsParams) *protocol.PublishDiagnosticsParams {
	t.Helper()
	select {
	case params := <-diagnostics:
		return params
	case <-time.After(5 * time.Second):
		t.Fatal("expected published diagnostics")
		return nil
	}
}

// The analysis runner never fires; only accepted snapshots reach the publisher.
func TestSemanticDiagnosticsPublishAndInvalidate(t *testing.T) {
	server := newServer()
	server.analysis = analysis.NewRunner(time.Hour)
	t.Cleanup(server.stopSemanticAnalysis)

	client := &recordingClient{diagnostics: make(chan *protocol.PublishDiagnosticsParams, 16)}
	server.client = client

	documentURI := uri.File("/workspace/order.skel")
	require.NoError(t, server.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{TextDocument: protocol.TextDocumentItem{
		URI: documentURI, LanguageID: "skel", Version: 1,
		Text: "domain demo.order\ndata Order { owner: Missing }\n",
	}}))

	server.acceptSemanticAnalysis(analysis.Result{
		Revision: server.workspace.Snapshot().Revision(),
		Diagnostics: map[uri.URI][]protocol.Diagnostic{
			documentURI: {{
				Severity: protocol.DiagnosticSeverityError,
				Code:     protocol.String("semantic.reference"),
				Message:  protocol.String("unknown type Missing"),
			}},
		},
	})

	published := receiveDiagnostics(t, client.diagnostics)
	assert.Equal(t, documentURI, published.URI)
	assert.Equal(t, protocol.NewOptional(int32(1)), published.Version)
	require.Len(t, published.Diagnostics, 1)
	assert.Equal(t, protocol.String("semantic.reference"), published.Diagnostics[0].Code)

	require.NoError(t, server.DidChange(t.Context(), &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI}, Version: 2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{&protocol.TextDocumentContentChangeWholeDocument{
			Text: "domain demo.order\ndata Order { owner: int }\n",
		}},
	}))

	invalidated := receiveDiagnostics(t, client.diagnostics)
	assert.Equal(t, documentURI, invalidated.URI)
	assert.Equal(t, protocol.NewOptional(int32(2)), invalidated.Version)
	assert.Empty(t, invalidated.Diagnostics)
}

func TestQueuedSemanticResultCannotRestoreInvalidatedDiagnostics(t *testing.T) {
	s := newServer()
	s.analysis = analysis.NewRunner(time.Hour)
	t.Cleanup(s.stopSemanticAnalysis)
	client := &recordingClient{diagnostics: make(chan *protocol.PublishDiagnosticsParams, 16)}
	s.client = client
	u := uri.File("/workspace/input.skel")
	s.workspace.Put(u, "domain demo\ndata User { value: Missing }\n", 1, true)
	result := analysis.Result{Revision: s.workspace.Revision(), Diagnostics: map[uri.URI][]protocol.Diagnostic{u: {{Message: protocol.String("stale")}}}}
	s.stateMu.Lock()
	done := make(chan struct{})
	go func() { defer close(done); s.acceptSemanticAnalysis(result) }()
	s.workspace.Put(u, "domain demo\ndata User { value: int }\n", 2, true)
	s.invalidateSemanticDiagnostics(t.Context())
	s.stateMu.Unlock()
	<-done
	assert.Empty(t, s.semantic)
	assert.Empty(t, client.diagnostics)
}

func TestSupersededConfigurationResultIsRejected(t *testing.T) {
	s := newServer()
	s.analysis = analysis.NewRunner(time.Hour)
	t.Cleanup(s.stopSemanticAnalysis)
	client := &recordingClient{diagnostics: make(chan *protocol.PublishDiagnosticsParams, 16)}
	s.client = client
	u := uri.File("/workspace/input.skel")
	s.workspace.Put(u, "domain demo\n", 1, true)
	s.scheduleSemanticAnalysis()
	s.scheduleSemanticAnalysis()
	s.acceptSemanticAnalysis(analysis.Result{Generation: 1, Revision: s.workspace.Revision(), Diagnostics: map[uri.URI][]protocol.Diagnostic{u: {{Message: protocol.String("old settings")}}}})
	assert.Empty(t, s.semantic)
	assert.Empty(t, client.diagnostics)
}
