package lsp

import (
	"context"
	"slices"
	"time"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/lsp/analysis"
)

const semanticAnalysisDelay = 200 * time.Millisecond

func (s *_Server) stopSemanticAnalysis() {
	s.analysis.Stop()
	s.publisher.stop()
}

func (s *_Server) rememberClient(ctx context.Context) {
	client, ok := protocol.ClientFromContext(ctx)
	if !ok {
		return
	}
	s.mu.Lock()
	s.client = client
	s.mu.Unlock()
}

// invalidateSemanticDiagnostics removes results produced from an older
// workspace snapshot before scheduling a replacement. Syntax diagnostics are
// retained because they are computed directly from the current document.
func (s *_Server) invalidateSemanticDiagnostics(ctx context.Context) {
	s.rememberClient(ctx)
	s.diagnosticGeneration++
	pending := s.publisher.advance(s.diagnosticGeneration)
	s.mu.Lock()
	stale := append([]uri.URI{}, pending...)
	for documentURI := range s.semantic {
		stale = append(stale, documentURI)
	}
	s.semantic = map[uri.URI][]protocol.Diagnostic{}
	client := s.client
	s.mu.Unlock()
	s.scheduleSemanticAnalysis()

	if client == nil {
		return
	}
	slices.Sort(stale)
	stale = slices.Compact(stale)
	snapshot := s.workspace.Snapshot()
	for _, documentURI := range stale {
		_ = s.publishDocumentDiagnostics(ctx, client, documentURI, snapshot.Document(documentURI))
	}
}

func (s *_Server) scheduleSemanticAnalysis() {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()
	if client == nil {
		return
	}
	s.analysis.Schedule(s.workspace.Snapshot(), s.analysisOptions(), s.acceptSemanticAnalysis)
}

func (s *_Server) acceptSemanticAnalysis(result analysis.Result) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if !s.analysis.IsCurrent(result) {
		return
	}
	if result.Revision != s.workspace.Revision() {
		return
	}
	s.mu.Lock()
	client := s.client
	if client == nil {
		s.mu.Unlock()
		return
	}
	previous := s.semantic
	s.semantic = result.Diagnostics
	changed := make(map[uri.URI]bool, len(previous)+len(result.Diagnostics))
	for documentURI := range previous {
		changed[documentURI] = true
	}
	for documentURI := range result.Diagnostics {
		changed[documentURI] = true
	}
	s.mu.Unlock()

	documentURIs := make([]uri.URI, 0, len(changed))
	for documentURI := range changed {
		documentURIs = append(documentURIs, documentURI)
	}
	slices.Sort(documentURIs)
	snapshot := s.workspace.Snapshot()
	ctx := context.Background()
	for _, documentURI := range documentURIs {
		_ = s.publishDocumentDiagnostics(ctx, client, documentURI, snapshot.Document(documentURI))
	}
}
