package loader

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"go.yorun.ai/skel/internal/source"
)

// Entry is the portion of filesystem metadata needed to discover compiler inputs.
type Entry struct {
	Name      string
	Directory bool
}

// Provider supplies immutable revisions through one discovery contract. Paths
// are absolute logical paths; a provider need not access the working tree.
type Provider interface {
	Stat(context.Context, string) (Entry, error)
	ReadDir(context.Context, string) ([]Entry, error)
	Read(context.Context, string) (*source.Document, error)
}

type FileSystem struct{}

func (FileSystem) Stat(ctx context.Context, path string) (Entry, error) {
	if err := ctx.Err(); err != nil {
		return Entry{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Name: info.Name(), Directory: info.IsDir()}, nil
}
func (FileSystem) ReadDir(ctx context.Context, path string) ([]Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, Entry{Name: entry.Name(), Directory: entry.IsDir()})
	}
	return result, nil
}
func (FileSystem) Read(ctx context.Context, path string) (*source.Document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return source.New(source.ID(path), path, 0, string(content)), nil
}

// Memory is a fixed set of revisions with directories inferred from their paths.
// Callers retain no mutable state inside the provider.
type Memory struct {
	documents map[string]*source.Document
	entries   map[string]Entry
}

func NewMemory(documents ...*source.Document) *Memory {
	memory := new(Memory{documents: map[string]*source.Document{}, entries: map[string]Entry{}})
	for _, document := range documents {
		path := filepath.Clean(document.AnalysisPath())
		memory.documents[path] = document
		memory.entries[path] = Entry{Name: filepath.Base(path)}
		for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
			memory.entries[parent] = Entry{Name: filepath.Base(parent), Directory: true}
			if parent == filepath.Dir(parent) {
				break
			}
		}
	}
	return memory
}
func (m *Memory) Stat(ctx context.Context, path string) (Entry, error) {
	if err := ctx.Err(); err != nil {
		return Entry{}, err
	}
	entry, ok := m.entries[filepath.Clean(path)]
	if !ok {
		return Entry{}, &fs.PathError{Op: "stat", Path: path, Err: fs.ErrNotExist}
	}
	return entry, nil
}
func (m *Memory) ReadDir(ctx context.Context, path string) ([]Entry, error) {
	entry, err := m.Stat(ctx, path)
	if err != nil {
		return nil, err
	}
	if !entry.Directory {
		return nil, fmt.Errorf("%s is not a directory", path)
	}
	result := []Entry{}
	path = filepath.Clean(path)
	for name, entry := range m.entries {
		if name != path && filepath.Dir(name) == path {
			result = append(result, entry)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
func (m *Memory) Read(ctx context.Context, path string) (*source.Document, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	document := m.documents[filepath.Clean(path)]
	if document == nil {
		return nil, &fs.PathError{Op: "read", Path: path, Err: fs.ErrNotExist}
	}
	return document, nil
}
