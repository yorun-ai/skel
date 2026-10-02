package compiler

import (
	"context"
	"errors"

	"go.yorun.ai/skelc/internal/analyzer"
	"go.yorun.ai/skelc/internal/loader"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

func parseFileWithImports(file *loader.SourceFile, imports []*analyzer.Analysis) (*analyzer.Analysis, error) {
	return analyzeFixture(loader.Result{Files: []*loader.SourceFile{file}}, imports)
}
func parseDomainFilesWithImports(_ *loader.SourceFile, files []*loader.SourceFile, imports []*analyzer.Analysis) (*analyzer.Analysis, error) {
	return analyzeFixture(loader.Result{Files: files, IsDir: true}, imports)
}
func analyzeFixture(loaded loader.Result, imports []*analyzer.Analysis) (*analyzer.Analysis, error) {
	sources, err := prepareInput(context.Background(), loaded, false)
	if err != nil {
		return nil, err
	}
	contents := []*grammar.SkelContent{}
	for _, source := range sources {
		contents = append(contents, source.Parsed)
	}
	result, failures := analyzer.Analyze(mergeDomainContents(contents), imports)
	return result, errors.Join(failures...)
}
