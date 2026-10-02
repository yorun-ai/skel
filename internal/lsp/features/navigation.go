package features

import (
	"context"

	"go.lsp.dev/protocol"
)

func (s *Service) Definition(_ context.Context, params *protocol.DefinitionParams) (protocol.DefinitionResult, error) {
	snapshot := s.Snapshot
	document := snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return protocol.LocationSlice{}, nil
	}
	occurrence, ok := occurrenceAt(document, params.Position)
	if !ok {
		return protocol.LocationSlice{}, nil
	}
	locations := make([]protocol.Location, 0)
	for _, location := range snapshot.Definitions(snapshot.ResolveKey(document, occurrence.Key)) {
		locations = append(locations, protocol.Location{URI: location.Document.URI, Range: location.Definition.Range})
	}
	sortLocations(locations)
	return protocol.LocationSlice(locations), nil
}

func (s *Service) References(_ context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	snapshot := s.Snapshot
	document := snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return []protocol.Location{}, nil
	}
	occurrence, ok := occurrenceAt(document, params.Position)
	if !ok {
		return []protocol.Location{}, nil
	}
	locations := make([]protocol.Location, 0)
	definitions := make(map[protocol.Location]bool)
	for _, location := range snapshot.Definitions(snapshot.ResolveKey(document, occurrence.Key)) {
		definitions[protocol.Location{URI: location.Document.URI, Range: location.Definition.Range}] = true
	}
	for _, occurrenceLocation := range snapshot.Occurrences(snapshot.ResolveKey(document, occurrence.Key)) {
		location := protocol.Location{URI: occurrenceLocation.Document.URI, Range: occurrenceLocation.Occurrence.Range}
		if params.Context.IncludeDeclaration || !definitions[location] {
			locations = append(locations, location)
		}
	}
	sortLocations(locations)
	return locations, nil
}

func (s *Service) DocumentSymbol(_ context.Context, params *protocol.DocumentSymbolParams) (protocol.DocumentSymbolResult, error) {
	document := s.Snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return protocol.DocumentSymbolSlice{}, nil
	}
	return documentSymbols(document.Symbols), nil
}
