// Package sourcediff coordinates compilation and cached source baselines for
// schema comparisons. Semantic comparison rules live in schema/diff.
package sourcediff

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/schema"
	"go.yorun.ai/skel/schema/diff"
)

// ErrSourceCompilation identifies a source input that could not be compiled
// while preparing a schema diff.
var ErrSourceCompilation = errors.New("schema source compilation failed")

// Option configures a source compatibility comparison.
type Option struct {
	// BaselineSkelIn selects an explicit baseline file or directory. An empty
	// value reads the domain's source directory from Git HEAD.
	BaselineSkelIn string
	// Strict rejects migration warnings in the candidate; historical baselines
	// retain compatibility with older declarations.
	Strict bool
}

type _CachedSourceBaseline struct {
	head           string
	repositoryRoot string
	domain         *schema.Domain
	err            error
}

type _CachedGitFailure struct {
	expires time.Time
	err     error
}

const gitFailureCacheDuration = 2 * time.Second

// Differ caches the current immutable Git baseline for each domain so
// repeated editor analysis does not re-read and recompile unchanged history.
type Differ struct {
	mu          sync.Mutex
	baselines   map[string]_CachedSourceBaseline
	gitFailures map[string]_CachedGitFailure
	now         func() time.Time
}

// New creates a source differ with an empty Git baseline cache.
func New() *Differ {
	return &Differ{
		baselines: map[string]_CachedSourceBaseline{}, gitFailures: map[string]_CachedGitFailure{}, now: time.Now,
	}
}

// DiffWorkspaceDomain compares one successfully analyzed in-memory domain with
// either an explicit source baseline or the same source directory at Git HEAD.
func DiffWorkspaceDomain(ctx context.Context, candidate compiler.WorkspaceDomain, option Option) (*diff.Report, error) {
	return New().DiffWorkspaceDomain(ctx, candidate, option)
}

// DiffSource compares a candidate file or directory with either an explicit
// source baseline or the same path at Git HEAD.
func DiffSource(ctx context.Context, candidateSkelIn string, option Option) (*diff.Report, error) {
	compiled, err := compiler.CompileImportContext(ctx, compiler.Option{SkelIn: candidateSkelIn, Strict: option.Strict})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %w", ErrSourceCompilation, err)
	}
	root, err := filepath.Abs(candidateSkelIn)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	candidate := compiler.WorkspaceDomain{Name: compiled.Domain.Name(), Root: root, Schema: compiled.Domain}
	if !info.IsDir() {
		candidate.Sources = []compiler.Source{{Path: root, Root: root}}
	} else {
		candidate.Sources = []compiler.Source{{Path: filepath.Join(root, "domain.skel"), Root: root, DirectoryInput: true}}
	}
	// CLI relative baseline paths are relative to the invocation directory.
	if strings.TrimSpace(option.BaselineSkelIn) != "" {
		option.BaselineSkelIn, err = filepath.Abs(option.BaselineSkelIn)
		if err != nil {
			return nil, err
		}
	}
	return New().DiffWorkspaceDomain(ctx, candidate, option)
}

// DiffWorkspaceDomain compares a domain while reusing its unchanged Git
// baseline across calls to the same differ.
func (d *Differ) DiffWorkspaceDomain(ctx context.Context, candidate compiler.WorkspaceDomain, option Option) (*diff.Report, error) {
	if option.Strict {
		diagnostics := compiler.MigrationDiagnostics(candidate.Schema)
		compiler.ApplyStrictMode(diagnostics)
		if diagnostics.HasErrors() {
			return nil, fmt.Errorf("%w: %w", ErrSourceCompilation, diagnostics)
		}
	}
	candidateSchema := candidate.Schema
	var err error
	baselineSkelIn := strings.TrimSpace(option.BaselineSkelIn)
	var baselineSchema *schema.Domain
	gitRepositoryRoot := ""
	if baselineSkelIn == "" {
		baselineSchema, gitRepositoryRoot, err = compileGitBaseline(ctx, d, candidate)
	} else {
		if !filepath.IsAbs(baselineSkelIn) {
			directory := candidate.Root
			if workspaceDomainIsFile(candidate) {
				directory = filepath.Dir(directory)
			}
			baselineSkelIn = filepath.Join(directory, baselineSkelIn)
		}
		baseline, compileErr := compiler.CompileImportContext(ctx, compiler.Option{SkelIn: baselineSkelIn})
		if compileErr != nil {
			if errors.Is(compileErr, context.Canceled) || errors.Is(compileErr, context.DeadlineExceeded) {
				return nil, compileErr
			}
			return nil, fmt.Errorf("%w: compile schema compatibility baseline %s: %w", ErrSourceCompilation, baselineSkelIn, compileErr)
		}
		baselineSchema = baseline.Domain
	}
	if err != nil {
		return nil, err
	}
	report, err := diff.Compare(baselineSchema, candidateSchema)
	if err != nil {
		return nil, err
	}
	if gitRepositoryRoot != "" {
		remapReportBaselinePositions(report, gitRepositoryRoot)
	}
	return report, nil
}

func (d *Differ) cachedBaseline(key, head string) (*schema.Domain, string, error, bool) {
	if d == nil {
		return nil, "", nil, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	baseline := d.baselines[key]
	if baseline.head != head {
		return nil, "", nil, false
	}
	return baseline.domain, baseline.repositoryRoot, baseline.err, true
}

func (d *Differ) storeBaseline(key, head, repositoryRoot string, domain *schema.Domain, err error) {
	if d == nil || (domain == nil && err == nil) {
		return
	}
	d.mu.Lock()
	if d.baselines == nil {
		d.baselines = map[string]_CachedSourceBaseline{}
	}
	d.baselines[key] = _CachedSourceBaseline{
		head: head, repositoryRoot: repositoryRoot, domain: domain, err: err,
	}
	d.mu.Unlock()
}

func (d *Differ) cachedGitFailure(root string) error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	failure, ok := d.gitFailures[root]
	if !ok {
		return nil
	}
	if !d.currentTime().Before(failure.expires) {
		delete(d.gitFailures, root)
		return nil
	}
	return failure.err
}

func (d *Differ) storeGitFailure(root string, err error) {
	if d == nil || err == nil {
		return
	}
	d.mu.Lock()
	if d.gitFailures == nil {
		d.gitFailures = map[string]_CachedGitFailure{}
	}
	d.gitFailures[root] = _CachedGitFailure{expires: d.currentTime().Add(gitFailureCacheDuration), err: err}
	d.mu.Unlock()
}

func (d *Differ) clearGitFailure(root string) {
	if d == nil {
		return
	}
	d.mu.Lock()
	delete(d.gitFailures, root)
	d.mu.Unlock()
}

func (d *Differ) currentTime() time.Time {
	if d != nil && d.now != nil {
		return d.now()
	}
	return time.Now()
}

func remapReportBaselinePositions(report *diff.Report, repositoryRoot string) {
	for _, change := range report.Changes {
		if change.Baseline == nil || change.Baseline.File == "" {
			continue
		}
		relative, err := filepath.Rel(repositoryRoot, change.Baseline.File)
		if err == nil && !pathEscapesRoot(relative) {
			change.Baseline.File = "HEAD:" + filepath.ToSlash(relative)
		}
	}
}

// workspaceDomainIsFile identifies a single-file compiler input without consulting disk.
func workspaceDomainIsFile(candidate compiler.WorkspaceDomain) bool {
	return len(candidate.Sources) == 1 && filepath.Clean(candidate.Root) == filepath.Clean(candidate.Sources[0].Path)
}

func remapDiagnosticBaseline(item *compiler.Diagnostic, root string) {
	remap := func(path string) string {
		if path == "" {
			return path
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || pathEscapesRoot(relative) {
			return path
		}
		return "HEAD:" + filepath.ToSlash(relative)
	}
	item.Position.File = remap(item.Position.File)
	item.Range.Start.File = remap(item.Range.Start.File)
	item.Range.End.File = remap(item.Range.End.File)
	for i := range item.Related {
		item.Related[i].Range.Start.File = remap(item.Related[i].Range.Start.File)
		item.Related[i].Range.End.File = remap(item.Related[i].Range.End.File)
	}
}
