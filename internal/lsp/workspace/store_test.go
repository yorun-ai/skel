package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestStorePreservesRemoteWorkspaceURIsAndOpenContents(t *testing.T) {
	rootPath := t.TempDir()
	userPath := filepath.Join(rootPath, "user.skel")
	require.NoError(t, os.WriteFile(userPath, []byte("domain demo.user\ndata User {}\n"), 0o600))
	rootURI, err := uri.From(uri.Components{
		Scheme: "vscode-remote", Authority: "ssh-remote+test", Path: filepath.ToSlash(rootPath),
	})
	require.NoError(t, err)
	userURI, err := uri.JoinPath(rootURI, "user.skel")
	require.NoError(t, err)

	store := New()
	store.AddRoot(rootURI)
	snapshot := store.Snapshot()
	require.NotNil(t, snapshot.Document(userURI))
	assert.Nil(t, snapshot.Document(uri.File(userPath)))

	store.Put(userURI, "domain demo.user\ndata Profile {}\n", 1, true)
	snapshot = store.Snapshot()
	require.Len(t, snapshot.Documents(), 1)
	assert.Equal(t, int32(1), snapshot.Document(userURI).Version)
	assert.Equal(t, "Profile", snapshot.Document(userURI).Definitions[0].Name)
}

func TestStoreAddsAndRemovesWorkspaceRoots(t *testing.T) {
	firstPath := t.TempDir()
	secondPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(firstPath, "first.skel"), []byte("domain demo.first\ndata First {}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(secondPath, "second.skel"), []byte("domain demo.second\ndata Second {}\n"), 0o600))
	firstURI := uri.File(firstPath)
	secondURI := uri.File(secondPath)
	firstDocumentURI := uri.File(filepath.Join(firstPath, "first.skel"))
	secondDocumentURI := uri.File(filepath.Join(secondPath, "second.skel"))

	store := New()
	store.AddRoot(firstURI)
	require.NotNil(t, store.Snapshot().Document(firstDocumentURI))

	removed := store.RemoveRoot(firstURI)
	store.AddRoot(secondURI)
	assert.Equal(t, []uri.URI{firstDocumentURI}, removed)
	snapshot := store.Snapshot()
	assert.Nil(t, snapshot.Document(firstDocumentURI))
	assert.NotNil(t, snapshot.Document(secondDocumentURI))
}

func TestStoreReconcilesGeneratedDirectoryFromPartialFileEvents(t *testing.T) {
	root := t.TempDir()
	store := New()
	store.AddRoot(uri.File(root))
	domainURI := uri.File(filepath.Join(root, "domain.skel"))
	typesURI := uri.File(filepath.Join(root, "types.skel"))
	serviceURI := uri.File(filepath.Join(root, "service.skel"))
	for name, content := range map[string]string{
		"domain.skel":  "domain demo\n",
		"types.skel":   "domain demo\ndata Status {}\n",
		"service.skel": "domain demo\ndata Response { status: Status }\n",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), 0o600))
	}
	// A generator can produce the entire directory before the client delivers
	// its first event; sibling events need not all arrive.
	store.ApplyFileChanges([]protocol.FileEvent{{URI: serviceURI, Type: protocol.FileChangeTypeCreated}})
	require.NotNil(t, store.Snapshot().Document(domainURI))
	require.NotNil(t, store.Snapshot().Document(typesURI))
	// A delayed deletion event must not remove a file already recreated on disk.
	store.ApplyFileChanges([]protocol.FileEvent{{URI: typesURI, Type: protocol.FileChangeTypeDeleted}})
	require.NotNil(t, store.Snapshot().Document(typesURI))
	// Preserve unsaved buffers while reconciling both changed and deleted siblings.
	store.Put(serviceURI, "domain demo\ndata Unsaved {}\n", 4, true)
	require.NoError(t, os.Remove(typesURI.FsPath()))
	changed := store.ApplyFileChanges([]protocol.FileEvent{{URI: serviceURI, Type: protocol.FileChangeTypeChanged}})
	assert.Contains(t, changed, typesURI)
	assert.Nil(t, store.Snapshot().Document(typesURI))
	assert.Equal(t, "Unsaved", store.Snapshot().Document(serviceURI).Definitions[0].Name)
	assert.Equal(t, int32(4), store.Snapshot().Document(serviceURI).Version)
}

func TestStoreRefreshDirectoryPreservesRemoteURIs(t *testing.T) {
	root := t.TempDir()
	documentURI, err := uri.From(uri.Components{Scheme: "vscode-remote", Authority: "ssh-remote+test", Path: filepath.ToSlash(filepath.Join(root, "service.skel"))})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "types.skel"), []byte("domain demo\ndata Status {}\n"), 0o600))
	store := New()
	store.Put(documentURI, "domain demo\ndata Response { status: Status }\n", 1, true)
	changed := store.RefreshDirectory(documentURI)
	require.Len(t, changed, 1)
	assert.Equal(t, "vscode-remote", changed[0].Scheme())
	assert.Equal(t, "ssh-remote+test", changed[0].Authority())
	require.Len(t, store.Snapshot().Documents(), 2)
	revision := store.Snapshot().Revision()
	assert.Empty(t, store.RefreshDirectory(documentURI))
	assert.Equal(t, revision, store.Snapshot().Revision())
	assert.Empty(t, store.RefreshDirectory(uri.URI("untitled:Untitled-1")))
}

func TestSnapshotReusesIndexesAndPreservesOldRevision(t *testing.T) {
	store := New()
	documentURI := uri.File("/audit/input.skel")
	store.Put(documentURI, "domain demo\ndata User {}\n", 1, true)
	before := store.Snapshot()
	require.NotNil(t, store.cached)
	cached := store.cached
	assert.Equal(t, before.Revision(), store.Snapshot().Revision())
	assert.Same(t, cached, store.cached)
	store.Put(documentURI, "domain demo\ndata Account {}\n", 2, true)
	assert.Nil(t, store.cached)
	after := store.Snapshot()
	assert.Equal(t, "User", before.Document(documentURI).Definitions[0].Name)
	assert.Equal(t, "Account", after.Document(documentURI).Definitions[0].Name)
	assert.Equal(t, after.Revision(), store.Revision())
	assert.Same(t, after.Document(documentURI), store.Document(documentURI))
	store.Put(documentURI, "domain demo\ndata Stale {}\n", 1, true)
	assert.Equal(t, after.Revision(), store.Revision())
	assert.Equal(t, "Account", store.Document(documentURI).Definitions[0].Name)
}

func TestConcurrentSnapshotsAndUpdatesPreserveNewestDocument(t *testing.T) {
	store := New()
	documentURI := uri.File("/audit/input.skel")
	store.Put(documentURI, "domain demo\ndata Initial {}\n", 1, true)
	var group sync.WaitGroup
	for worker := range 4 {
		group.Go(func() {
			for version := int32(2 + worker); version < 100; version += 4 {
				store.Put(documentURI, fmt.Sprintf("domain demo\ndata Value%d {}\n", version), version, true)
				snapshot := store.Snapshot()
				document := snapshot.Document(documentURI)
				if document == nil || len(document.Definitions) != 1 {
					t.Error("invalid snapshot")
				}
			}
		})
	}
	group.Wait()
	assert.Equal(t, int32(99), store.Document(documentURI).Version)
	assert.Equal(t, "Value99", store.Snapshot().Document(documentURI).Definitions[0].Name)
}
