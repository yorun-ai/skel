package api

import (
	"go.yorun.ai/skel/codegen"
	internal "go.yorun.ai/skel/internal/api"
)

// NewGolangGenerator returns the built-in Go binding for use with codegen.Generate
// or codegen.Run. It selects the surface configured in option. Out is used for
// package naming, but rendering does not write there. Files target "" or, for
// split public output, "pub"; codegen.Run supplies their actual destinations.
func NewGolangGenerator(option GolangOption) (codegen.Generator, error) {
	return internal.NewGolangGenerator(option)
}

// NewTypeScriptGenerator returns the built-in client binding. It selects the API
// surface configured in option; generated files use the primary target "".
func NewTypeScriptGenerator(option TypeScriptOption) (codegen.Generator, error) {
	return internal.NewTypeScriptGenerator(option)
}

// NewSkeletonGenerator returns the built-in public-contract binding. Generated
// files use the primary target "". Construction and rendering do not write files.
func NewSkeletonGenerator(option SkeletonOption) (codegen.Generator, error) {
	return internal.NewSkeletonGenerator(option)
}
