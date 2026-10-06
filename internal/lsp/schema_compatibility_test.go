package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skel/internal/lsp/features"
	"go.yorun.ai/skel/internal/schema"
	"go.yorun.ai/skel/internal/testutil"
)

func TestSchemaCompatibilityCodeLensAndCommandUseInMemoryDocument(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int }\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "other.skel"), []byte("domain demo\ndata User { id: bool }\n"), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline", "contract.skel", "other.skel")

	server := newServer()
	documentURI := uri.File(path)
	server.loadWorkspace(uri.File(root))
	server.putDocument(documentURI, "domain demo\ndata User { id: string }\n", 2, true)

	lenses, err := server.CodeLens(t.Context(), &protocol.CodeLensParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}})
	require.NoError(t, err)
	require.Len(t, lenses, 1)
	assert.Equal(t, features.CommandShowSchemaCompatibility, lenses[0].Command.Command)
	assert.Equal(t, "Check schema compatibility", lenses[0].Command.Title)

	argument, err := json.Marshal(string(documentURI))
	require.NoError(t, err)
	value, err := server.ExecuteCommand(t.Context(), &protocol.ExecuteCommandParams{
		Command: commandSchemaDiff, Arguments: []protocol.LSPAny{argument},
	})
	require.NoError(t, err)
	var report schema.Report
	require.NoError(t, json.Unmarshal(value, &report))
	require.Len(t, report.Changes, 1)
	assert.Equal(t, schema.ImpactBreaking, report.Changes[0].Impact)
	assert.Equal(t, "data.member.type.changed", report.Changes[0].Code)
}

func TestInitializeAdvertisesSchemaCompatibilityCapabilities(t *testing.T) {
	server := newServer()
	options, err := json.Marshal(_InitializationOptions{SchemaCompatibility: _SchemaCompatibilitySettings{
		Diagnostics: true, IncludeCompatible: true, CodeLens: false, Baseline: "../baseline",
	}})
	require.NoError(t, err)
	result, err := server.Initialize(t.Context(), &protocol.InitializeParams{InitializationOptions: options})
	require.NoError(t, err)
	require.NotNil(t, result.Capabilities.CodeLensProvider)
	assert.Equal(t, []string{commandSchemaDiff}, result.Capabilities.ExecuteCommandProvider.Commands)
	assert.Equal(t, _SchemaCompatibilitySettings{
		Diagnostics: true, IncludeCompatible: true, CodeLens: false, Baseline: "../baseline",
	}, server.schemaCompatibility)
}

func TestRemoteSchemaCommandRetainsDocumentIdentity(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int }\n"), 0600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline", "contract.skel")
	server := newServer()
	remote := uri.URI("vscode-remote://ssh-remote+host" + string(uri.File(path))[len("file://"):])
	server.putDocument(remote, "domain demo\ndata User { id: string }\n", 1, true)
	argument, err := json.Marshal(string(remote))
	require.NoError(t, err)
	value, err := server.ExecuteCommand(t.Context(), &protocol.ExecuteCommandParams{Command: commandSchemaDiff, Arguments: []protocol.LSPAny{argument}})
	require.NoError(t, err)
	var report schema.Report
	require.NoError(t, json.Unmarshal(value, &report))
	require.Len(t, report.Changes, 1)
	assert.Equal(t, string(remote), report.Changes[0].Candidate.File)
}
