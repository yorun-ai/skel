package features

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/lsp/source"
)

func (s *Service) PrepareRename(_ context.Context, params *protocol.PrepareRenameParams) (protocol.PrepareRenameResult, error) {
	snapshot := s.Snapshot
	document := snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return nil, nil
	}
	occurrence, ok := occurrenceAt(document, params.Position)
	if !ok || len(snapshot.Definitions(snapshot.ResolveKey(document, occurrence.Key))) != 1 {
		return nil, nil
	}
	range_ := occurrence.Range
	return &range_, nil
}

func (s *Service) Rename(_ context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	if !isIdentifierValue(params.NewName) {
		return nil, fmt.Errorf("invalid Skel identifier %q", params.NewName)
	}
	snapshot := s.Snapshot
	document := snapshot.Document(params.TextDocument.URI)
	if document == nil {
		return nil, nil
	}
	occurrence, ok := occurrenceAt(document, params.Position)
	if !ok || len(snapshot.Definitions(snapshot.ResolveKey(document, occurrence.Key))) != 1 {
		return nil, nil
	}
	oldName := strings.TrimPrefix(occurrence.Key, domainFromKey(occurrence.Key)+".")
	for _, candidate := range snapshot.DocumentsFor(document, domainFromKey(occurrence.Key)) {
		if candidate.Parsed != nil {
			for _, entry := range candidate.Parsed.Entries {
				if entry.Data == nil {
					continue
				}
				for _, parameter := range entry.Data.TypeParameters {
					if parameter.Name.Value == params.NewName && params.NewName != oldName {
						return nil, fmt.Errorf("Skel name %s conflicts with a generic parameter", params.NewName)
					}
				}
			}
		}
		for _, definition := range candidate.Definitions {
			if definition.Name == params.NewName && definition.Key != occurrence.Key {
				return nil, fmt.Errorf("Skel declaration %s already exists", definition.Key)
			}
		}
	}
	changes := map[uri.URI][]protocol.TextEdit{}
	for _, location := range snapshot.Occurrences(snapshot.ResolveKey(document, occurrence.Key)) {
		changes[location.Document.URI] = append(changes[location.Document.URI], protocol.TextEdit{Range: location.Occurrence.Range, NewText: params.NewName})
	}
	for documentURI := range changes {
		slices.SortFunc(changes[documentURI], func(left, right protocol.TextEdit) int {
			return source.ComparePosition(left.Range.Start, right.Range.Start)
		})
	}
	if s.DocumentChangesSupport {
		result := &protocol.WorkspaceEdit{}
		uris := make([]uri.URI, 0, len(changes))
		for documentURI := range changes {
			uris = append(uris, documentURI)
		}
		slices.Sort(uris)
		for _, documentURI := range uris {
			target := snapshot.Document(documentURI)
			var version *int32
			if target.Open {
				version = new(target.Version)
			}
			edits := make([]protocol.TextDocumentEditElement, 0, len(changes[documentURI]))
			for _, edit := range changes[documentURI] {
				edits = append(edits, new(edit))
			}
			result.DocumentChanges = append(result.DocumentChanges, &protocol.TextDocumentEdit{TextDocument: protocol.OptionalVersionedTextDocumentIdentifier{TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI}, Version: version}, Edits: edits})
		}
		return result, nil
	}
	return &protocol.WorkspaceEdit{Changes: changes}, nil
}

func domainFromKey(key string) string {
	if index := strings.LastIndex(key, "."); index >= 0 {
		return key[:index]
	}
	return ""
}
