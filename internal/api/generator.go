package api

import (
	"go.yorun.ai/skel/codegen"
	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/binding/skeleton"
	"go.yorun.ai/skel/internal/codegen/binding/typescript"
)

func NewGolangGenerator(option GolangOption) (codegen.Generator, error) {
	return golang.NewGenerator(option)
}
func NewTypeScriptGenerator(option TypeScriptOption) (codegen.Generator, error) {
	return typescript.NewGenerator(option)
}
func NewSkeletonGenerator(option SkeletonOption) (codegen.Generator, error) {
	return skeleton.NewGenerator(option)
}
