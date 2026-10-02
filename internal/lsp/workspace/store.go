// Package workspace builds document indexes, owns the mutable set of documents
// visible to the language server, and exposes immutable snapshots for analysis
// and language features.
package workspace

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// Store owns open documents, workspace roots, and on-disk document indexes.
type Store struct {
	mu             sync.RWMutex
	diskMu         sync.Mutex
	cached         *Snapshot
	documents      map[uri.URI]*Document
	open           map[uri.URI]bool
	workspaceFiles map[uri.URI]map[uri.URI]struct{}
	revision       uint64
}

// New creates an empty workspace store.
func New() *Store {
	return &Store{
		documents:      map[uri.URI]*Document{},
		open:           map[uri.URI]bool{},
		workspaceFiles: map[uri.URI]map[uri.URI]struct{}{},
	}
}

// Put indexes an in-memory document. Open documents take precedence over
// workspace files loaded from disk.
func (s *Store) Put(documentURI uri.URI, content string, version int32, open bool) {
	document := BuildDocument(documentURI, documentURI.FsPath(), content, version)
	s.mu.Lock()
	defer s.mu.Unlock()
	if previous := s.documents[documentURI]; open && s.open[documentURI] && previous != nil && version <= previous.Version {
		return
	}
	document.Open = open
	s.documents[documentURI] = document
	if open {
		s.open[documentURI] = true
	}
	s.revision++
	s.cached = nil
}

// Close closes an editor document and restores its on-disk contents when it
// remains tracked by a workspace root. It reports whether the document still
// exists in the store.
func (s *Store) Close(documentURI uri.URI) bool {
	s.diskMu.Lock()
	defer s.diskMu.Unlock()
	s.mu.Lock()
	delete(s.open, documentURI)
	previous := s.documents[documentURI]
	tracked := s.documentTrackedLocked(documentURI)
	s.mu.Unlock()
	var document *Document
	if tracked {
		if content, err := os.ReadFile(documentURI.FsPath()); err == nil {
			document = BuildDocument(documentURI, documentURI.FsPath(), string(content), 0)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.open[documentURI] || s.documents[documentURI] != previous {
		return s.documents[documentURI] != nil
	}
	if document != nil && s.documentTrackedLocked(documentURI) {
		s.documents[documentURI] = document
	} else {
		delete(s.documents, documentURI)
		s.untrackLocked(documentURI)
	}
	s.revision++
	s.cached = nil
	return s.documents[documentURI] != nil
}

// ApplyFileChanges reconciles each affected source directory once. Watcher
// events can be coalesced or delayed while a generator replaces several files;
// their type is a hint to reload, not the authoritative on-disk state.
func (s *Store) ApplyFileChanges(changes []protocol.FileEvent) []uri.URI {
	s.diskMu.Lock()
	defer s.diskMu.Unlock()
	changed := []uri.URI{}
	directories := map[uri.URI]bool{}
	for _, change := range changes {
		directory, ok := sourceDirectory(change.URI)
		if !ok || directories[directory] {
			continue
		}
		directories[directory] = true
		changed = append(changed, s.refreshDirectory(directory)...)
	}
	slices.Sort(changed)
	return changed
}

// RefreshDirectory discovers siblings of an opened document, including files
// whose creation notifications were missed. Open buffers remain authoritative.
func (s *Store) RefreshDirectory(documentURI uri.URI) []uri.URI {
	directory, ok := sourceDirectory(documentURI)
	if !ok {
		return nil
	}
	s.diskMu.Lock()
	defer s.diskMu.Unlock()
	return s.refreshDirectory(directory)
}

func sourceDirectory(documentURI uri.URI) (uri.URI, bool) {
	if !documentURI.IsFile() && documentURI.Scheme() != "vscode-remote" {
		return "", false
	}
	directory, err := uri.JoinPath(documentURI, "..")
	return directory, err == nil
}

func (s *Store) refreshDirectory(directory uri.URI) []uri.URI {
	entries, err := os.ReadDir(directory.FsPath())
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	candidates := map[uri.URI]bool{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".skel" || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		documentURI, err := uri.JoinPath(directory, entry.Name())
		if err == nil {
			candidates[documentURI] = true
		}
	}
	s.mu.RLock()
	for documentURI := range s.documents {
		if parent, ok := sourceDirectory(documentURI); ok && parent == directory {
			candidates[documentURI] = true
		}
	}
	s.mu.RUnlock()
	changed := []uri.URI{}
	for documentURI := range candidates {
		s.mu.RLock()
		previous, opened := s.documents[documentURI], s.open[documentURI]
		s.mu.RUnlock()
		if opened {
			continue
		}
		content, err := os.ReadFile(documentURI.FsPath())
		if err != nil && !os.IsNotExist(err) {
			continue
		}
		var document *Document
		if err == nil {
			if previous != nil && previous.Source == string(content) {
				document = previous
			} else {
				document = BuildDocument(documentURI, documentURI.FsPath(), string(content), 0)
			}
		}
		s.mu.Lock()
		if !s.open[documentURI] && s.documents[documentURI] == previous {
			if document == nil {
				delete(s.documents, documentURI)
				s.untrackLocked(documentURI)
			} else {
				s.trackLocked(documentURI)
				s.documents[documentURI] = document
			}
			if document != previous {
				changed = append(changed, documentURI)
				s.revision++
				s.cached = nil
			}
		}
		s.mu.Unlock()
	}
	slices.Sort(changed)
	return changed
}

// AddRoot discovers and indexes Skel files below a workspace root.
func (s *Store) AddRoot(rootURI uri.URI) {
	s.diskMu.Lock()
	defer s.diskMu.Unlock()
	rootPath := rootURI.FsPath()
	documents := map[uri.URI]*Document{}
	_ = filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() && path != rootPath && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || filepath.Ext(path) != ".skel" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		documentURI := documentURI(rootURI, rootPath, path)
		documents[documentURI] = BuildDocument(documentURI, path, string(content), 0)
		return nil
	})

	s.mu.Lock()
	defer s.mu.Unlock()
	tracked := make(map[uri.URI]struct{}, len(documents))
	for documentURI, document := range documents {
		tracked[documentURI] = struct{}{}
		if !s.open[documentURI] {
			s.documents[documentURI] = document
		}
	}
	s.workspaceFiles[rootURI] = tracked
	s.revision++
	s.cached = nil
}

// RemoveRoot removes documents that are no longer open or tracked by another
// workspace root and returns the removed URIs.
func (s *Store) RemoveRoot(rootURI uri.URI) []uri.URI {
	s.mu.Lock()
	defer s.mu.Unlock()
	tracked := s.workspaceFiles[rootURI]
	delete(s.workspaceFiles, rootURI)
	removed := make([]uri.URI, 0, len(tracked))
	for documentURI := range tracked {
		if s.open[documentURI] || s.documentTrackedLocked(documentURI) {
			continue
		}
		delete(s.documents, documentURI)
		removed = append(removed, documentURI)
	}
	s.revision++
	s.cached = nil
	return removed
}

// Snapshot returns an immutable view of the current workspace.
func (s *Store) Snapshot() Snapshot {
	for {
		s.mu.RLock()
		if s.cached != nil {
			snapshot := *s.cached
			s.mu.RUnlock()
			return snapshot
		}
		revision := s.revision
		documents := make(map[uri.URI]*Document, len(s.documents))
		ordered := make([]*Document, 0, len(s.documents))
		for documentURI, document := range s.documents {
			documents[documentURI] = document
			ordered = append(ordered, document)
		}
		s.mu.RUnlock()
		slices.SortFunc(ordered, func(a, b *Document) int { return strings.Compare(string(a.URI), string(b.URI)) })
		snapshot := newSnapshot(revision, documents, ordered)
		s.mu.Lock()
		if s.revision == revision {
			s.cached = &snapshot
			s.mu.Unlock()
			return snapshot
		}
		s.mu.Unlock()
	}
}

func (s *Store) trackLocked(documentURI uri.URI) {
	for rootURI, tracked := range s.workspaceFiles {
		if contains(rootURI, documentURI) {
			tracked[documentURI] = struct{}{}
		}
	}
}

func (s *Store) untrackLocked(documentURI uri.URI) {
	for _, tracked := range s.workspaceFiles {
		delete(tracked, documentURI)
	}
}

func (s *Store) documentTrackedLocked(documentURI uri.URI) bool {
	for _, tracked := range s.workspaceFiles {
		if _, ok := tracked[documentURI]; ok {
			return true
		}
	}
	return false
}

func documentURI(rootURI uri.URI, rootPath, path string) uri.URI {
	if rootURI.IsFile() {
		return uri.File(path)
	}
	relative, err := filepath.Rel(rootPath, path)
	if err != nil {
		return uri.File(path)
	}
	documentURI, err := uri.JoinPath(rootURI, filepath.ToSlash(relative))
	if err != nil {
		return uri.File(path)
	}
	return documentURI
}

func contains(rootURI, documentURI uri.URI) bool {
	if rootURI.Scheme() != documentURI.Scheme() || rootURI.Authority() != documentURI.Authority() {
		return false
	}
	relative, err := filepath.Rel(rootURI.FsPath(), documentURI.FsPath())
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// Revision reads the current generation without constructing a workspace index.
func (s *Store) Revision() uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.revision }

// Document reads a single immutable document without constructing a snapshot.
func (s *Store) Document(documentURI uri.URI) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.documents[documentURI]
}
