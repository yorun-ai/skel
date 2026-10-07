package binding

import "go.yorun.ai/skel/schema"

// ImportBinding belongs to one generation target, never to the semantic schema.
type ImportBinding struct {
	Domain   string
	Alias    string
	Explicit bool
	Path     string
}

// TypeBindings is built before rendering and is read-only during generation.
type TypeBindings map[*schema.Type]*ImportBinding
