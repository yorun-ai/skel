package binding

import (
	"context"

	"go.yorun.ai/skel/internal/codegen"
)

// FileSink lets renderers produce files without owning filesystem operations.
type FileSink func(path, content string) error

// FileCollector accumulates one generation invocation, including multiple targets.
// It is local to the invocation; renderers do not write to it concurrently.
type FileCollector struct {
	Context context.Context
	Files   []codegen.File
}

func (c *FileCollector) Sink(target string) FileSink {
	return func(path, content string) error {
		if c.Context != nil {
			if err := c.Context.Err(); err != nil {
				return err
			}
		}
		c.Files = append(c.Files, codegen.File{Target: target, Path: path, Content: content})
		return nil
	}
}
