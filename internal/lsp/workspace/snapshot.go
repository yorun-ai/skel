package workspace

import (
	"path/filepath"
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skel/internal/binding"
	"go.yorun.ai/skel/internal/lsp/source"
)

// Snapshot is an immutable workspace view used by one request or analysis.
type Snapshot struct {
	revision            uint64
	documents           map[uri.URI]*Document
	ordered             []*Document
	byDomain            map[string][]*Document
	definitions         map[string][]DefinitionLocation
	occurrences         map[string][]OccurrenceLocation
	symbols             map[string][]binding.Symbol
	documentOccurrences map[uri.URI][]Occurrence
	roots               map[uri.URI]string
	domainRoots         map[string]map[string]bool
}

// DefinitionLocation identifies a definition and its containing document.
type DefinitionLocation struct {
	Document   *Document
	Definition Definition
}

// OccurrenceLocation identifies an occurrence and its containing document.
type OccurrenceLocation struct {
	Document   *Document
	Occurrence Occurrence
}

func newSnapshot(revision uint64, documents map[uri.URI]*Document, ordered []*Document) Snapshot {
	snapshot := Snapshot{
		revision: revision, documents: documents, ordered: ordered,
		byDomain: map[string][]*Document{}, definitions: map[string][]DefinitionLocation{},
		occurrences: map[string][]OccurrenceLocation{}, roots: map[uri.URI]string{}, domainRoots: map[string]map[string]bool{}, symbols: map[string][]binding.Symbol{}, documentOccurrences: map[uri.URI][]Occurrence{},
	}
	directories := map[uri.URI]bool{}
	for _, document := range ordered {
		if filepath.Base(document.Path) == "domain.skel" {
			if directory, ok := sourceDirectory(document.URI); ok {
				directories[directory] = true
			}
		}
	}
	for _, document := range ordered {
		root := string(document.URI)
		if directory, ok := sourceDirectory(document.URI); ok && directories[directory] {
			root = string(directory)
		}
		snapshot.roots[document.URI] = root
		if snapshot.domainRoots[document.Domain] == nil {
			snapshot.domainRoots[document.Domain] = map[string]bool{}
		}
		snapshot.domainRoots[document.Domain][root] = true
		snapshot.byDomain[document.Domain] = append(snapshot.byDomain[document.Domain], document)
		for _, symbol := range document.Bindings.Symbols {
			key := root + "\x00" + symbol.ID.Key()
			snapshot.symbols[key] = append(snapshot.symbols[key], symbol)
		}
		for _, definition := range document.Definitions {
			if !definition.Confirmed {
				continue
			}
			key := root + "\x00" + definition.Key
			snapshot.definitions[key] = append(snapshot.definitions[key], DefinitionLocation{Document: document, Definition: definition})
		}
	}
	for _, document := range ordered {
		occurrences := bindOccurrences(document, func(id binding.SymbolID) []binding.Symbol {
			return snapshot.symbols[snapshot.ResolveKey(document, id.Key())]
		})
		snapshot.documentOccurrences[document.URI] = occurrences
		for _, occurrence := range occurrences {
			if occurrence.Binding.Status != binding.Resolved && occurrence.Binding.Status != binding.Ambiguous {
				continue
			}
			key := snapshot.ResolveKey(document, occurrence.Key)
			if key != "" {
				snapshot.occurrences[key] = append(snapshot.occurrences[key], OccurrenceLocation{Document: document, Occurrence: occurrence})
			}
		}
	}
	return snapshot
}

// Revision identifies the store state represented by the snapshot.
func (s Snapshot) Revision() uint64 { return s.revision }

// Document returns a document by URI.
func (s Snapshot) Document(documentURI uri.URI) *Document { return s.documents[documentURI] }

// Documents returns all documents in stable URI order.
func (s Snapshot) Documents() []*Document { return append([]*Document{}, s.ordered...) }

// DocumentsMap returns a copy of the URI lookup map.
func (s Snapshot) DocumentsMap() map[uri.URI]*Document {
	result := make(map[uri.URI]*Document, len(s.documents))
	for documentURI, document := range s.documents {
		result[documentURI] = document
	}
	return result
}

// Definitions returns declarations for a key produced by ResolveKey.
func (s Snapshot) Definitions(key string) []DefinitionLocation {
	return append([]DefinitionLocation{}, s.definitions[key]...)
}

// Occurrences returns references for a key produced by ResolveKey.
func (s Snapshot) Occurrences(key string) []OccurrenceLocation {
	return append([]OccurrenceLocation{}, s.occurrences[key]...)
}

// ResolveKey binds local references to the compiler input and imported references
// only to an unambiguous input. Copies of a domain must never share rename edits.
func (s Snapshot) ResolveKey(document *Document, key string) string {
	id := binding.ParseKey(key)
	domain := id.Domain
	if !strings.Contains(key, ".") {
		return ""
	}
	root := s.roots[document.URI]
	if domain != document.Domain {
		roots := s.domainRoots[domain]
		if len(roots) != 1 {
			return ""
		}
		for candidate := range roots {
			root = candidate
		}
	}
	return root + "\x00" + key
}

// DocumentsFor resolves a domain using the same input boundaries as navigation.
func (s Snapshot) DocumentsFor(document *Document, domain string) []*Document {
	key := s.ResolveKey(document, domain+".")
	if key == "" {
		return nil
	}
	root := key[:strings.IndexByte(key, 0)]
	result := []*Document{}
	for _, candidate := range s.byDomain[domain] {
		if s.roots[candidate.URI] == root {
			result = append(result, candidate)
		}
	}
	return result
}

// OccurrenceAt reads bindings for this workspace revision without mutating the
// shared per-document syntax index.
func (s Snapshot) OccurrenceAt(document *Document, position protocol.Position) (Occurrence, bool) {
	for _, occurrence := range s.documentOccurrences[document.URI] {
		if occurrence.Binding.Status != binding.Resolved && occurrence.Binding.Status != binding.Ambiguous {
			continue
		}
		if source.ContainsPosition(occurrence.Range, position) {
			return occurrence, true
		}
	}
	return Occurrence{}, false
}
