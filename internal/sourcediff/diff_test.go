package sourcediff

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/testutil"
	"go.yorun.ai/skel/schema"
)

func TestDiffWorkspaceDomainUsesInMemoryCandidateAndGitHeadBaseline(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo with space")
	require.NoError(t, os.MkdirAll(root, 0o700))
	path := filepath.Join(root, "contract.skel")
	baseline := "domain demo\ndata User { id: int }\n"
	require.NoError(t, os.WriteFile(path, []byte(baseline), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline", "contract.skel")

	candidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: string }\n")
	differ := New()
	report, err := differ.DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Equal(t, schema.ImpactBreaking, report.Changes[0].Impact)
	assert.Equal(t, "data.member.type.changed", report.Changes[0].Code)
	require.NotNil(t, report.Changes[0].Baseline)
	assert.Equal(t, "HEAD:contract.skel", report.Changes[0].Baseline.File)
	require.NotNil(t, report.Changes[0].Candidate)
	assert.Equal(t, path, report.Changes[0].Candidate.File)

	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: string }\n"), 0o600))
	testutil.Commit(t, root, "new baseline", "contract.skel")
	nextCandidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: bool }\n")
	nextReport, err := differ.DiffWorkspaceDomain(t.Context(), nextCandidate, Option{})
	require.NoError(t, err)
	require.Len(t, nextReport.Changes, 1)
	assert.Contains(t, nextReport.Changes[0].Message, "string to bool")
	assert.Len(t, differ.baselines, 1)
}

func TestDiffWorkspaceDomainReportsUnavailableGitHistory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	candidate := workspaceDomain(t, root, path, "domain demo\ndata User {}\n")

	_, err := DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.ErrorIs(t, err, ErrGitHistoryUnavailable)
}

func TestDiffWorkspaceDomainCachesUnavailableGitHistoryTemporarily(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	candidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: string }\n")
	current := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)
	differ := New()
	differ.now = func() time.Time { return current }

	_, err := differ.DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.ErrorIs(t, err, ErrGitHistoryUnavailable)
	assert.Len(t, differ.gitFailures, 1)

	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int }\n"), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline", "contract.skel")
	_, err = differ.DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.ErrorIs(t, err, ErrGitHistoryUnavailable)

	current = current.Add(gitFailureCacheDuration)
	report, err := differ.DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Empty(t, differ.gitFailures)
}

func TestDiffWorkspaceDomainInvalidatesCachedFailureWhenHeadChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int id: string }\n"), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "invalid baseline", "contract.skel")
	candidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: string }\n")
	differ := New()

	_, err := differ.DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compile Git HEAD schema compatibility baseline")
	assert.Len(t, differ.baselines, 1)

	require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int }\n"), 0o600))
	testutil.Commit(t, root, "valid baseline", "contract.skel")
	report, err := differ.DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Len(t, differ.baselines, 1)
}

func TestDiffWorkspaceDomainSelectsDomainAcrossMultipleFiles(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"user.skel":    "domain demo.user\ndata User { id: int }\n",
		"profile.skel": "domain demo.user\ndata Profile { name: string }\n",
		"order.skel":   "domain demo.order\ndata Order { id: int }\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), 0o600))
	}
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "baseline")

	sources := []compiler.Source{}
	for name, content := range files {
		if name == "profile.skel" {
			content = "domain demo.user\ndata Profile { name: bool }\n"
		}
		sources = append(sources, compiler.Source{Path: filepath.Join(root, name), Root: root, Content: []byte(content)})
	}
	domains := workspaceDomains(t, sources)
	var candidate compiler.WorkspaceDomain
	for _, domain := range domains {
		if domain.Name == "demo.user" {
			candidate = domain
		}
	}
	require.NotNil(t, candidate.Model)

	report, err := DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Equal(t, "data.member.type.changed", report.Changes[0].Code)
	require.NotNil(t, report.Changes[0].Baseline)
	assert.Equal(t, "HEAD:profile.skel", report.Changes[0].Baseline.File)
	require.NotNil(t, report.Changes[0].Candidate)
	assert.Equal(t, filepath.Join(root, "profile.skel"), report.Changes[0].Candidate.File)
}

func TestDiffWorkspaceDomainReportsExplicitBaselineCompilePath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.skel")
	baselinePath := filepath.Join(root, "invalid-baseline.skel")
	require.NoError(t, os.WriteFile(baselinePath, []byte("domain demo\ndata User { id string }\n"), 0o600))
	candidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: string }\n")

	_, err := DiffWorkspaceDomain(t.Context(), candidate, Option{BaselineSkelIn: baselinePath})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compile schema compatibility baseline "+baselinePath)
	assert.NotContains(t, err.Error(), "skelc-schema-baseline-")
}

func TestDiffWorkspaceDomainResolvesExplicitBaselineFromSourceDirectory(t *testing.T) {
	workspace := t.TempDir()
	root := filepath.Join(workspace, "current")
	baselineRoot := filepath.Join(workspace, "baseline")
	require.NoError(t, os.MkdirAll(root, 0o700))
	require.NoError(t, os.MkdirAll(baselineRoot, 0o700))
	path := filepath.Join(root, "contract.skel")
	baselinePath := filepath.Join(baselineRoot, "contract.skel")
	require.NoError(t, os.WriteFile(baselinePath, []byte("domain demo\ndata User { id: int }\n"), 0o600))
	candidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: string }\n")

	report, err := DiffWorkspaceDomain(t.Context(), candidate, Option{BaselineSkelIn: "../baseline/contract.skel"})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Equal(t, "data.member.type.changed", report.Changes[0].Code)
}

func workspaceDomain(t *testing.T, root, path, content string) compiler.WorkspaceDomain {
	t.Helper()
	return workspaceDomains(t, []compiler.Source{{Path: path, Root: root, Content: []byte(content)}})[0]
}

func workspaceDomains(t *testing.T, sources []compiler.Source) []compiler.WorkspaceDomain {
	t.Helper()
	analyzer := compiler.NewWorkspaceAnalyzer()
	diagnostics, domains, err := analyzer.AnalyzeDomainsContext(t.Context(), sources)
	require.NoError(t, err)
	require.Empty(t, diagnostics)
	require.NotEmpty(t, domains)
	return domains
}

func TestGitBaselinePreservesQuotedFileNames(t *testing.T) {
	for _, name := range []string{"用户.skel", " leading.skel", "quoted\"name.skel", "line\nbreak.skel", "tab\tname.skel"} {
		t.Run(name, func(t *testing.T) {
			if runtime.GOOS == "windows" && strings.ContainsAny(name, "\"\n\t") {
				t.Skip("unsupported Windows filename")
			}
			root := t.TempDir()
			path := filepath.Join(root, name)
			require.NoError(t, os.WriteFile(path, []byte("domain demo\ndata User { id: int }\n"), 0600))
			testutil.InitRepository(t, root)
			_, err := gitOutput(t.Context(), root, "config", "core.quotePath", "true")
			require.NoError(t, err)
			testutil.Commit(t, root, "baseline", name)
			candidate := workspaceDomain(t, root, path, "domain demo\ndata User { id: string }\n")
			report, err := New().DiffWorkspaceDomain(t.Context(), candidate, Option{})
			require.NoError(t, err)
			require.Len(t, report.Changes, 1)
			assert.Equal(t, schema.ImpactBreaking, report.Changes[0].Impact)
		})
	}
}

func TestGitBaselineIgnoresInvalidIndependentDomain(t *testing.T) {
	testutil.RequireGit(t)
	root := t.TempDir()
	userPath := filepath.Join(root, "user.skel")
	require.NoError(t, os.WriteFile(userPath, []byte("domain demo.user\ndata User { id: int }\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "other.skel"), []byte("domain other\ndata Broken { value: Missing }\n"), 0o600))
	testutil.InitRepository(t, root)
	testutil.Commit(t, root, "independent baseline domains")
	candidate := workspaceDomains(t, []compiler.Source{{Path: userPath, Root: root, Content: []byte("domain demo.user\ndata User { id: string }\n")}})[0]
	report, err := DiffWorkspaceDomain(t.Context(), candidate, Option{})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
}
