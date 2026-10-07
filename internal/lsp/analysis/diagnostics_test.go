package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/lsp/workspace"
)

func TestSemanticStrictModePreservesCachedErrors(t *testing.T) {
	documentURI := uri.File("/workspace/order.skel")
	document := workspace.BuildDocument(documentURI, documentURI.FsPath(), "domain demo.order\nservice OrderService { method ping {} }\n", 1)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})
	analyzer := compiler.NewWorkspaceAnalyzer()
	for _, strict := range []bool{false, true, false} {
		diagnostics, domains, err := SemanticWorkspace(t.Context(), analyzer, sources, paths, strict)
		require.NoError(t, err)
		require.Empty(t, domains)
		require.Len(t, diagnostics[documentURI], 1)
		severity := protocol.DiagnosticSeverityError
		assert.Equal(t, severity, diagnostics[documentURI][0].Severity)
	}
}

func TestSemanticDiagnosticsDoNotResolveImportsAcrossDomainRoots(t *testing.T) {
	userURI := uri.File("/workspace/user/user.skel")
	orderURI := uri.File("/workspace/order/order.skel")
	documents := map[uri.URI]*workspace.Document{
		userURI:  workspace.BuildDocument(userURI, userURI.FsPath(), "domain demo.user\ndata User {}\n", 2),
		orderURI: workspace.BuildDocument(orderURI, orderURI.FsPath(), "domain demo.order\nimport demo.user as user\ndata Order { owner: user.Missing }\n", 7),
	}

	sources, paths := SemanticSources(documents)
	diagnostics, err := SemanticDiagnostics(context.Background(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	assert.Empty(t, diagnostics)
}

func TestSemanticDiagnosticsRejectImportAliasAtItsSource(t *testing.T) {
	documentURI := uri.File("/workspace/order.skel")
	document := workspace.BuildDocument(documentURI, documentURI.FsPath(), "domain app\nimport second as first\nimport first as a\n", 1)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})
	diagnostics, err := SemanticDiagnostics(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	require.Len(t, diagnostics[documentURI], 1)
	d := diagnostics[documentURI][0]
	assert.Equal(t, protocol.String(compiler.DiagnosticCodeSemanticDuplicate), d.Code)
	assert.Equal(t, protocol.Position{Line: 1, Character: 17}, d.Range.Start)
	assert.Equal(t, protocol.Position{Line: 1, Character: 22}, d.Range.End)
	require.Len(t, d.RelatedInformation, 1)
	assert.Equal(t, protocol.Position{Line: 2, Character: 7}, d.RelatedInformation[0].Location.Range.Start)
}

func TestSemanticDiagnosticsDoNotDuplicateSyntaxErrors(t *testing.T) {
	documentURI := uri.File("/workspace/user.skel")
	document := workspace.BuildDocument(documentURI, "/workspace/user.skel", "domain demo.user\ndata User {", 2)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})

	diagnostics, err := SemanticDiagnostics(context.Background(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	assert.Empty(t, diagnostics)
}

func TestSemanticDiagnosticsPublishMultipleErrorsForOneDocument(t *testing.T) {
	documentURI := uri.File("/workspace/data.skel")
	document := workspace.BuildDocument(documentURI, "/workspace/data.skel", `domain demo
data User { missing: MissingUser }
data Order { missing: MissingOrder }
`, 3)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})

	diagnostics, err := SemanticDiagnostics(context.Background(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)

	require.Len(t, diagnostics[documentURI], 2)
	assert.Contains(t, diagnostics[documentURI][0].Message, "MissingUser")
	assert.Contains(t, diagnostics[documentURI][1].Message, "MissingOrder")
}

func TestSemanticDiagnosticsIncludeDuplicateRelatedLocation(t *testing.T) {
	documentURI := uri.File("/workspace/data.skel")
	document := workspace.BuildDocument(documentURI, "/workspace/data.skel", "domain demo\ndata User {}\ndata User {}\n", 1)
	sources, paths := SemanticSources(map[uri.URI]*workspace.Document{documentURI: document})

	diagnostics, err := SemanticDiagnostics(context.Background(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	require.Len(t, diagnostics[documentURI], 1)
	diagnostic := diagnostics[documentURI][0]
	assert.Equal(t, protocol.String(compiler.DiagnosticCodeSemanticDuplicate), diagnostic.Code)
	require.Len(t, diagnostic.RelatedInformation, 1)
	assert.Equal(t, protocol.Position{Line: 1, Character: 5}, diagnostic.RelatedInformation[0].Location.Range.Start)
}

func TestSemanticDiagnosticsKeepSameNamedDomainDirectoriesIndependent(t *testing.T) {
	sourceURI := uri.File("/workspace/domain/base/skel/actor.skel")
	generatedURI := uri.File("/workspace/domain/base/pub/skeled/skel/types.skel")
	documents := map[uri.URI]*workspace.Document{
		sourceURI: workspace.BuildDocument(
			sourceURI,
			sourceURI.FsPath(),
			"domain base\npub resource User { action read }\n",
			1,
		),
		generatedURI: workspace.BuildDocument(
			generatedURI,
			generatedURI.FsPath(),
			"domain base\npub resource User { action read }\n",
			1,
		),
	}

	sources, paths := SemanticSources(documents)
	diagnostics, err := SemanticDiagnostics(context.Background(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	assert.Empty(t, diagnostics)
}

func TestSemanticDiagnosticsMergeSameNamedDomainFilesWithDomainFile(t *testing.T) {
	firstURI := uri.File("/workspace/domain/base/skel/first.skel")
	secondURI := uri.File("/workspace/domain/base/skel/second.skel")
	documents := map[uri.URI]*workspace.Document{
		firstURI: workspace.BuildDocument(
			firstURI,
			firstURI.FsPath(),
			"domain base\ndata User {}\n",
			1,
		),
		secondURI: workspace.BuildDocument(
			secondURI,
			secondURI.FsPath(),
			"domain base\ndata User {}\n",
			1,
		),
	}

	domainURI := uri.File("/workspace/domain/base/skel/domain.skel")
	documents[domainURI] = workspace.BuildDocument(domainURI, domainURI.FsPath(), "domain base\n", 1)

	sources, paths := SemanticSources(documents)
	diagnostics, err := SemanticDiagnostics(context.Background(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	require.Len(t, diagnostics[secondURI], 1)
	assert.Equal(t, protocol.String(compiler.DiagnosticCodeSemanticDuplicate), diagnostics[secondURI][0].Code)
}

func TestSemanticDiagnosticsKeepStandaloneFormatterFixturesIndependent(t *testing.T) {
	documents := map[uri.URI]*workspace.Document{}
	for _, name := range []string{"complete.input.skel", "complete.golden.skel"} {
		path, err := filepath.Abs(filepath.Join("../../formatter/testdata", name))
		require.NoError(t, err)
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		documentURI := uri.File(path)
		documents[documentURI] = workspace.BuildDocument(documentURI, path, string(content), 1)
	}
	sources, paths := SemanticSources(documents)
	diagnostics, domains, err := SemanticWorkspace(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths, true)
	require.NoError(t, err)
	// Both fixtures declare the same demo.user declarations, so analyzing them
	// as one workspace group would report duplicates instead of staying quiet.
	require.Empty(t, diagnostics)
	require.Len(t, domains, 2)
	for _, domain := range domains {
		require.Len(t, domain.Sources, 1)
		assert.Equal(t, domain.Sources[0].Path, domain.Root)
		assert.Equal(t, "demo.user", domain.Name)
	}
}

func TestSemanticDiagnosticsChangeGroupingWithDomainFile(t *testing.T) {
	firstURI := uri.File("/workspace/first.skel")
	secondURI := uri.File("/workspace/second.skel")
	domainURI := uri.File("/workspace/domain.skel")
	documents := map[uri.URI]*workspace.Document{
		firstURI:  workspace.BuildDocument(firstURI, firstURI.FsPath(), "domain demo\ndata User {}\n", 1),
		secondURI: workspace.BuildDocument(secondURI, secondURI.FsPath(), "domain demo\ndata Order { user: User }\n", 1),
	}
	analyzer := compiler.NewWorkspaceAnalyzer()
	for _, hasDomainFile := range []bool{false, true, false} {
		if hasDomainFile {
			documents[domainURI] = workspace.BuildDocument(domainURI, domainURI.FsPath(), "domain demo\n", 1)
		} else {
			delete(documents, domainURI)
		}
		sources, paths := SemanticSources(documents)
		diagnostics, err := SemanticDiagnostics(t.Context(), analyzer, sources, paths)
		require.NoError(t, err)
		if hasDomainFile {
			assert.Empty(t, diagnostics)
		} else {
			require.Len(t, diagnostics[secondURI], 1)
			assert.Contains(t, diagnostics[secondURI][0].Message, "User")
		}
	}
}

func TestSemanticDiagnosticsRecoverAfterPartialGenerationNotifications(t *testing.T) {
	for _, recovery := range []string{"watched file", "opened file"} {
		t.Run(recovery, func(t *testing.T) {
			root := t.TempDir()
			serviceURI := uri.File(filepath.Join(root, "service.skel"))
			service := "domain demo\ndata Response { status: Status }\n"
			require.NoError(t, os.WriteFile(serviceURI.FsPath(), []byte(service), 0o600))
			store := workspace.New()
			store.AddRoot(uri.File(root))
			analyzer := compiler.NewWorkspaceAnalyzer()
			sources, paths := SemanticSources(store.Snapshot().DocumentsMap())
			diagnostics, err := SemanticDiagnostics(t.Context(), analyzer, sources, paths)
			require.NoError(t, err)
			require.Len(t, diagnostics[serviceURI], 1)
			require.Contains(t, diagnostics[serviceURI][0].Message, "definition of Status not found")
			// No notifications arrive for either of these generated siblings.
			require.NoError(t, os.WriteFile(filepath.Join(root, "domain.skel"), []byte("domain demo\n"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(root, "types.skel"), []byte("domain demo\ndata Status {}\n"), 0o600))
			if recovery == "watched file" {
				store.ApplyFileChanges([]protocol.FileEvent{{URI: serviceURI, Type: protocol.FileChangeTypeChanged}})
			} else {
				store.Put(serviceURI, service, 1, true)
				store.RefreshDirectory(serviceURI)
			}
			sources, paths = SemanticSources(store.Snapshot().DocumentsMap())
			diagnostics, err = SemanticDiagnostics(t.Context(), analyzer, sources, paths)
			require.NoError(t, err)
			assert.Empty(t, diagnostics)
		})
	}
}

func TestSemanticSourcesPreserveRemoteAuthority(t *testing.T) {
	docs := map[uri.URI]*workspace.Document{}
	for _, raw := range []string{"vscode-remote://ssh-remote+a/workspace/input.skel", "vscode-remote://ssh-remote+b/workspace/input.skel", "file:///workspace/input.skel"} {
		u := uri.URI(raw)
		docs[u] = workspace.BuildDocument(u, u.FsPath(), "domain demo\ndata User { value: Missing }\n", 1)
	}
	sources, paths := SemanticSources(docs)
	require.Len(t, paths, 3)
	diagnostics, err := SemanticDiagnostics(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths)
	require.NoError(t, err)
	require.Len(t, diagnostics, 3)
	for u := range docs {
		require.Len(t, diagnostics[u], 1)
		assert.Equal(t, protocol.String(compiler.DiagnosticCodeSemanticReference), diagnostics[u][0].Code)
	}
}

func TestSemanticDirectorySourceRules(t *testing.T) {
	for _, tt := range []struct{ name, domain, body, code string }{
		{"missing header", "domain demo\n", "data User {}\n", compiler.DiagnosticCodeDomainMissing},
		{"mismatch", "domain demo\n", "domain wrong\ndata User {}\n", compiler.DiagnosticCodeDomainMismatch},
		{"domain entries", "domain demo\ndata User {}\n", "domain demo\n", compiler.DiagnosticCodeDomainFileContent},
		{"sibling decorator", "domain demo\n", "@desc(\"wrong\")\ndomain demo\ndata User {}\n", compiler.DiagnosticCodeDomainDecorator},
	} {
		t.Run(tt.name, func(t *testing.T) {
			docs := map[uri.URI]*workspace.Document{}
			for name, text := range map[string]string{"domain.skel": tt.domain, "user.skel": tt.body} {
				u := uri.File("/workspace/" + name)
				docs[u] = workspace.BuildDocument(u, u.FsPath(), text, 1)
			}
			sources, paths := SemanticSources(docs)
			diagnostics, domains, err := SemanticWorkspace(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths, false)
			require.NoError(t, err)
			assert.Empty(t, domains)
			found := false
			for _, items := range diagnostics {
				for _, d := range items {
					if d.Code == protocol.String(tt.code) {
						found = true
					}
				}
			}
			assert.True(t, found, "diagnostics: %v", diagnostics)
		})
	}
}

func TestRemoteDomainFileDoesNotGroupOtherAuthorities(t *testing.T) {
	docs := map[uri.URI]*workspace.Document{}
	for raw, text := range map[string]string{
		"vscode-remote://ssh-remote+a/workspace/domain.skel": "domain demo\n",
		"vscode-remote://ssh-remote+a/workspace/user.skel":   "domain demo\ndata User {}\n",
		"vscode-remote://ssh-remote+b/workspace/user.skel":   "domain other\ndata User {}\n",
	} {
		u := uri.URI(raw)
		docs[u] = workspace.BuildDocument(u, u.FsPath(), text, 1)
	}
	sources, paths := SemanticSources(docs)
	diagnostics, domains, err := SemanticWorkspace(t.Context(), compiler.NewWorkspaceAnalyzer(), sources, paths, false)
	require.NoError(t, err)
	assert.Empty(t, diagnostics)
	require.Len(t, domains, 2)
	for _, domain := range domains {
		physical := FilesystemDomain(domain)
		assert.NotContains(t, physical.Root, "://")
		for _, src := range physical.Sources {
			assert.NotContains(t, src.Path, "://")
		}
	}
}
