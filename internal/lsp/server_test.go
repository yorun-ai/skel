package lsp

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type recordingClient struct {
	protocol.UnimplementedClient
	diagnostics chan *protocol.PublishDiagnosticsParams
}

func (c *recordingClient) PublishDiagnostics(_ context.Context, params *protocol.PublishDiagnosticsParams) error {
	c.diagnostics <- params
	return nil
}

func TestServeLifecycle(t *testing.T) {
	serverStream, clientStream := net.Pipe()
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- Serve(t.Context(), serverStream, serverStream, false)
	}()

	_, connection, server := protocol.NewClient(
		t.Context(), protocol.UnimplementedClient{}, jsonrpc2.NewStream(clientStream),
	)
	t.Cleanup(func() { _ = connection.Close() })

	root := uri.File(t.TempDir())
	result, err := server.Initialize(t.Context(), &protocol.InitializeParams{
		WorkspaceFolders: protocol.NewNullable([]protocol.WorkspaceFolder{{URI: root, Name: "test"}}),
	})
	require.NoError(t, err)
	assert.Equal(t, protocol.PositionEncodingKindUTF16, result.Capabilities.PositionEncoding)
	require.NotNil(t, result.Capabilities.CompletionProvider)
	assert.Equal(t, []string{".", "@"}, result.Capabilities.CompletionProvider.TriggerCharacters)
	assert.Equal(t, protocol.Boolean(true), result.Capabilities.HoverProvider)
	assert.Equal(t, protocol.Boolean(true), result.Capabilities.WorkspaceSymbolProvider)
	assert.Equal(t, protocol.Boolean(true), result.Capabilities.DocumentFormattingProvider)
	require.NotNil(t, result.Capabilities.CodeLensProvider)
	assert.Equal(t, []string{commandSchemaDiff}, result.Capabilities.ExecuteCommandProvider.Commands)
	require.NotNil(t, result.Capabilities.Workspace)
	require.NotNil(t, result.Capabilities.Workspace.WorkspaceFolders)
	require.NotNil(t, result.Capabilities.Workspace.WorkspaceFolders.Supported)
	assert.True(t, *result.Capabilities.Workspace.WorkspaceFolders.Supported)
	assert.Equal(t, protocol.Boolean(true), result.Capabilities.Workspace.WorkspaceFolders.ChangeNotifications)
	rename, ok := result.Capabilities.RenameProvider.(*protocol.RenameOptions)
	require.True(t, ok)
	require.NotNil(t, rename.PrepareProvider)
	assert.True(t, *rename.PrepareProvider)
	require.NoError(t, server.Initialized(t.Context(), &protocol.InitializedParams{}))
	require.NoError(t, server.Shutdown(t.Context()))
	require.NoError(t, server.Exit(t.Context()))

	select {
	case err := <-serverDone:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("language server did not exit after the exit notification")
	}
}

func TestInitializeLoadsWorkspaceRootsWithLegacyFallbacks(t *testing.T) {
	root := t.TempDir()
	documents := map[string]uri.URI{}
	for _, name := range []string{"folders", "uri", "path"} {
		directory := filepath.Join(root, name)
		require.NoError(t, os.Mkdir(directory, 0o700))
		path := filepath.Join(directory, "input.skel")
		require.NoError(t, os.WriteFile(path, []byte("domain demo."+name+"\ndata Value { id: string }\n"), 0o600))
		documents[name] = uri.File(path)
	}
	for _, test := range []struct {
		name         string
		folders      bool
		emptyFolders bool
		rootURI      bool
		rootPath     bool
		want         string
	}{
		{name: "workspace folders take precedence", folders: true, rootURI: true, rootPath: true, want: "folders"},
		{name: "root URI takes precedence over path", rootURI: true, rootPath: true, want: "uri"},
		{name: "root path fallback", rootPath: true, want: "path"},
		{name: "explicit empty folders disable fallback", emptyFolders: true, rootURI: true, rootPath: true},
		{name: "no workspace roots"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := newServer()
			t.Cleanup(server.stopSemanticAnalysis)
			params := new(protocol.InitializeParams)
			if test.folders {
				params.WorkspaceFolders = protocol.NewNullable([]protocol.WorkspaceFolder{{URI: uri.File(filepath.Join(root, "folders")), Name: "test"}})
			} else if test.emptyFolders {
				params.WorkspaceFolders = protocol.NewNullable([]protocol.WorkspaceFolder{})
			}
			if test.rootURI {
				//lint:ignore SA1019 Verify the supported rootUri fallback and its precedence.
				params.RootURI = new(uri.File(filepath.Join(root, "uri")))
			}
			if test.rootPath {
				//lint:ignore SA1019 Verify workspace discovery for legacy clients sending rootPath.
				params.RootPath = protocol.NewNullable(filepath.Join(root, "path"))
			}
			_, err := server.Initialize(t.Context(), params)
			require.NoError(t, err)
			snapshot := server.workspace.Snapshot()
			if test.want == "" {
				assert.Empty(t, snapshot.Documents())
				return
			}
			require.Len(t, snapshot.Documents(), 1)
			assert.NotNil(t, snapshot.Document(documents[test.want]))
		})
	}
}

func waitForDiagnostics(
	t *testing.T,
	diagnostics <-chan *protocol.PublishDiagnosticsParams,
	accept func(*protocol.PublishDiagnosticsParams) bool,
) *protocol.PublishDiagnosticsParams {
	t.Helper()
	// The wait covers analysis work rather than a fixed delay, and race-enabled
	// suite runs share a small CI runner, so keep a generous budget.
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case params := <-diagnostics:
			if accept(params) {
				return params
			}
		case <-timer.C:
			t.Fatal("timed out waiting for matching diagnostics")
		}
	}
}
