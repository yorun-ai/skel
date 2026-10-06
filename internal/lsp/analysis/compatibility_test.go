package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/lsp/workspace"
	"go.yorun.ai/skel/internal/sourcediff"
	"go.yorun.ai/skel/internal/testutil"
)

func TestCompatibilityDiagnosticsLocateActorCapabilities(t *testing.T) {
	for _, capability := range []string{"auth", "permission"} {
		for _, added := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/added=%t", capability, added), func(t *testing.T) {
				root := t.TempDir()
				path, baselinePath := filepath.Join(root, "contract.skel"), filepath.Join(root, "baseline.skel")
				body := "permission {}"
				if capability == "auth" {
					body = "auth { credential { token: string } info { id: string } }"
				}
				before := "domain demo\nactor UserActor {\n    via client {}\n}\n"
				after := "domain demo\nactor UserActor {\n    via client {}\n    " + body + "\n}\n"
				code := "schema.actor." + capability + ".added"
				position := protocol.Position{Line: 3, Character: 4}
				if !added {
					before, after = after, before
					code = "schema.actor." + capability + ".removed"
					position = protocol.Position{Line: 1, Character: 6}
				}
				require.NoError(t, os.WriteFile(baselinePath, []byte(before), 0o600))
				documentURI := uri.File(path)
				document := workspace.BuildDocument(documentURI, path, after, 1)
				sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})
				diagnostics, domains, err := SemanticWorkspace(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths, false)
				require.NoError(t, err)
				require.Empty(t, diagnostics)
				appendCompatibilityDiagnostics(t.Context(), sourcediff.New(), diagnostics, domains, sources, paths,
					CompatibilityOptions{Enabled: true, BaselineSkelIn: baselinePath})
				require.Len(t, diagnostics[documentURI], 1)
				result := diagnostics[documentURI][0]
				assert.Equal(t, protocol.String(code), result.Code)
				assert.Equal(t, position, result.Range.Start)
			})
		}
	}
}

func TestCompatibilityDiagnosticsUseInMemorySourceAndImpactSeverity(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int }\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "other.skel"), []byte("domain demo\ndata User { id: bool }\n"), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline", "contract.skel", "other.skel")

	documentURI := uri.File(path)
	document := workspace.BuildDocument(documentURI, path, "domain demo\ndata User { id: string }\n", 2)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})
	analyzer := compiler.NewWorkspaceAnalyzer()
	diagnostics, domains, err := SemanticWorkspace(t.Context(), analyzer, sources, paths, false)
	require.NoError(t, err)
	require.Empty(t, diagnostics)

	appendCompatibilityDiagnostics(t.Context(), sourcediff.New(), diagnostics, domains, sources, paths, CompatibilityOptions{Enabled: true})
	require.Len(t, diagnostics[documentURI], 1)
	result := diagnostics[documentURI][0]
	assert.Equal(t, protocol.String("schema.data.member.type.changed"), result.Code)
	assert.Equal(t, protocol.DiagnosticSeverityWarning, result.Severity)
	assert.Equal(t, protocol.Position{Line: 1, Character: 12}, result.Range.Start)
	assert.Contains(t, result.Message, "[BREAKING]")
}

func TestCompatibilityDiagnosticsPlaceRemovedDeclarationAtDomain(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User {}\n"), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline", "contract.skel")

	documentURI := uri.File(path)
	document := workspace.BuildDocument(documentURI, path, "domain demo\n", 2)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})
	diagnostics, domains, err := SemanticWorkspace(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths, false)
	require.NoError(t, err)
	require.Empty(t, diagnostics)

	appendCompatibilityDiagnostics(t.Context(), sourcediff.New(), diagnostics, domains, sources, paths, CompatibilityOptions{Enabled: true})
	require.Len(t, diagnostics[documentURI], 1)
	result := diagnostics[documentURI][0]
	assert.Equal(t, protocol.String("schema.declaration.removed"), result.Code)
	assert.Equal(t, protocol.Position{Line: 0, Character: 7}, result.Range.Start)
}

func TestCompatibilityDiagnosticsReportExplicitBaselineFailure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	baselinePath := filepath.Join(root, "baseline.skel")
	require.NoError(t, os.WriteFile(baselinePath, []byte("domain demo\ndata User { id string }\n"), 0o600))
	documentURI := uri.File(path)
	document := workspace.BuildDocument(documentURI, path, "domain demo\ndata User { id: string }\n", 1)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})
	diagnostics, domains, err := SemanticWorkspace(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths, false)
	require.NoError(t, err)
	require.Empty(t, diagnostics)

	appendCompatibilityDiagnostics(t.Context(), sourcediff.New(), diagnostics, domains, sources, paths,
		CompatibilityOptions{Enabled: true, BaselineSkelIn: "baseline.skel"})
	require.Len(t, diagnostics[documentURI], 1)
	result := diagnostics[documentURI][0]
	assert.Equal(t, protocol.String("schema.baseline"), result.Code)
	assert.Contains(t, result.Message, baselinePath)
}
