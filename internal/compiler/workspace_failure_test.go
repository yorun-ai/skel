package compiler

import (
	"testing"

	"github.com/stretchr/testify/require"
	textsource "go.yorun.ai/skel/internal/source"
)

func TestWorkspaceCachesFailuresAndInvalidatesReverseDependencies(t *testing.T) {
	engine := NewWorkspaceAnalyzer()
	inputs := []Source{
		{Path: "/shared.skel", Content: []byte("domain shared\npub data Item { value: Missing }\n")},
		{Path: "/consumer.skel", Content: []byte("domain consumer\nimport shared\ndata Box { value: shared.Item }\n")},
		{Path: "/other.skel", Content: []byte("domain other\ndata Value { id: int }\n")},
	}
	initial := engine.Analyze(inputs)
	require.Len(t, initial, 1)
	repeated := engine.Analyze(inputs)
	require.Equal(t, initial, repeated)
	require.Zero(t, engine.Stats().AnalyzedDomains)
	// An independent edit must still reuse the cached failed semantic outcome.
	inputs[2].Content = []byte("domain other\ndata Value { id: string }\n")
	require.Equal(t, initial, engine.Analyze(inputs))
	require.Equal(t, 1, engine.Stats().AnalyzedDomains)
	inputs[0].Content = []byte("domain shared\npub data Item { value: int }\n")
	require.Empty(t, engine.Analyze(inputs))
	require.Equal(t, 2, engine.Stats().AnalyzedDomains)
	// Removing an import invalidates consumers even though their text is unchanged.
	require.NotEmpty(t, engine.Analyze(inputs[1:]))
	require.Empty(t, engine.Analyze(inputs))
	require.Equal(t, 2, engine.Stats().AnalyzedDomains)
}

func TestWorkspaceFailureDiagnosticsAreOwnedByEachResult(t *testing.T) {
	engine := NewWorkspaceAnalyzer()
	inputs := []Source{{Path: "/invalid.skel", Content: []byte("domain demo\ndata Value { id: int id: string }\n")}}
	first := engine.Analyze(inputs)
	require.NotEmpty(t, first)
	require.NotEmpty(t, first[0].Related)
	original := first[0].Related[0].Message
	first[0].Related[0].Message = "mutated by caller"
	require.Equal(t, original, engine.Analyze(inputs)[0].Related[0].Message)
}

func TestWorkspaceRevisionIdentityDoesNotMixVirtualSources(t *testing.T) {
	engine := NewWorkspaceAnalyzer()
	text := "domain demo\ndata Value { member: Missing }\n"
	for _, id := range []textsource.ID{"untitled:first", "untitled:second"} {
		document := textsource.New(id, "", 1, text)
		diagnostics := engine.Analyze([]Source{{Path: string(id), Document: document}})
		require.NotEmpty(t, diagnostics)
		require.Equal(t, string(id), diagnostics[0].Position.File)
	}
}

func TestVersionOnlyChangeReusesAnalysisAndReturnsCurrentRevision(t *testing.T) {
	engine := NewWorkspaceAnalyzer()
	text := "domain demo\ndata Value {}\n"
	for _, version := range []int64{1, 2} {
		input := Source{Path: "/value.skel", Document: textsource.New("file:///value.skel", "/value.skel", version, text)}
		diagnostics, domains, err := engine.AnalyzeDomainsContext(t.Context(), []Source{input})
		require.NoError(t, err)
		require.Empty(t, diagnostics)
		require.Len(t, domains, 1)
		require.Equal(t, version, domains[0].Sources[0].Document.Version())
	}
	require.Zero(t, engine.Stats().AnalyzedDomains)
}

func TestReusedSyntaxDiagnosticsCannotBeMutatedThroughEarlierResults(t *testing.T) {
	engine := NewWorkspaceAnalyzer()
	inputs := []Source{
		{Path: "/syntax.skel", Content: []byte("domain demo\ndata Value { id string }\n")},
		{Path: "/independent.skel", Content: []byte("domain other\ndata Other {}\n")},
	}
	first := engine.Analyze(inputs)
	require.NotEmpty(t, first)
	require.NotNil(t, first[0].Suggestion)
	original := first[0].Suggestion.Message
	first[0].Suggestion.Message = "changed"
	inputs[1].Content = []byte("domain other\ndata Another {}\n")
	next := engine.Analyze(inputs)
	require.NotEmpty(t, next)
	require.NotNil(t, next[0].Suggestion)
	require.Equal(t, original, next[0].Suggestion.Message)
}
