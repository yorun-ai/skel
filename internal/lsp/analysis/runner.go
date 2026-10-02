package analysis

import (
	"context"
	"sync"
	"time"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/lsp/workspace"
	"go.yorun.ai/skelc/internal/schema/sourcediff"
)

// Result is the semantic diagnostics produced for one workspace revision.
type Result struct {
	Revision    uint64
	Generation  uint64
	Diagnostics map[uri.URI][]protocol.Diagnostic
}

// CompatibilityOptions controls continuous schema compatibility diagnostics.
type CompatibilityOptions struct {
	Enabled           bool
	IncludeCompatible bool
	BaselineSkelIn    string
}

// Options controls semantic and compatibility diagnostics.
type Options struct {
	Strict        bool
	Compatibility CompatibilityOptions
}

// Runner debounces workspace analysis and cancels superseded work.
type Runner struct {
	mu                sync.Mutex
	delay             time.Duration
	generation        uint64
	timer             *time.Timer
	cancel            context.CancelFunc
	workspaceAnalyzer *compiler.WorkspaceAnalyzer
	compatibility     *sourcediff.Differ
}

// NewRunner creates a semantic analysis runner.
func NewRunner(delay time.Duration) *Runner {
	return &Runner{
		delay: delay, workspaceAnalyzer: compiler.NewWorkspaceAnalyzer(), compatibility: sourcediff.New(),
	}
}

// Schedule replaces pending analysis with analysis of snapshot.
func (r *Runner) Schedule(snapshot workspace.Snapshot, option Options, accept func(Result)) {
	r.mu.Lock()
	r.generation++
	generation := r.generation
	if r.timer != nil {
		r.timer.Stop()
	}
	if r.cancel != nil {
		r.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.timer = time.AfterFunc(r.delay, func() {
		r.run(ctx, generation, snapshot, option, accept)
	})
	r.mu.Unlock()
}

// Stop cancels pending and active analysis.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generation++
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
}

func (r *Runner) run(ctx context.Context, generation uint64, snapshot workspace.Snapshot, option Options, accept func(Result)) {
	if ctx.Err() != nil {
		return
	}
	sources, paths := SemanticSources(snapshot.DocumentsMap())
	diagnostics, domains, err := SemanticWorkspace(ctx, r.workspaceAnalyzer, sources, paths, option.Strict)
	if err != nil {
		return
	}
	if option.Compatibility.Enabled {
		appendCompatibilityDiagnostics(ctx, r.compatibility, diagnostics, domains, sources, paths, option.Compatibility)
	}
	r.mu.Lock()
	if generation != r.generation || ctx.Err() != nil {
		r.mu.Unlock()
		return
	}
	r.timer = nil
	r.cancel = nil
	r.mu.Unlock()
	accept(Result{Generation: generation, Revision: snapshot.Revision(), Diagnostics: diagnostics})
}

// IsCurrent also rejects results superseded after the runner released its lock.
func (r *Runner) IsCurrent(result Result) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return result.Generation == 0 || result.Generation == r.generation
}
