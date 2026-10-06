package compiler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/internal/analyzer"
)

func TestAnalyzeWorkspaceValidatesCrossDomainTypes(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{
		{Path: "/workspace/user.skel", Content: []byte("domain demo.user\npub data User { id: int }\n")},
		{Path: "/workspace/order.skel", Content: []byte("domain demo.order\nimport demo.user as user\ndata Order { owner: user.Missing }\n")},
	})

	require.Len(t, diagnostics, 1)
	assert.Equal(t, DiagnosticCodeSemanticReference, diagnostics[0].Code)
	assert.Equal(t, "/workspace/order.skel", diagnostics[0].Position.File)
	assert.Equal(t, 3, diagnostics[0].Position.Line)
	assert.Contains(t, diagnostics[0].Message, "definition of user.Missing not found")
}

func TestAnalyzeWorkspaceMergesSameDomainFiles(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{
		{Path: "/workspace/data.skel", Content: []byte("domain demo.user\ndata User { id: int }\n")},
		{Path: "/workspace/service.skel", Content: []byte("domain demo.user\npub service UserService { method get { output User } }\n")},
	})

	assert.Empty(t, diagnostics)
}

func TestAnalyzeDomainsContextReturnsSuccessfulInMemoryDomains(t *testing.T) {
	analyzer := NewWorkspaceAnalyzer()
	diagnostics, domains, err := analyzer.AnalyzeDomainsContext(t.Context(), []Source{{
		Path: "/workspace/skel/data.skel", Root: "/workspace/skel",
		Content: []byte("domain demo.user\ndata User { id: int }\n"),
	}})

	require.NoError(t, err)
	require.Empty(t, diagnostics)
	require.Len(t, domains, 1)
	assert.Equal(t, "demo.user", domains[0].Name)
	assert.Equal(t, "/workspace/skel", domains[0].Root)
	assert.Equal(t, "demo.user", domains[0].Schema.Name())
	require.Len(t, domains[0].Sources, 1)
}

func TestAnalyzeWorkspaceKeepsSameNamedRootsIndependent(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{
		{
			Path: "/workspace/source/actor.skel", Root: "/workspace/source",
			Content: []byte("domain demo\nresource User { action read }\n"),
		},
		{
			Path: "/workspace/generated/types.skel", Root: "/workspace/generated",
			Content: []byte("domain demo\nresource User { action read }\n"),
		},
	})

	assert.Empty(t, diagnostics)
}

func TestAnalyzeWorkspaceDoesNotResolveImportsForRootedSources(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{
		{
			Path: "/workspace/source/user.skel", Root: "/workspace/source",
			Content: []byte("domain demo.user\npub data User {}\n"),
		},
		{
			Path: "/workspace/order/order.skel", Root: "/workspace/order",
			Content: []byte("domain demo.order\nimport demo.user as user\ndata Order { user: user.Missing }\n"),
		},
	})

	assert.Empty(t, diagnostics)
}

func TestAnalyzeWorkspaceReturnsStructuredDuplicatePosition(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{
		{Path: "/workspace/a.skel", Content: []byte("domain demo\ndata User {}\n")},
		{Path: "/workspace/b.skel", Content: []byte("domain demo\ndata User {}\n")},
	})

	require.Len(t, diagnostics, 1)
	assert.Equal(t, DiagnosticCodeSemanticDuplicate, diagnostics[0].Code)
	assert.Equal(t, "/workspace/b.skel", diagnostics[0].Position.File)
	assert.Equal(t, 2, diagnostics[0].Position.Line)
	assert.NotContains(t, diagnostics[0].Message, "/workspace/b.skel:2:6")
}

func TestAnalyzeWorkspaceCollectsMultipleSemanticDiagnosticsInOneDomain(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/data.skel",
		Content: []byte(`domain demo
data User { missing: MissingUser }
data Order { missing: MissingOrder }
`),
	}})

	require.Len(t, diagnostics, 2)
	assert.Equal(t, []int{2, 3}, []int{diagnostics[0].Position.Line, diagnostics[1].Position.Line})
	assert.Contains(t, diagnostics[0].Message, "definition of MissingUser not found")
	assert.Contains(t, diagnostics[1].Message, "definition of MissingOrder not found")
}

func TestAnalyzeWorkspaceSuggestsCorrectedIdentifierCase(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/data.skel", Content: []byte("domain demo\ndata user_profile {}\n"),
	}})

	require.Len(t, diagnostics, 1)
	assert.Equal(t, DiagnosticCodeSemanticNaming, diagnostics[0].Code)
	require.NotNil(t, diagnostics[0].Suggestion)
	assert.Equal(t, "replace with UserProfile", diagnostics[0].Suggestion.Message)
	assert.Equal(t, "UserProfile", diagnostics[0].Suggestion.Replacement)
	assert.True(t, diagnostics[0].Suggestion.Replace)
}

func TestAnalyzeWorkspaceSuppressesInvalidDeclarationCascades(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/data.skel",
		Content: []byte(`domain demo
data User {
    id: int
    id: string
}
data Order { owner: User }
data Product { missing: MissingProduct }
`),
	}})

	require.Len(t, diagnostics, 2)
	assert.Contains(t, diagnostics[0].Message, "duplicated DataMember id")
	assert.Contains(t, diagnostics[1].Message, "definition of MissingProduct not found")
	for _, diagnostic := range diagnostics {
		assert.NotContains(t, diagnostic.Message, "definition of User not found")
	}
}

func TestAnalyzeWorkspaceCollectsMultipleMemberDiagnostics(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/data.skel",
		Content: []byte(`domain demo
data User {
    first: MissingFirst
    second: MissingSecond
}
`),
	}})

	require.Len(t, diagnostics, 2)
	assert.Contains(t, diagnostics[0].Message, "definition of MissingFirst not found")
	assert.Contains(t, diagnostics[1].Message, "definition of MissingSecond not found")
}

func TestAnalyzeWorkspaceLimitsDiagnosticsPerDomain(t *testing.T) {
	var source strings.Builder
	source.WriteString("domain demo\n")
	for index := 0; index < analyzer.MaxDiagnosticsPerDomain+10; index++ {
		_, _ = fmt.Fprintf(&source, "data Type%d { missing: Missing%d }\n", index, index)
	}

	diagnostics := AnalyzeWorkspace([]Source{{Path: "/workspace/data.skel", Content: []byte(source.String())}})

	assert.Len(t, diagnostics, analyzer.MaxDiagnosticsPerDomain)
}

func TestAnalyzeWorkspaceSuppressesDependentCascade(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{
		{Path: "/workspace/user.skel", Domain: "demo.user", Content: []byte("domain demo.user\ndata User {")},
		{Path: "/workspace/order.skel", Content: []byte("domain demo.order\nimport demo.user as user\ndata Order { owner: user.User }\n")},
	})

	require.Len(t, diagnostics, 1)
	assert.Equal(t, DiagnosticCodeSyntaxEOF, diagnostics[0].Code)
	assert.Equal(t, "/workspace/user.skel", diagnostics[0].Position.File)
}

func TestAnalyzeWorkspaceReportsMissingImport(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/order.skel", Content: []byte("domain demo.order\nimport demo.user as user\ndata Order {}\n"),
	}})

	require.Len(t, diagnostics, 1)
	assert.Equal(t, DiagnosticCodeImportMissing, diagnostics[0].Code)
	assert.Equal(t, 2, diagnostics[0].Position.Line)
}

func TestAnalyzeWorkspaceReportsMultipleMissingImports(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path:    "/workspace/order.skel",
		Content: []byte("domain demo.order\nimport demo.user as user\nimport demo.product\ndata Order {}\n"),
	}})

	require.Len(t, diagnostics, 2)
	assert.Equal(t, []int{2, 3}, []int{diagnostics[0].Position.Line, diagnostics[1].Position.Line})
	assert.Equal(t, DiagnosticCodeImportMissing, diagnostics[0].Code)
	assert.Equal(t, DiagnosticCodeImportMissing, diagnostics[1].Code)
}

func TestWorkspaceHardCyclesUseRealEdges(t *testing.T) {
	for _, tt := range []struct {
		name, text string
		cyclic     bool
	}{
		{"mixed soft and hard", "data Alpha { beta: Beta }\ndata Beta { alpha: Alpha gamma: Gamma? }\ndata Gamma { beta: Beta? }", true},
		{"branching hard", "data Alpha { beta: Beta gamma: Gamma }\ndata Beta { alpha: Alpha }\ndata Gamma { alpha: Alpha }", true},
		{"nullable break", "data Alpha { beta: Beta }\ndata Beta { alpha: Alpha? }", false},
		{"self", "data Alpha { alpha: Alpha }", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for range 10 {
				diagnostics := AnalyzeWorkspace([]Source{{Path: "input.skel", Content: []byte("domain demo\n" + tt.text)}})
				if tt.cyclic && len(diagnostics) == 0 {
					t.Fatal("hard cycle was accepted")
				}
				if !tt.cyclic && len(diagnostics) != 0 {
					t.Fatalf("soft cycle rejected: %v", diagnostics)
				}
				if tt.cyclic && diagnostics[0].Message != "hard reference chain detected: Alpha -> Beta -> Alpha, try nullable/list/map instead" && tt.name != "self" {
					t.Fatalf("unexpected cycle path: %v", diagnostics)
				}
			}
		})
	}
}
