package lsp

import (
	"context"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skel/internal/lsp/features"
)

func (s *_Server) featureService() features.Service {
	s.mu.RLock()
	snippetSupport := s.snippetSupport
	documentChangesSupport := s.documentChangesSupport
	s.mu.RUnlock()
	return features.Service{Snapshot: s.workspace.Snapshot(), SnippetSupport: snippetSupport, DocumentChangesSupport: documentChangesSupport}
}

func (s *_Server) Completion(ctx context.Context, params *protocol.CompletionParams) (protocol.CompletionResult, error) {
	service := s.featureService()
	return service.Completion(ctx, params)
}

func (s *_Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	service := s.featureService()
	return service.Hover(ctx, params)
}

func (s *_Server) Definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.DefinitionResult, error) {
	service := s.featureService()
	return service.Definition(ctx, params)
}

func (s *_Server) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	service := s.featureService()
	return service.References(ctx, params)
}

func (s *_Server) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) (protocol.DocumentSymbolResult, error) {
	service := s.featureService()
	return service.DocumentSymbol(ctx, params)
}

func (s *_Server) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) (protocol.WorkspaceSymbolResult, error) {
	service := s.featureService()
	return service.Symbols(ctx, params)
}

func (s *_Server) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (protocol.PrepareRenameResult, error) {
	service := s.featureService()
	return service.PrepareRename(ctx, params)
}

func (s *_Server) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	service := s.featureService()
	return service.Rename(ctx, params)
}

func (s *_Server) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	service := s.featureService()
	return service.Formatting(ctx, params)
}

func (s *_Server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CommandOrCodeAction, error) {
	service := s.featureService()
	return service.CodeAction(ctx, params)
}
