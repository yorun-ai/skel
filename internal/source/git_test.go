package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yorun.ai/skelc/internal/testutil"
)

func TestGitProviderPinsRevisionAndPreservesFileNames(t *testing.T) {
	testutil.RequireGit(t)
	root := t.TempDir()
	testutil.InitRepository(t, root)
	name := "用户\t\"\n.skel"
	path := filepath.Join(root, name)
	require.NoError(t, os.WriteFile(path, []byte("domain before\n"), 0o600))
	testutil.Commit(t, root, "before")
	head, err := GitBytes(t.Context(), root, "rev-parse", "HEAD")
	require.NoError(t, err)
	provider, err := NewGit(t.Context(), root, strings.TrimSpace(string(head)), root)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("domain after\n"), 0o600))
	testutil.Commit(t, root, "after")
	entries, err := provider.ReadDir(t.Context(), root)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, name, entries[0].Name)
	document, err := provider.Read(t.Context(), path)
	require.NoError(t, err)
	require.Equal(t, "domain before\n", document.Text())
	require.Equal(t, path, document.Path())
	require.Contains(t, string(document.ID()), strings.TrimSpace(string(head)))
}
