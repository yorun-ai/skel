package compiler

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/internal/parser/grammar"
	textsource "go.yorun.ai/skel/internal/source"
	"go.yorun.ai/skel/internal/symbol"
	"go.yorun.ai/skel/schema"
)

// Source is an in-memory Skel document used by workspace analysis. Domain is a
// best-effort hint used to suppress cascading diagnostics when syntax is
// temporarily incomplete and the full domain declaration cannot be parsed.
// Root identifies one logical compiler input so separate copies of the same
// named domain in a larger editor workspace are not merged.
type Source struct {
	// Document, when supplied, owns the authoritative immutable source revision.
	Document       *textsource.Document
	Bindings       *symbol.Document
	Path           string
	Domain         string
	Root           string
	ExpectedDomain string
	// DirectoryInput enables the same domain.skel constraints as directory compilation.
	DirectoryInput   bool
	Content          []byte
	Parsed           *grammar.SkelContent
	ParseDiagnostics Diagnostics
}

type _CachedWorkspaceParse struct {
	hash        [32]byte
	content     *grammar.SkelContent
	diagnostics Diagnostics
}

// WorkspaceAnalyzer caches syntax trees and complete domain outcomes across
// workspace snapshots. A changed domain invalidates only itself and reverse
// dependents recorded in the explicit reverse dependency graph.
type WorkspaceAnalyzer struct {
	mu      sync.Mutex
	gate    chan struct{}
	parses  map[string]_CachedWorkspaceParse
	domains map[string]_CachedWorkspaceDomain
	stats   WorkspaceAnalysisStats
	options AnalysisOptions
	graph   _WorkspaceGraph
	result  _CachedWorkspaceResult
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
// snapshot. Schema and its reachable declarations are immutable after publication.
// Sources contains the exact input revisions used to build Schema.
type WorkspaceDomain struct {
	Name    string
	Root    string
	Schema  *schema.Domain
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
	diagnostics, _, err := w.AnalyzeDomainsContext(ctx, sources)
	return diagnostics, err
}

func (w *WorkspaceAnalyzer) AnalyzeDomainsContext(ctx context.Context, sources []Source) ([]Diagnostic, []WorkspaceDomain, error) {
	return w.AnalyzeWithOptionsContext(ctx, sources, AnalysisOptions{})
}

// AnalyzeWithOptionsContext is the shared semantic pipeline for compile, check,
// and workspace analysis. The gate is cancellable, including while queued.
func (w *WorkspaceAnalyzer) AnalyzeWithOptionsContext(ctx context.Context, sources []Source, options AnalysisOptions) ([]Diagnostic, []WorkspaceDomain, error) {
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
	if w.options != options {
		w.domains = map[string]_CachedWorkspaceDomain{}
		w.result = _CachedWorkspaceResult{}
		w.graph = _WorkspaceGraph{}
	}
	w.options = options
	return w.analyze(ctx, sources, options.AllowUnresolvedImports)
}

// Stats returns cache usage from the most recent analysis.
func (w *WorkspaceAnalyzer) Stats() WorkspaceAnalysisStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stats
}

func (w *WorkspaceAnalyzer) analyze(ctx context.Context, sources []Source, allowMissingImports bool) ([]Diagnostic, []WorkspaceDomain, error) {
	ordered := append([]Source{}, sources...)
	for i := range ordered {
		if ordered[i].Document == nil {
			ordered[i].Document = textsource.New(textsource.ID(ordered[i].Path), ordered[i].Path, 0, string(ordered[i].Content))
		} else {
			if ordered[i].Path == "" {
				ordered[i].Path = ordered[i].Document.AnalysisPath()
			}
		}
	}
	slices.SortFunc(ordered, func(left, right Source) int {
		return strings.Compare(left.Path, right.Path)
	})

	fingerprint := workspaceFingerprint(ordered)
	if w.result.fingerprint == fingerprint {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		w.stats.ReusedSources = len(ordered)
		w.stats.ReusedDomains = w.result.domainCount
		return cloneDiagnostics(w.result.diagnostics), cloneWorkspaceDomains(w.result.domains), nil
	}

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
			name := source.Domain
			if source.ExpectedDomain != "" {
				name = source.ExpectedDomain
			}
			if name != "" {
				domain := workspaceDomain(
					domains,
					workspaceDomainKey(name, source.Root),
					name,
					source.Root,
				)
				domain.invalid = true
			}
			continue
		}
		if content.Domain == nil || content.Domain.Name == nil || content.Domain.Name.String() == "" {
			position := schema.Position{File: source.Path, Line: 1, Column: 1}
			diagnostics = append(diagnostics, Diagnostic{
				Code: DiagnosticCodeDomainMissing, Severity: DiagnosticSeverityError, Position: position, Range: sourceRangeAtDocument(position, source.Document),
				Message: "missing domain declaration",
			})
			if source.ExpectedDomain != "" {
				workspaceDomain(domains, workspaceDomainKey(source.ExpectedDomain, source.Root), source.ExpectedDomain, source.Root).invalid = true
			}
			continue
		}
		if source.DirectoryInput && len(syntaxDiagnostics) == 0 {
			if issue := inspectDirectorySource(source.Path, source.ExpectedDomain, content); issue != nil {
				diagnostics = append(diagnostics, Diagnostic{Code: issue.code, Severity: DiagnosticSeverityError, Position: issue.position, Message: issue.message})
				name := source.ExpectedDomain
				if name == "" {
					name = content.Domain.Name.String()
				}
				workspaceDomain(domains, workspaceDomainKey(name, source.Root), name, source.Root).invalid = true
				continue
			}
		}
		name := content.Domain.Name.String()
		if source.ExpectedDomain != "" && name != source.ExpectedDomain {
			position := parser.SourcePosition(content.Domain.Name.Pos)
			diagnostics = append(diagnostics, Diagnostic{
				Code: DiagnosticCodeDomainMismatch, Severity: DiagnosticSeverityError, Position: position, Range: sourceRangeAtDocument(position, source.Document),
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
		if source.Bindings == nil {
			source.Bindings = symbol.Build(source.Document, content)
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
	w.invalidateDomainCache(domains, domainsByName)
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
			if w.options.IncludeWarnings {
				diagnostics = appendAnalysisWarnings(diagnostics, domain.analysis.Warnings())
			}
		}
	}
	contentByPath := make(map[string]*textsource.Document, len(ordered))
	for _, source := range ordered {
		contentByPath[filepath.Clean(source.Path)] = source.Document
	}
	for index := range diagnostics {
		completeDiagnostic(&diagnostics[index], contentByPath)
	}
	slices.SortFunc(diagnostics, compareDiagnostics)
	activePaths := make(map[string]bool, len(ordered))
	for _, source := range ordered {
		activePaths[workspaceParseKey(source)] = true
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
	schemas := make([]WorkspaceDomain, 0, len(keys))
	for _, key := range keys {
		domain := domains[key]
		if domain.state != workspaceDomainComplete || domain.analysis == nil || domain.syntaxInvalid {
			continue
		}
		schemas = append(schemas, WorkspaceDomain{
			Name: domain.name, Root: domain.root, Schema: domain.analysis.Schema(), Sources: append([]Source{}, domain.sources...),
		})
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	w.result = _CachedWorkspaceResult{fingerprint: fingerprint, diagnostics: cloneDiagnostics(diagnostics), domains: cloneWorkspaceDomains(schemas), domainCount: len(domains)}
	return diagnostics, schemas, nil
}

func (w *WorkspaceAnalyzer) parseWorkspaceSource(ctx context.Context, source Source) (*grammar.SkelContent, Diagnostics) {
	if source.Parsed != nil || len(source.ParseDiagnostics) > 0 {
		w.stats.ReusedSources++
		if len(source.ParseDiagnostics) > 0 {
			return source.Parsed, cloneDiagnostics(source.ParseDiagnostics)
		}
		return source.Parsed, nil
	}
	hash := source.Document.Digest()
	if cached, ok := w.parses[workspaceParseKey(source)]; ok && cached.hash == hash {
		w.stats.ReusedSources++
		return cached.content, cloneDiagnostics(cached.diagnostics)
	}
	w.stats.ParsedSources++
	content, diagnostics, err := ParseSourceRecoveringContext(ctx, source.Path, []byte(source.Document.Text()))
	if err != nil {
		return nil, nil
	}
	w.parses[workspaceParseKey(source)] = _CachedWorkspaceParse{hash: hash, content: content, diagnostics: cloneDiagnostics(diagnostics)}
	return content, diagnostics
}

func workspaceParseKey(input Source) string { return string(input.Document.ID()) + "\x00" + input.Path }
