package analysis

import (
	"path/filepath"

	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/lsp/source"
)

// Build line indexes only for files that have diagnostic locations. A clean
// workspace does not need another copy or scan of every source document.
type _SourceBuffers struct {
	contents map[string][]byte
	indexed  map[string]source.Buffer
}

func newSourceBuffers(sources []compiler.Source) *_SourceBuffers {
	buffers := new(_SourceBuffers{contents: make(map[string][]byte, len(sources)), indexed: map[string]source.Buffer{}})
	for _, input := range sources {
		buffers.contents[filepath.Clean(input.Path)] = input.Content
	}
	return buffers
}

func (b *_SourceBuffers) get(path string) source.Buffer {
	path = filepath.Clean(path)
	if buffer, ok := b.indexed[path]; ok {
		return buffer
	}
	buffer := source.New(string(b.contents[path]))
	b.indexed[path] = buffer
	return buffer
}
