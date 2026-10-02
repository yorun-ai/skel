package lsp

import (
	"context"
	"errors"
	"slices"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (s *_Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	document := params.TextDocument
	s.putDocument(document.URI, document.Text, document.Version, true)
	changed := s.workspace.RefreshDirectory(document.URI)
	s.invalidateSemanticDiagnostics(ctx)
	for _, documentURI := range changed {
		if err := s.publishDiagnostics(ctx, documentURI); err != nil {
			return err
		}
	}
	return s.publishDiagnostics(ctx, document.URI)
}

func (s *_Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if len(params.ContentChanges) == 0 {
		return nil
	}
	change, ok := params.ContentChanges[len(params.ContentChanges)-1].(*protocol.TextDocumentContentChangeWholeDocument)
	if !ok {
		return errors.New("skelc lsp requires full document synchronization")
	}
	s.putDocument(params.TextDocument.URI, change.Text, params.TextDocument.Version, true)
	s.invalidateSemanticDiagnostics(ctx)
	return s.publishDiagnostics(ctx, params.TextDocument.URI)
}

func (s *_Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	documentURI := params.TextDocument.URI
	exists := s.workspace.Close(documentURI)
	s.invalidateSemanticDiagnostics(ctx)
	if exists {
		return s.publishDiagnostics(ctx, documentURI)
	}
	client, ok := protocol.ClientFromContext(ctx)
	if !ok {
		return nil
	}
	return s.publishDocumentDiagnostics(ctx, client, documentURI, nil)
}

func (s *_Server) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	changed := s.workspace.ApplyFileChanges(params.Changes)
	s.invalidateSemanticDiagnostics(ctx)
	for _, documentURI := range changed {
		if err := s.publishDiagnostics(ctx, documentURI); err != nil {
			return err
		}
	}
	return nil
}

func (s *_Server) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	removed := make([]uri.URI, 0)
	for _, folder := range params.Event.Removed {
		removed = append(removed, s.workspace.RemoveRoot(folder.URI)...)
	}
	for _, folder := range params.Event.Added {
		s.workspace.AddRoot(folder.URI)
	}
	s.invalidateSemanticDiagnostics(ctx)
	client, ok := protocol.ClientFromContext(ctx)
	if !ok {
		return nil
	}
	slices.Sort(removed)
	snapshot := s.workspace.Snapshot()
	for _, documentURI := range removed {
		if snapshot.Document(documentURI) != nil {
			continue
		}
		if err := s.publishDocumentDiagnostics(ctx, client, documentURI, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *_Server) putDocument(documentURI uri.URI, content string, version int32, open bool) {
	s.workspace.Put(documentURI, content, version, open)
}

func (s *_Server) loadWorkspace(rootURI uri.URI) {
	s.workspace.AddRoot(rootURI)
}
