package lsp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skel/internal/lsp/analysis"
)

func TestStrictSettingsInitializationAndChanges(t *testing.T) {
	for _, test := range []struct {
		value    string
		fallback bool
		want     bool
	}{
		{value: `{}`, fallback: true, want: true},
		{value: `{"strict":true}`, want: true},
		{value: `{"strict":false}`, fallback: true},
		{value: `{"skelc":{"strict":true}}`, want: true},
		{value: `{"skelc":{"strict":false}}`, fallback: true},
		{value: `{"schemaCompatibility":{"codeLens":false}}`, fallback: true, want: true},
		{value: `{"strict":"true"}`, fallback: true, want: true},
	} {
		assert.Equal(t, test.want, decodeStrictSettings(protocol.LSPAny(test.value), test.fallback), test.value)
	}
	server := newServer()
	t.Cleanup(server.stopSemanticAnalysis)
	_, err := server.Initialize(t.Context(), &protocol.InitializeParams{InitializationOptions: protocol.LSPAny(`{"strict":true}`)})
	require.NoError(t, err)
	assert.True(t, server.strict)
	require.NoError(t, server.DidChangeConfiguration(t.Context(), &protocol.DidChangeConfigurationParams{Settings: protocol.LSPAny(`{"strict":false}`)}))
	assert.False(t, server.strict)
}

func TestStrictConfigurationRepublishesDiagnostics(t *testing.T) {
	server := newServer()
	// The republished diagnostics come from the analysis runner. Drop the debounce
	// delay so the wait below covers analysis work instead of a fixed timer that a
	// loaded CI runner can delay.
	server.analysis = analysis.NewRunner(0)
	t.Cleanup(server.stopSemanticAnalysis)
	documentURI := uri.File("/workspace/order.skel")
	client := &recordingClient{diagnostics: make(chan *protocol.PublishDiagnosticsParams, 16)}
	server.client = client
	require.NoError(t, server.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{TextDocument: protocol.TextDocumentItem{
		URI: documentURI, LanguageID: "skel", Version: 1, Text: "domain demo.order\nservice OrderService { method ping {} }\n",
	}}))
	for _, strict := range []bool{false, true, false} {
		settings, err := json.Marshal(map[string]bool{"strict": strict})
		require.NoError(t, err)
		require.NoError(t, server.DidChangeConfiguration(t.Context(), &protocol.DidChangeConfigurationParams{Settings: settings}))
		severity := protocol.DiagnosticSeverityError
		waitForDiagnostics(t, client.diagnostics, func(params *protocol.PublishDiagnosticsParams) bool {
			return params.URI == documentURI && len(params.Diagnostics) == 1 && params.Diagnostics[0].Severity == severity
		})
	}
}

func TestDecodeChangedSettingsMergesPartialUpdates(t *testing.T) {
	fallback := _SchemaCompatibilitySettings{
		Diagnostics: true, IncludeCompatible: true, CodeLens: true, Baseline: "../baseline",
	}
	for _, test := range []struct {
		name     string
		settings any
	}{
		{name: "direct", settings: map[string]any{"schemaCompatibility": map[string]any{"codeLens": false}}},
		{name: "skelc envelope", settings: map[string]any{"skelc": map[string]any{
			"schemaCompatibility": map[string]any{"codeLens": false},
		}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			content, err := json.Marshal(test.settings)
			require.NoError(t, err)
			assert.Equal(t, _SchemaCompatibilitySettings{
				Diagnostics: true, IncludeCompatible: true, CodeLens: false, Baseline: "../baseline",
			}, decodeChangedSettings(content, fallback))
		})
	}
}

func TestDidChangeConfigurationClearsStaleCompatibilityDiagnostics(t *testing.T) {
	server := newServer()
	t.Cleanup(server.stopSemanticAnalysis)
	documentURI := uri.File("/workspace/domain.skel")
	client := &recordingClient{diagnostics: make(chan *protocol.PublishDiagnosticsParams, 1)}
	server.client = client
	server.schemaCompatibility = _SchemaCompatibilitySettings{Diagnostics: true, CodeLens: true}
	server.semantic[documentURI] = []protocol.Diagnostic{{Code: protocol.String("schema.declaration.removed")}}
	server.publisher.visible[documentURI] = true
	settings, err := json.Marshal(map[string]any{
		"schemaCompatibility": map[string]any{"diagnostics": false},
	})
	require.NoError(t, err)

	require.NoError(t, server.DidChangeConfiguration(t.Context(), &protocol.DidChangeConfigurationParams{Settings: settings}))
	assert.False(t, server.schemaCompatibility.Diagnostics)
	assert.True(t, server.schemaCompatibility.CodeLens)
	assert.Empty(t, server.semantic)
	select {
	case published := <-client.diagnostics:
		assert.Equal(t, documentURI, published.URI)
		assert.Empty(t, published.Diagnostics)
	case <-time.After(5 * time.Second):
		t.Fatal("expected stale compatibility diagnostics to be cleared")
	}
}
