package compiler

import (
	"context"
	"fmt"
	"path/filepath"

	"go.yorun.ai/skelc/internal/analyzer"
	"go.yorun.ai/skelc/internal/hasher"
	"go.yorun.ai/skelc/internal/parser"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

type _WorkspaceDomain struct {
	key           string
	name          string
	root          string
	contents      []*grammar.SkelContent
	invalid       bool
	syntaxInvalid bool
	merged        *grammar.SkelContent
	analysis      *analyzer.Analysis
	state         _WorkspaceDomainState
	sources       []Source
	fingerprint   string
}

type _CachedWorkspaceDomain struct {
	fingerprint string
	analysis    *analyzer.Analysis
	diagnostics Diagnostics
}

type _WorkspaceImportResolution struct {
	analyses      []*analyzer.Analysis
	valid         bool
	hasUnresolved bool
}

type _WorkspaceDomainState uint8

const (
	workspaceDomainPending _WorkspaceDomainState = iota
	workspaceDomainVisiting
	workspaceDomainComplete
	workspaceDomainFailed
)

func workspaceDomainKey(name, root string) string {
	if root == "" {
		return name
	}
	return filepath.Clean(root) + "\x00" + name
}

func workspaceDomain(domains map[string]*_WorkspaceDomain, key, name, root string) *_WorkspaceDomain {
	domain := domains[key]
	if domain == nil {
		domain = &_WorkspaceDomain{key: key, name: name, root: root}
		domains[key] = domain
	}
	return domain
}

func (w *WorkspaceAnalyzer) analyzeWorkspaceDomain(
	ctx context.Context,
	domain *_WorkspaceDomain,
	domainsByName map[string][]*_WorkspaceDomain,
	diagnostics *[]Diagnostic,
	allowMissingImports bool,
) bool {
	if ctx.Err() != nil {
		return false
	}
	if domain == nil || domain.invalid || domain.merged == nil {
		return false
	}
	switch domain.state {
	case workspaceDomainComplete:
		return true
	case workspaceDomainFailed:
		return false
	case workspaceDomainVisiting:
		position := parser.SourcePosition(domain.merged.Domain.Name.Pos)
		*diagnostics = append(*diagnostics, Diagnostic{
			Code: DiagnosticCodeImportCycle, Severity: DiagnosticSeverityError, Position: position, Range: SourceRange{Start: position, End: position},
			Message: fmt.Sprintf("cyclic domain import involving %s", domain.name),
		})
		domain.state = workspaceDomainFailed
		return false
	}

	domain.state = workspaceDomainVisiting
	imports, completed := w.resolveWorkspaceDomainImports(ctx, domain, domainsByName, diagnostics, allowMissingImports)
	if !completed {
		return false
	}
	if !imports.valid {
		domain.state = workspaceDomainFailed
		return false
	}
	return w.analyzeResolvedWorkspaceDomain(ctx, domain, imports, diagnostics)
}

func (w *WorkspaceAnalyzer) resolveWorkspaceDomainImports(
	ctx context.Context,
	domain *_WorkspaceDomain,
	domainsByName map[string][]*_WorkspaceDomain,
	diagnostics *[]Diagnostic,
	allowMissingImports bool,
) (_WorkspaceImportResolution, bool) {
	resolution := _WorkspaceImportResolution{
		analyses: make([]*analyzer.Analysis, 0, len(domain.merged.Imports)),
		valid:    true,
	}
	seenImports := map[string]bool{}
	for _, importDecl := range domain.merged.Imports {
		if ctx.Err() != nil {
			return resolution, false
		}
		name := importDecl.Domain.String()
		if seenImports[name] {
			continue
		}
		seenImports[name] = true
		if domain.root != "" && !w.options.ResolveIsolatedImports {
			resolution.hasUnresolved = true
			continue
		}
		candidates := domainsByName[name]
		if len(candidates) > 1 {
			resolution.hasUnresolved = true
			continue
		}
		var imported *_WorkspaceDomain
		if len(candidates) == 1 {
			imported = candidates[0]
		}
		if imported != nil && (imported.invalid || imported.syntaxInvalid) {
			resolution.valid = false
			continue
		}
		if imported == nil || imported.merged == nil {
			if allowMissingImports {
				resolution.hasUnresolved = true
				continue
			}
			*diagnostics = append(*diagnostics, Diagnostic{
				Code: DiagnosticCodeImportMissing, Severity: DiagnosticSeverityError, Position: parser.SourcePosition(importDecl.Pos),
				Message: fmt.Sprintf("skel import %s not found in the workspace", name),
			})
			resolution.valid = false
			continue
		}
		if !w.analyzeWorkspaceDomain(ctx, imported, domainsByName, diagnostics, allowMissingImports) {
			if ctx.Err() != nil {
				return resolution, false
			}
			resolution.valid = false
			continue
		}
		resolution.analyses = append(resolution.analyses, imported.analysis)
	}
	return resolution, true
}

func (w *WorkspaceAnalyzer) analyzeResolvedWorkspaceDomain(
	ctx context.Context,
	domain *_WorkspaceDomain,
	imports _WorkspaceImportResolution,
	diagnostics *[]Diagnostic,
) bool {
	domain.fingerprint = w.graph.inputs[domain.key]
	if cached, ok := w.domains[domain.key]; ok && cached.fingerprint == domain.fingerprint {
		w.stats.ReusedDomains++
		domain.analysis = cached.analysis
		domain.state = workspaceDomainComplete
		if cached.analysis == nil {
			domain.state = workspaceDomainFailed
		}
		*diagnostics = append(*diagnostics, cloneDiagnostics(cached.diagnostics)...)
		return domain.state == workspaceDomainComplete
	}
	var analysis *analyzer.Analysis
	var analysisErrors []error
	w.stats.AnalyzedDomains++
	if imports.hasUnresolved {
		analysis, analysisErrors, _ = analyzer.AnalyzeImportContext(ctx, domain.merged)
	} else {
		analysis, analysisErrors, _ = analyzer.AnalyzeContext(ctx, domain.merged, imports.analyses)
	}
	if ctx.Err() != nil {
		return false
	}
	local := Diagnostics{}
	for _, analysisError := range analysisErrors {
		local = append(local, diagnosticFromError(domain.merged.Pos.Filename, DiagnosticCodeSemanticValidation, analysisError))
	}
	if len(local) == 0 {
		if err := hasher.FillHashes(analysis.Model()); err != nil {
			local = append(local, diagnosticFromError(domain.merged.Pos.Filename, DiagnosticCodeSemanticValidation, err))
		}
	}
	if ctx.Err() != nil {
		return false
	}
	if len(local) > 0 {
		*diagnostics = append(*diagnostics, local...)
		domain.state = workspaceDomainFailed
		w.domains[domain.key] = _CachedWorkspaceDomain{fingerprint: domain.fingerprint, diagnostics: cloneDiagnostics(local)}
		return false
	}

	domain.analysis = analysis
	domain.state = workspaceDomainComplete
	w.domains[domain.key] = _CachedWorkspaceDomain{fingerprint: domain.fingerprint, analysis: analysis}
	return true
}
