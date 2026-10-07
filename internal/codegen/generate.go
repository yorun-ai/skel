// Package codegen implements the shared Go generator SDK. It does not depend on
// a compiler, CLI or target language. Generators return files; this package owns
// validation and managed filesystem publication.
package codegen

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skel/internal/codegen/output"
)

const GeneratedFileMarker = output.GeneratedFileMarker

// Generator is implemented by Go-written bindings for any output language.
// Generate treats the input and its reachable semantic objects as read-only,
// returns files without publishing them, and observes ctx during long operations.
type Generator interface {
	Generate(context.Context, Input) ([]File, error)
}
type GeneratorFunc func(context.Context, Input) ([]File, error)

func (f GeneratorFunc) Generate(ctx context.Context, in Input) ([]File, error) {
	if f == nil {
		return nil, fmt.Errorf("codegen generator function is nil")
	}
	return f(ctx, in)
}

// Generate renders and validates the complete file set without touching disk.
// Results are marked for ownership and sorted by target and relative path.
func Generate(ctx context.Context, in Input, generator Generator) ([]File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !in.Valid() {
		return nil, fmt.Errorf("codegen input is uninitialized")
	}
	if generator == nil {
		return nil, fmt.Errorf("codegen generator is nil")
	}
	files, err := generator.Generate(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return prepareFiles(files)
}

// Run generates then publishes all outputs together. An empty target name is
// conventional for the primary output. Include targets with no generated files
// to remove stale owned files. Unrelated files are retained.
func Run(ctx context.Context, in Input, generator Generator, targets map[string]string) error {
	files, err := Generate(ctx, in, generator)
	if err != nil {
		return err
	}
	return WriteFiles(ctx, files, targets)
}

// WriteFiles validates all files before staging and commits all target directories
// through the same transaction used by built-in generators. Existing files at
// generated paths are replaced; obsolete marked files are removed.
func WriteFiles(ctx context.Context, files []File, targets map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	files, err := prepareFiles(files)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(targets))
	for name, dir := range targets {
		if strings.TrimSpace(dir) == "" {
			return fmt.Errorf("empty output directory for target %q", name)
		}
		names = append(names, name)
	}
	slices.Sort(names)
	indices := map[string]int{}
	dirs := make([]string, 0, len(names))
	for i, name := range names {
		indices[name] = i
		dir, err := filepath.Abs(targets[name])
		if err != nil {
			return err
		}
		for _, other := range dirs {
			if containsDirectory(dir, other) || containsDirectory(other, dir) {
				return fmt.Errorf("%w: overlapping output directories %s and %s", output.ErrOutputOperation, other, dir)
			}
		}
		dirs = append(dirs, dir)
	}
	for _, file := range files {
		if _, ok := indices[file.Target]; !ok {
			return fmt.Errorf("no output directory for target %q", file.Target)
		}
	}
	return output.RunManagedOutputs(dirs, func(staged []string) error {
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return err
			}
			dest := filepath.Join(staged[indices[file.Target]], filepath.FromSlash(file.Path))
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(dest, []byte(file.Content), 0o644); err != nil {
				return err
			}
		}
		return ctx.Err()
	})
}

func prepareFiles(files []File) ([]File, error) {
	result := slices.Clone(files)
	seen := map[string]bool{}
	for i := range result {
		f := &result[i]
		if f.Path == "" || f.Path == "." || !filepath.IsLocal(f.Path) || path.Clean(f.Path) != f.Path || strings.ContainsAny(f.Path, "\\:\x00") {
			return nil, fmt.Errorf("invalid generated path %q", f.Path)
		}
		key := f.Target + "\x00" + f.Path
		if strings.ContainsRune(f.Target, '\x00') || seen[key] {
			return nil, fmt.Errorf("duplicate or invalid generated path %q in target %q", f.Path, f.Target)
		}
		seen[key] = true
		content := strings.TrimRight(f.Content, "\n") + "\n"
		if f.CommentPrefix != "" {
			if !slices.Contains([]string{"//", "#", "--", ";", "%"}, f.CommentPrefix) {
				return nil, fmt.Errorf("unsupported generated comment prefix %q", f.CommentPrefix)
			}
			if !output.HasGeneratedFileMarker([]byte(content)) {
				content = f.CommentPrefix + " " + GeneratedFileMarker + "\n\n" + content
			}
		}
		marked, err := output.MarkGeneratedFile(f.Path, content)
		if err != nil {
			return nil, err
		}
		f.Content = marked
	}
	for _, f := range result {
		for parent := path.Dir(f.Path); parent != "."; parent = path.Dir(parent) {
			if seen[f.Target+"\x00"+parent] {
				return nil, fmt.Errorf("generated file %q is also a directory", parent)
			}
		}
	}
	slices.SortFunc(result, func(a, b File) int {
		if n := strings.Compare(a.Target, b.Target); n != 0 {
			return n
		}
		return strings.Compare(a.Path, b.Path)
	})
	return result, nil
}
func containsDirectory(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && filepath.IsLocal(relative)
}
