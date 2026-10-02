package lsp

import (
	"context"
	"encoding/json"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/lsp/analysis"
)

type _SchemaCompatibilitySettings struct {
	Diagnostics       bool   `json:"diagnostics"`
	IncludeCompatible bool   `json:"includeCompatible"`
	CodeLens          bool   `json:"codeLens"`
	Baseline          string `json:"baseline"`
}

type _InitializationOptions struct {
	SchemaCompatibility _SchemaCompatibilitySettings `json:"schemaCompatibility"`
}

type _SchemaCompatibilitySettingsPatch struct {
	Diagnostics       *bool   `json:"diagnostics"`
	IncludeCompatible *bool   `json:"includeCompatible"`
	CodeLens          *bool   `json:"codeLens"`
	Baseline          *string `json:"baseline"`
}

func defaultSchemaCompatibilitySettings() _SchemaCompatibilitySettings {
	return _SchemaCompatibilitySettings{CodeLens: true}
}

func decodeInitializationSettings(value protocol.LSPAny) _SchemaCompatibilitySettings {
	settings := defaultSchemaCompatibilitySettings()
	if len(value) == 0 {
		return settings
	}
	options := _InitializationOptions{SchemaCompatibility: settings}
	if json.Unmarshal(value, &options) == nil {
		return options.SchemaCompatibility
	}
	return settings
}

func decodeChangedSettings(value protocol.LSPAny, fallback _SchemaCompatibilitySettings) _SchemaCompatibilitySettings {
	if len(value) == 0 {
		return fallback
	}
	var envelope struct {
		Skelc struct {
			SchemaCompatibility *_SchemaCompatibilitySettingsPatch `json:"schemaCompatibility"`
		} `json:"skelc"`
		SchemaCompatibility *_SchemaCompatibilitySettingsPatch `json:"schemaCompatibility"`
	}
	if json.Unmarshal(value, &envelope) != nil {
		return fallback
	}
	if envelope.Skelc.SchemaCompatibility != nil {
		return applySchemaCompatibilitySettings(fallback, *envelope.Skelc.SchemaCompatibility)
	}
	if envelope.SchemaCompatibility != nil {
		return applySchemaCompatibilitySettings(fallback, *envelope.SchemaCompatibility)
	}
	return fallback
}

func applySchemaCompatibilitySettings(settings _SchemaCompatibilitySettings, patch _SchemaCompatibilitySettingsPatch) _SchemaCompatibilitySettings {
	if patch.Diagnostics != nil {
		settings.Diagnostics = *patch.Diagnostics
	}
	if patch.IncludeCompatible != nil {
		settings.IncludeCompatible = *patch.IncludeCompatible
	}
	if patch.CodeLens != nil {
		settings.CodeLens = *patch.CodeLens
	}
	if patch.Baseline != nil {
		settings.Baseline = *patch.Baseline
	}
	return settings
}

func decodeStrictSettings(value protocol.LSPAny, fallback bool) bool {
	var settings struct {
		Strict *bool `json:"strict"`
		Skelc  *struct {
			Strict *bool `json:"strict"`
		} `json:"skelc"`
	}
	if json.Unmarshal(value, &settings) != nil {
		return fallback
	}
	if settings.Skelc != nil && settings.Skelc.Strict != nil {
		return *settings.Skelc.Strict
	}
	if settings.Strict != nil {
		return *settings.Strict
	}
	return fallback
}

func (s *_Server) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	s.stateMu.Lock()

	s.mu.Lock()
	previous := s.schemaCompatibility
	previousStrict := s.strict
	s.schemaCompatibility = decodeChangedSettings(params.Settings, previous)
	s.strict = decodeStrictSettings(params.Settings, previousStrict)
	compatibilityChanged := previous != s.schemaCompatibility
	changed := compatibilityChanged || previousStrict != s.strict
	client := s.client
	refreshCodeLens := s.codeLensRefreshSupport
	s.mu.Unlock()
	if !changed {
		s.stateMu.Unlock()
		return nil
	}
	s.invalidateSemanticDiagnostics(ctx)
	s.stateMu.Unlock()
	if client != nil && refreshCodeLens && compatibilityChanged {
		_ = client.CodeLensRefresh(ctx)
	}
	return nil
}

func (s *_Server) analysisOptions() analysis.Options {
	s.mu.RLock()
	settings := s.schemaCompatibility
	strict := s.strict
	s.mu.RUnlock()
	return analysis.Options{
		Strict: strict,
		Compatibility: analysis.CompatibilityOptions{
			Enabled: settings.Diagnostics, IncludeCompatible: settings.IncludeCompatible, BaselineSkelIn: settings.Baseline,
		},
	}
}
