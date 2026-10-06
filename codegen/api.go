// Package codegen is the Go SDK for implementing Skel language bindings.
// Input selects output declarations from a validated semantic model. Generators
// return files, keeping target names and imports separate from semantic data.
// Model values borrowed from Input are read-only for its lifetime.
package codegen

import (
	"context"
	internal "go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/model"
)

type Input = internal.Input
type Selection = internal.Selection
type Surface = internal.Surface
type ApiFilter = internal.ApiFilter
type Declarations = internal.Declarations
type File = internal.File
type Generator = internal.Generator
type GeneratorFunc = internal.GeneratorFunc
type TypeVisitor = internal.TypeVisitor

const (
	SurfaceFull         = internal.SurfaceFull
	SurfacePublic       = internal.SurfacePublic
	SurfaceAPI          = internal.SurfaceAPI
	GeneratedFileMarker = internal.GeneratedFileMarker
)

// Prepare validates renderer invariants and selects declarations from a resolved
// semantic domain. The model is borrowed read-only; declaration types are not copied.
func Prepare(domain *model.Domain, selection Selection) (Input, error) {
	return internal.Prepare(domain, selection)
}

// Generate returns marked, sorted files without writing to output directories.
func Generate(ctx context.Context, in Input, generator Generator) ([]File, error) {
	return internal.Generate(ctx, in, generator)
}

// Run generates and publishes all named target directories in one managed operation.
// Generated paths replace existing files, stale marked files are removed, and
// unrelated files are preserved. Target directories must not overlap.
func Run(ctx context.Context, in Input, generator Generator, targets map[string]string) error {
	return internal.Run(ctx, in, generator, targets)
}

// WriteFiles publishes an already rendered file set with the same checks as Run.
func WriteFiles(ctx context.Context, files []File, targets map[string]string) error {
	return internal.WriteFiles(ctx, files, targets)
}

// WalkType visits structural children without following named declarations.
func WalkType(kind *model.Type, visit TypeVisitor) error { return internal.WalkType(kind, visit) }

// WalkTypes visits several structural roots, deduplicating shared type nodes.
func WalkTypes(kinds []*model.Type, visit TypeVisitor) error { return internal.WalkTypes(kinds, visit) }

// WalkTypeGraph also follows named data members and terminates on recursive types.
func WalkTypeGraph(kind *model.Type, visit TypeVisitor) error {
	return internal.WalkTypeGraph(kind, visit)
}

// WalkTypeGraphs follows named data members from several roots.
func WalkTypeGraphs(kinds []*model.Type, visit TypeVisitor) error {
	return internal.WalkTypeGraphs(kinds, visit)
}

// InstantiateMembers substitutes generic arguments in a data reference
// without recursively expanding named declarations or changing the model.
func InstantiateMembers(kind *model.Type) ([]*model.DataMember, error) {
	return internal.InstantiateMembers(kind)
}
