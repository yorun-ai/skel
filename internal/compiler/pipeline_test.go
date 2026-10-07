package compiler

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/source"
)

func TestPipelineMatchesDiskAndMemoryProviders(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"domain.skel": "domain demo\n", "types.skel": "domain demo\npub data Value { id: uuid }\n"}
	revisions := []*source.Document{}
	for name, text := range files {
		path := filepath.Join(root, name)
		writeFile(t, path, text)
		revisions = append(revisions, source.New(source.ID("memory:"+name), path, 1, text))
	}
	disk, err := Compile(Option{SkelIn: root})
	require.NoError(t, err)
	memory := loader.NewMemory(revisions...)
	frozen, err := CompileImportFrom(t.Context(), memory, Option{SkelIn: root})
	require.NoError(t, err)
	require.Equal(t, disk.Domain.Hash(), frozen.Domain.Hash())
	loaded, err := loader.LoadFrom(t.Context(), memory, root)
	require.NoError(t, err)
	inputs, err := prepareInput(t.Context(), loaded, true)
	require.NoError(t, err)
	diagnostics, domains, err := NewWorkspaceAnalyzer().AnalyzeDomainsContext(t.Context(), inputs)
	require.NoError(t, err)
	require.Empty(t, diagnostics)
	require.Len(t, domains, 1)
	require.Equal(t, disk.Domain.Hash(), domains[0].Schema.Hash())
}

func TestPipelineOptionsSeparateUnresolvedImportsAndRejectPartialSchemas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.skel")
	writeFile(t, path, "domain demo\nimport shared\ndata Value { other: shared.Value }\n")
	_, err := Compile(Option{SkelIn: path})
	require.Error(t, err)
	_, err = CompileImport(Option{SkelIn: path})
	require.NoError(t, err)
	checked, err := Check(Option{SkelIn: path})
	require.NoError(t, err)
	require.Empty(t, checked.Diagnostics)
	engine := NewWorkspaceAnalyzer()
	inputs := []Source{{Path: path, Content: []byte("domain demo\ndata Valid {}\ndata Broken { id int }\n")}}
	diagnostics, domains, err := engine.AnalyzeDomainsContext(t.Context(), inputs)
	require.NoError(t, err)
	require.NotEmpty(t, diagnostics)
	require.Empty(t, domains)
}
