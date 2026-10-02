package lsp

import (
	"context"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	lspdiagnostic "go.yorun.ai/skelc/internal/lsp/diagnostic"
	"go.yorun.ai/skelc/internal/lsp/index"
)

func (s *_Server) publishDiagnostics(ctx context.Context, documentURI uri.URI) error {
	client, ok := protocol.ClientFromContext(ctx)
	if !ok {
		return nil
	}
	s.rememberClient(ctx)
	return s.publishDiagnosticsWithClient(ctx, client, documentURI)
}

func (s *_Server) publishDiagnosticsWithClient(ctx context.Context, client protocol.Client, documentURI uri.URI) error {
	return s.publishDocumentDiagnostics(ctx, client, documentURI, s.workspace.Document(documentURI))
}

func (s *_Server) publishDocumentDiagnostics(ctx context.Context, client protocol.Client, documentURI uri.URI, document *index.Document) error {
	s.mu.RLock()
	semantic := append([]protocol.Diagnostic{}, s.semantic[documentURI]...)
	s.mu.RUnlock()
	diagnostics := semantic
	if document != nil {
		for _, diagnostic := range document.ParseDiagnostics {
			diagnostics = append(diagnostics, lspdiagnostic.ToProtocol(diagnostic, document.Source, nil))
		}
	}
	params := &protocol.PublishDiagnosticsParams{URI: documentURI, Diagnostics: diagnostics}
	if document != nil {
		params.Version = protocol.NewOptional(document.Version)
	}
	return client.PublishDiagnostics(ctx, params)
}
