package lsp

import (
	"context"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	lspdiagnostic "go.yorun.ai/skel/internal/lsp/diagnostic"
	"go.yorun.ai/skel/internal/lsp/workspace"
)

func (s *_Server) publishDiagnostics(ctx context.Context, documentURI uri.URI) {
	client, ok := protocol.ClientFromContext(ctx)
	if !ok {
		return
	}
	s.rememberClient(ctx)
	s.publishDocumentDiagnostics(client, documentURI, s.workspace.Document(documentURI))
}

func (s *_Server) publishDocumentDiagnostics(client protocol.Client, documentURI uri.URI, document *workspace.Document) {
	s.mu.RLock()
	semantic := append([]protocol.Diagnostic{}, s.semantic[documentURI]...)
	s.mu.RUnlock()
	diagnostics := semantic
	if document != nil {
		for _, diagnostic := range document.ParseDiagnostics {
			diagnostics = append(diagnostics, lspdiagnostic.ToProtocolBuffer(diagnostic, document.Buffer, nil))
		}
	}
	params := &protocol.PublishDiagnosticsParams{URI: documentURI, Diagnostics: diagnostics}
	if document != nil {
		params.Version = protocol.NewOptional(document.Version)
	}
	s.publisher.enqueue(_DiagnosticBatch{generation: s.diagnosticGeneration, client: client, params: params, removed: document == nil})
}
