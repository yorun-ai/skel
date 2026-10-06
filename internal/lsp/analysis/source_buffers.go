package analysis

import (
	"path/filepath"

	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/lsp/source"
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
		path := filepath.Clean(input.Path)
		if input.Document != nil {
			buffers.indexed[path] = source.FromDocument(input.Document)
		} else {
			buffers.contents[path] = input.Content
		}
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
