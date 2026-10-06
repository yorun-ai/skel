// Package lsp implements the Skel language server used by editor integrations.
package lsp

import (
	"context"
	"io"
	"sync"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skel/internal/lsp/analysis"
	"go.yorun.ai/skel/internal/lsp/workspace"
)

type _Server struct {
	protocol.UnimplementedServer

	mu                     sync.RWMutex
	stateMu                sync.Mutex // Serializes state transitions and immutable batch creation.
	diagnosticGeneration   uint64
	publisher              *_DiagnosticPublisher
	workspace              *workspace.Store
	semantic               map[uri.URI][]protocol.Diagnostic
	client                 protocol.Client
	analysis               *analysis.Runner
	snippetSupport         bool
	documentChangesSupport bool
	codeLensRefreshSupport bool
	schemaCompatibility    _SchemaCompatibilitySettings
	strict                 bool
	exit                   chan struct{}
	exitOnce               sync.Once
}

type _ReadWriteCloser struct {
	io.Reader
	io.Writer
	closer io.Closer
}

func (rw *_ReadWriteCloser) Close() error {
	if rw.closer != nil {
		return rw.closer.Close()
	}
	return nil
}

// Serve runs a Language Server Protocol connection over the supplied streams.
func Serve(ctx context.Context, input io.Reader, output io.Writer, strict bool) error {
	server := newServer()
	defer server.stopSemanticAnalysis()
	server.strict = strict
	closer, _ := input.(io.Closer)
	stream := jsonrpc2.NewStream(&_ReadWriteCloser{Reader: input, Writer: output, closer: closer})
	_, connection, _ := protocol.NewServer(ctx, server, stream)
	select {
	case <-connection.Done():
	case <-server.exit:
		if err := connection.Close(); err != nil {
			return err
		}
	}
	return connection.Err()
}

func newServer() *_Server {
	return &_Server{
		workspace:           workspace.New(),
		publisher:           newDiagnosticPublisher(),
		semantic:            map[uri.URI][]protocol.Diagnostic{},
		exit:                make(chan struct{}),
		analysis:            analysis.NewRunner(semanticAnalysisDelay),
		schemaCompatibility: defaultSchemaCompatibilitySettings(),
	}
}

func (s *_Server) Initialize(_ context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	s.mu.Lock()
	s.schemaCompatibility = decodeInitializationSettings(params.InitializationOptions)
	s.strict = decodeStrictSettings(params.InitializationOptions, s.strict)
	s.mu.Unlock()
	if textDocument := params.Capabilities.TextDocument; textDocument != nil &&
		textDocument.Completion != nil &&
		textDocument.Completion.CompletionItem != nil &&
		textDocument.Completion.CompletionItem.SnippetSupport != nil {
		s.snippetSupport = *textDocument.Completion.CompletionItem.SnippetSupport
	}
	if workspace := params.Capabilities.Workspace; workspace != nil && workspace.CodeLens != nil &&
		workspace.CodeLens.RefreshSupport != nil {
		s.codeLensRefreshSupport = *workspace.CodeLens.RefreshSupport
	}
	if workspace := params.Capabilities.Workspace; workspace != nil && workspace.WorkspaceEdit != nil && workspace.WorkspaceEdit.DocumentChanges != nil {
		s.documentChangesSupport = *workspace.WorkspaceEdit.DocumentChanges
	}
	if folders, ok := params.WorkspaceFolders.Get(); ok {
		for _, folder := range folders {
			s.loadWorkspace(folder.URI)
		}
	} else {
		//lint:ignore SA1019 Older LSP clients may send rootUri without workspaceFolders.
		if rootURI := params.RootURI; rootURI != nil {
			s.loadWorkspace(*rootURI)
		} else {
			//lint:ignore SA1019 Preserve the rootPath fallback for legacy LSP clients.
			if rootPath, ok := params.RootPath.Get(); ok {
				s.loadWorkspace(uri.File(rootPath))
			}
		}
	}
	openClose := true
	prepareRename := true
	workspaceFolders := true
	change := protocol.TextDocumentSyncKindFull
	resolveCodeLens := false
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			PositionEncoding:           protocol.PositionEncodingKindUTF16,
			TextDocumentSync:           &protocol.TextDocumentSyncOptions{OpenClose: &openClose, Change: &change},
			CompletionProvider:         &protocol.CompletionOptions{TriggerCharacters: []string{".", "@"}},
			HoverProvider:              protocol.Boolean(true),
			DefinitionProvider:         protocol.Boolean(true),
			ReferencesProvider:         protocol.Boolean(true),
			DocumentSymbolProvider:     protocol.Boolean(true),
			WorkspaceSymbolProvider:    protocol.Boolean(true),
			DocumentFormattingProvider: protocol.Boolean(true),
			RenameProvider:             &protocol.RenameOptions{PrepareProvider: &prepareRename},
			CodeActionProvider:         protocol.Boolean(true),
			CodeLensProvider:           &protocol.CodeLensOptions{ResolveProvider: &resolveCodeLens},
			ExecuteCommandProvider:     protocol.ExecuteCommandOptions{Commands: []string{commandSchemaDiff}},
			Workspace: &protocol.WorkspaceOptions{WorkspaceFolders: &protocol.WorkspaceFoldersServerCapabilities{
				Supported: &workspaceFolders, ChangeNotifications: protocol.Boolean(true),
			}},
		},
		ServerInfo: protocol.ServerInfo{Name: "skelc"},
	}, nil
}

func (s *_Server) Initialized(ctx context.Context, _ *protocol.InitializedParams) error {
	s.rememberClient(ctx)
	s.scheduleSemanticAnalysis()
	return nil
}

func (s *_Server) Shutdown(context.Context) error {
	s.stopSemanticAnalysis()
	return nil
}

func (s *_Server) Exit(context.Context) error {
	s.stopSemanticAnalysis()
	s.exitOnce.Do(func() { close(s.exit) })
	return nil
}
