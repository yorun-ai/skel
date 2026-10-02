package lsp

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/lsp/analysis"
)

type slowDiagnosticClient struct {
	protocol.UnimplementedClient
	started   chan struct{}
	release   chan struct{}
	published chan *protocol.PublishDiagnosticsParams
	finished  chan struct{}
	once      sync.Once
}

func (c *slowDiagnosticClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
	first := false
	c.once.Do(func() { first = true; close(c.started) })
	if first {
		if c.finished != nil {
			defer close(c.finished)
		}
		select {
		case <-c.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	select {
	case c.published <- params:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestSlowClientDoesNotBlockEditsAndDropsQueuedOldVersions(t *testing.T) {
	server := newServer()
	server.analysis = analysis.NewRunner(time.Hour)
	t.Cleanup(server.stopSemanticAnalysis)
	client := &slowDiagnosticClient{started: make(chan struct{}), release: make(chan struct{}), published: make(chan *protocol.PublishDiagnosticsParams, 8)}
	server.client = client
	documentURI := uri.File("/workspace/input.skel")
	server.workspace.Put(documentURI, "domain demo\n", 1, true)
	server.acceptSemanticAnalysis(analysis.Result{Revision: server.workspace.Revision(), Diagnostics: map[uri.URI][]protocol.Diagnostic{documentURI: {{Message: protocol.String("old")}}}})
	select {
	case <-client.started:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not start")
	}
	edited := make(chan error, 1)
	go func() {
		for version := int32(2); version <= 20; version++ {
			err := server.DidChange(t.Context(), &protocol.DidChangeTextDocumentParams{TextDocument: protocol.VersionedTextDocumentIdentifier{TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI}, Version: version}, ContentChanges: []protocol.TextDocumentContentChangeEvent{&protocol.TextDocumentContentChangeWholeDocument{Text: "domain demo\n"}}})
			if err != nil {
				edited <- err
				return
			}
		}
		edited <- nil
	}()
	select {
	case err := <-edited:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("editing blocked on diagnostic transport")
	}
	close(client.release)
	first := receiveDiagnostics(t, client.published)
	require.Equal(t, protocol.NewOptional(int32(1)), first.Version)
	latest := receiveDiagnostics(t, client.published)
	require.Equal(t, protocol.NewOptional(int32(20)), latest.Version)
	require.Empty(t, latest.Diagnostics)
}

func TestPublisherCoalescesClosedDocumentsAndCancelsOnShutdown(t *testing.T) {
	publisher := newDiagnosticPublisher()
	defer publisher.stop()
	client := &slowDiagnosticClient{started: make(chan struct{}), release: make(chan struct{}), published: make(chan *protocol.PublishDiagnosticsParams, 8)}
	active := uri.URI("untitled:active")
	publisher.enqueue(_DiagnosticBatch{client: client, params: &protocol.PublishDiagnosticsParams{URI: active, Version: protocol.NewOptional(int32(1))}})
	select {
	case <-client.started:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not start")
	}
	for i := range 1000 {
		documentURI := uri.URI(fmt.Sprintf("untitled:%d", i))
		publisher.enqueue(_DiagnosticBatch{client: client, params: &protocol.PublishDiagnosticsParams{URI: documentURI}})
		publisher.enqueue(_DiagnosticBatch{client: client, params: &protocol.PublishDiagnosticsParams{URI: documentURI}, removed: true})
	}
	publisher.mu.Lock()
	pending := len(publisher.pending)
	publisher.mu.Unlock()
	require.Zero(t, pending, "unpublished closed documents must not accumulate")
	publisher.enqueue(_DiagnosticBatch{client: client, params: &protocol.PublishDiagnosticsParams{URI: active}, removed: true})
	close(client.release)
	_ = receiveDiagnostics(t, client.published)
	cleared := receiveDiagnostics(t, client.published)
	require.Equal(t, active, cleared.URI)
	require.Empty(t, cleared.Diagnostics)
}

func TestPublisherShutdownCancelsBlockedTransport(t *testing.T) {
	publisher := newDiagnosticPublisher()
	defer publisher.stop()
	client := &slowDiagnosticClient{started: make(chan struct{}), release: make(chan struct{}), published: make(chan *protocol.PublishDiagnosticsParams), finished: make(chan struct{})}
	publisher.enqueue(_DiagnosticBatch{client: client, params: &protocol.PublishDiagnosticsParams{URI: "untitled:blocked"}})
	select {
	case <-client.started:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not start")
	}
	publisher.stop()
	select {
	case <-client.finished:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not cancel transport")
	}
}
