package loader

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go.yorun.ai/skel/internal/source"
)

// Git reads blobs from a pinned commit. It never extracts files or follows
// repository symlinks, and filenames are parsed from NUL-delimited tree records.
type Git struct {
	root, commit string
	entries      map[string]Entry
	blobs        map[string]string
}

func NewGit(ctx context.Context, root, commit, target string) (*Git, error) {
	relative, err := filepath.Rel(root, target)
	if err != nil || gitPathEscapesRoot(filepath.ToSlash(relative)) {
		return nil, fmt.Errorf("Git input %s is outside %s", target, root)
	}
	args := []string{"ls-tree", "-r", "-z", commit}
	if relative != "." {
		args = append(args, "--", filepath.ToSlash(relative))
	}
	output, err := GitBytes(ctx, root, args...)
	if err != nil {
		return nil, err
	}
	snapshot := new(Git{root: root, commit: commit, entries: map[string]Entry{}, blobs: map[string]string{}})
	for _, record := range strings.Split(string(output), "\x00") {
		header, name, ok := strings.Cut(record, "\t")
		if !ok {
			continue
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		snapshot.entries[path] = Entry{Name: filepath.Base(path)}
		snapshot.blobs[path] = fields[2]
		for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
			snapshot.entries[parent] = Entry{Name: filepath.Base(parent), Directory: true}
			if parent == root || parent == filepath.Dir(parent) {
				break
			}
		}
	}
	return snapshot, nil
}
func (g *Git) Stat(ctx context.Context, path string) (Entry, error) {
	if err := ctx.Err(); err != nil {
		return Entry{}, err
	}
	entry, ok := g.entries[filepath.Clean(path)]
	if !ok {
		return Entry{}, &fs.PathError{Op: "stat", Path: path, Err: fs.ErrNotExist}
	}
	return entry, nil
}
func (g *Git) ReadDir(ctx context.Context, path string) ([]Entry, error) {
	entry, err := g.Stat(ctx, path)
	if err != nil {
		return nil, err
	}
	if !entry.Directory {
		return nil, fmt.Errorf("%s is not a directory", path)
	}
	result := []Entry{}
	path = filepath.Clean(path)
	for name, entry := range g.entries {
		if name != path && filepath.Dir(name) == path {
			result = append(result, entry)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
func (g *Git) Read(ctx context.Context, path string) (*source.Document, error) {
	path = filepath.Clean(path)
	blob := g.blobs[path]
	if blob == "" {
		return nil, &fs.PathError{Op: "read", Path: path, Err: fs.ErrNotExist}
	}
	content, err := GitBytes(ctx, g.root, "cat-file", "blob", blob)
	if err != nil {
		return nil, err
	}
	relative, _ := filepath.Rel(g.root, path)
	return source.New(source.ID("git:"+g.root+"@"+g.commit+":"+filepath.ToSlash(relative)), path, 0, string(content)), nil
}

func GitBytes(ctx context.Context, directory string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", directory, "--literal-pathspecs"}, args...)...)
	content, err := command.Output()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err == nil {
		return content, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if message := strings.TrimSpace(string(exitError.Stderr)); message != "" {
			return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
		}
	}
	return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
}

func gitPathEscapesRoot(path string) bool {
	return path == ".." || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/")
}
