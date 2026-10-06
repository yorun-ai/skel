package binding

import "go.yorun.ai/skel/internal/model"

// ImportBinding belongs to one generation target, never to the semantic model.
type ImportBinding struct {
	Domain   string
	Alias    string
	Explicit bool
	Path     string
}

// TypeBindings is built before rendering and is read-only during generation.
type TypeBindings map[*model.Type]*ImportBinding
