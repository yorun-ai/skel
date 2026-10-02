package compiler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

// Source is an in-memory Skel document used by workspace analysis. Domain is a
// best-effort hint used to suppress cascading diagnostics when syntax is
// temporarily incomplete and the full domain declaration cannot be parsed.
// Root identifies one logical compiler input so separate copies of the same
// named domain in a larger editor workspace are not merged.
type Source struct {
	Path             string
	Domain           string
	Root             string
	ExpectedDomain   string
	Content          []byte
	Parsed           *grammar.SkelContent
	ParseDiagnostics Diagnostics
}

type _CachedWorkspaceParse struct {
	hash        [32]byte
	content     *grammar.SkelContent
	diagnostics Diagnostics
}

// WorkspaceAnalyzer caches syntax trees and successful domain analyses across
// workspace snapshots. A changed domain invalidates only itself and reverse
// dependents whose import fingerprints consequently change.
type WorkspaceAnalyzer struct {
	mu      sync.Mutex
	gate    chan struct{}
	parses  map[string]_CachedWorkspaceParse
	domains map[string]_CachedWorkspaceDomain
	stats   WorkspaceAnalysisStats
}

// WorkspaceAnalysisStats reports syntax and semantic cache usage from the most
// recent workspace analysis.
type WorkspaceAnalysisStats struct {
	ParsedSources   int
	ReusedSources   int
	AnalyzedDomains int
	ReusedDomains   int
}

// WorkspaceDomain is one successfully analyzed semantic domain in a workspace
// snapshot. Sources contains the exact in-memory inputs used to build Model.
type WorkspaceDomain struct {
	Name    string
	Root    string
	Model   *model.Domain
	Sources []Source
}

// NewWorkspaceAnalyzer creates an incremental workspace analyzer.
func NewWorkspaceAnalyzer() *WorkspaceAnalyzer {
	return &WorkspaceAnalyzer{gate: make(chan struct{}, 1), parses: map[string]_CachedWorkspaceParse{}, domains: map[string]_CachedWorkspaceDomain{}}
}

// AnalyzeWorkspace performs syntax and semantic analysis over an in-memory
// workspace. Independent failures in the same domain are collected up to the
// analyzer's diagnostic limit. Domains that depend on a syntactically or
// semantically invalid domain are skipped to avoid cascading errors.
func AnalyzeWorkspace(sources []Source) []Diagnostic {
	return NewWorkspaceAnalyzer().Analyze(sources)
}

// Analyze analyzes a workspace snapshot without cancellation.
func (w *WorkspaceAnalyzer) Analyze(sources []Source) []Diagnostic {
	diagnostics, _ := w.AnalyzeContext(context.Background(), sources)
	return diagnostics
}

// AnalyzeContext analyzes a workspace snapshot and honors cancellation.
func (w *WorkspaceAnalyzer) AnalyzeContext(ctx context.Context, sources []Source) ([]Diagnostic, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	select {
	case w.gate <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-w.gate }()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stats = WorkspaceAnalysisStats{}
	diagnostics, _, err := w.analyze(ctx, sources, false)
	return diagnostics, err
}

// AnalyzeDomainsContext analyzes a workspace snapshot and returns every
// semantic domain that compiled successfully alongside its diagnostics.
func (w *WorkspaceAnalyzer) AnalyzeDomainsContext(ctx context.Context, sources []Source) ([]Diagnostic, []WorkspaceDomain, error) {
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	select {
	case w.gate <- struct{}{}:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	defer func() { <-w.gate }()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stats = WorkspaceAnalysisStats{}
	return w.analyze(ctx, sources, false)
}

// Stats returns cache usage from the most recent analysis.
func (w *WorkspaceAnalyzer) Stats() WorkspaceAnalysisStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stats
}

func (w *WorkspaceAnalyzer) analyze(ctx context.Context, sources []Source, allowMissingImports bool) ([]Diagnostic, []WorkspaceDomain, error) {
	ordered := append([]Source{}, sources...)
	slices.SortFunc(ordered, func(left, right Source) int {
		return strings.Compare(left.Path, right.Path)
	})

	domains := map[string]*_WorkspaceDomain{}
	diagnostics := []Diagnostic{}
	for _, source := range ordered {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		content, syntaxDiagnostics := w.parseWorkspaceSource(ctx, source)
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		diagnostics = append(diagnostics, syntaxDiagnostics...)
		if content == nil {
			if source.Domain != "" {
				domain := workspaceDomain(
					domains,
					workspaceDomainKey(source.Domain, source.Root),
					source.Domain,
					source.Root,
				)
				domain.invalid = true
			}
			continue
		}
		if content.Domain == nil || content.Domain.Name == nil || content.Domain.Name.String() == "" {
			position := model.Position{File: source.Path, Line: 1, Column: 1}
			diagnostics = append(diagnostics, Diagnostic{
				Code: DiagnosticCodeDomainMissing, Severity: DiagnosticSeverityError, Position: position, Range: sourceRangeAt(position, source.Content),
				Message: "missing domain declaration",
			})
			continue
		}
		name := content.Domain.Name.String()
		if source.ExpectedDomain != "" && name != source.ExpectedDomain {
			position := parser.SourcePosition(content.Domain.Name.Pos)
			diagnostics = append(diagnostics, Diagnostic{
				Code: DiagnosticCodeDomainMismatch, Severity: DiagnosticSeverityError, Position: position, Range: sourceRangeAt(position, source.Content),
				Message: fmt.Sprintf("domain mismatch: found=%s, expected=%s", name, source.ExpectedDomain),
			})
			workspaceDomain(
				domains,
				workspaceDomainKey(source.ExpectedDomain, source.Root),
				source.ExpectedDomain,
				source.Root,
			).invalid = true
			continue
		}
		domain := workspaceDomain(domains, workspaceDomainKey(name, source.Root), name, source.Root)
		if len(syntaxDiagnostics) > 0 {
			domain.syntaxInvalid = true
		}
		domain.contents = append(domain.contents, content)
		domain.sources = append(domain.sources, source)
	}

	keys := make([]string, 0, len(domains))
	domainsByName := map[string][]*_WorkspaceDomain{}
	for key, domain := range domains {
		if len(domain.contents) > 0 {
			domain.merged = mergeDomainContents(domain.contents)
		}
		domainsByName[domain.name] = append(domainsByName[domain.name], domain)
		keys = append(keys, key)
	}
	for _, candidates := range domainsByName {
		slices.SortFunc(candidates, func(left, right *_WorkspaceDomain) int {
			return strings.Compare(left.key, right.key)
		})
	}
	slices.Sort(keys)
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		w.analyzeWorkspaceDomain(ctx, domains[key], domainsByName, &diagnostics, allowMissingImports)
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
	}
	for _, key := range keys {
		if domain := domains[key]; domain.analysis != nil {
			diagnostics = append(diagnostics, MigrationDiagnostics(domain.analysis.Model())...)
		}
	}
	contentByPath := make(map[string][]byte, len(ordered))
	for _, source := range ordered {
		contentByPath[filepath.Clean(source.Path)] = source.Content
	}
	for index := range diagnostics {
		completeDiagnostic(&diagnostics[index], contentByPath)
	}
	slices.SortFunc(diagnostics, compareDiagnostics)
	activePaths := make(map[string]bool, len(ordered))
	for _, source := range ordered {
		activePaths[source.Path] = true
	}
	for path := range w.parses {
		if !activePaths[path] {
			delete(w.parses, path)
		}
	}
	for key := range w.domains {
		if domains[key] == nil {
			delete(w.domains, key)
		}
	}
	models := make([]WorkspaceDomain, 0, len(keys))
	for _, key := range keys {
		domain := domains[key]
		if domain.state != workspaceDomainComplete || domain.analysis == nil {
			continue
		}
		models = append(models, WorkspaceDomain{
			Name: domain.name, Root: domain.root, Model: domain.analysis.Model(), Sources: append([]Source{}, domain.sources...),
		})
	}
	return diagnostics, models, nil
}

func (w *WorkspaceAnalyzer) parseWorkspaceSource(ctx context.Context, source Source) (*grammar.SkelContent, Diagnostics) {
	if source.Parsed != nil || len(source.ParseDiagnostics) > 0 {
		w.stats.ReusedSources++
		if len(source.ParseDiagnostics) > 0 {
			return source.Parsed, append(Diagnostics{}, source.ParseDiagnostics...)
		}
		return source.Parsed, nil
	}
	hash := sha256.Sum256(source.Content)
	if cached, ok := w.parses[source.Path]; ok && cached.hash == hash {
		w.stats.ReusedSources++
		return cached.content, append(Diagnostics{}, cached.diagnostics...)
	}
	w.stats.ParsedSources++
	content, diagnostics, err := ParseSourceRecoveringContext(ctx, source.Path, source.Content)
	if err != nil {
		return nil, nil
	}
	w.parses[source.Path] = _CachedWorkspaceParse{hash: hash, content: content, diagnostics: append(Diagnostics{}, diagnostics...)}
	return content, diagnostics
}
