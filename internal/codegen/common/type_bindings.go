package common

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

// Type adapts a semantic type for existing renderers without changing the model.
// Nested types keep their identities and are adapted as they are rendered.
func (b TypeBindings) Type(kind *model.Type) *model.Type {
	binding := b[kind]
	if binding == nil {
		return kind
	}
	adapted := *kind
	adapted.ExternalImportPath = binding.Path
	adapted.ExternalAlias = binding.Alias
	return &adapted
}
