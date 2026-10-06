package compiler

import (
	"context"
	"fmt"

	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/internal/parser/grammar"
)

func parseContentContext(ctx context.Context, sourceFile *loader.SourceFile) (*grammar.SkelContent, error) {
	parsed, err := parser.ParseSourceContext(ctx, sourceFile.FilePath, sourceFile.Content)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("parse %s failed: %w", sourceFile.FilePath, err)
	}

	content := parsed.Content
	if content.Domain == nil || content.Domain.Name == nil {
		return nil, fmt.Errorf("missing domain declaration in %s", sourceFile.FilePath)
	}
	if content.Domain.Name.String() == "" {
		return nil, fmt.Errorf("missing domain name in %s", sourceFile.FilePath)
	}
	return content, nil
}
