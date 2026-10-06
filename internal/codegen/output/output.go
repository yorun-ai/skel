// Package output commits generated files through managed multi-target transactions.
package output

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/util/fileutil"
)

const (
	generatedFileMarkerScanLimit = 4 * 1024
)

// ManagedOutput stages a complete generator run and commits files atomically.
// Generated-file markers distinguish skelc-owned output from handwritten files.
type ManagedOutput struct {
	targetDir string
	stageRoot string
	stageDir  string
	writeFile func(string, []byte, fs.FileMode) error
}

func NewManagedOutput(targetDir string) (*ManagedOutput, error) {
	targetDir = filepath.Clean(targetDir)
	parent := filepath.Dir(targetDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("create output parent %s: %w", parent, err)
	}
	stageRoot, err := os.MkdirTemp(parent, "."+filepath.Base(targetDir)+".skelc-stage-")
	if err != nil {
		return nil, fmt.Errorf("create output staging directory for %s: %w", targetDir, err)
	}
	stageDir := filepath.Join(stageRoot, filepath.Base(targetDir))
	if err := os.Mkdir(stageDir, 0o755); err != nil {
		_ = os.RemoveAll(stageRoot)
		return nil, fmt.Errorf("create output staging directory %s: %w", stageDir, err)
	}
	return &ManagedOutput{
		targetDir: targetDir,
		stageRoot: stageRoot,
		stageDir:  stageDir,
		writeFile: atomicWriteFile,
	}, nil
}

func (o *ManagedOutput) StageDir() string {
	return o.stageDir
}

func (o *ManagedOutput) Abort() {
	if o == nil || o.stageRoot == "" {
		return
	}
	_ = os.RemoveAll(o.stageRoot)
	o.stageRoot = ""
}

func collectStagedGeneratedOutputFiles(stageDir string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(stageDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("generated output contains symlink %s", path)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("generated output path %s is not a regular file", path)
		}
		relative, err := filepath.Rel(stageDir, path)
		if err != nil {
			return err
		}
		relative, err = cleanGeneratedOutputPath(relative)
		if err != nil {
			return err
		}
		marked, err := generatedFileHasMarker(path)
		if err != nil {
			return err
		}
		if !marked {
			return fmt.Errorf("generated output %s is missing the skelc ownership marker", path)
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inspect staged output %s: %w", stageDir, err)
	}
	slices.Sort(files)
	return files, nil
}

func collectMarkedGeneratedOutputFiles(targetDir string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(targetDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		marked, err := generatedFileHasMarker(path)
		if err != nil {
			return err
		}
		if !marked {
			return nil
		}
		relative, err := filepath.Rel(targetDir, path)
		if err != nil {
			return err
		}
		relative, err = cleanGeneratedOutputPath(relative)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inspect generated output directory %s: %w", targetDir, err)
	}
	slices.Sort(files)
	return files, nil
}

func generatedFileHasMarker(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	return generatedFileMarkerInReader(file)
}

func generatedFileMarkerInReader(reader io.Reader) (bool, error) {
	prefix, err := io.ReadAll(io.LimitReader(reader, generatedFileMarkerScanLimit))
	if err != nil {
		return false, err
	}
	return common.HasGeneratedFileMarker(prefix), nil
}

func (o *ManagedOutput) commitOutputFile(relative string) error {
	relative, err := cleanGeneratedOutputPath(filepath.FromSlash(relative))
	if err != nil {
		return err
	}
	source := filepath.Join(o.stageDir, relative)
	target := filepath.Join(o.targetDir, relative)
	if err := rejectOutputSymlink(o.targetDir, relative); err != nil {
		return err
	}
	content, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read staged output %s: %w", source, err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", target, err)
	}
	if err := o.writeFile(target, content, 0o644); err != nil {
		return fmt.Errorf("commit generated output %s: %w", target, err)
	}
	return nil
}

func collectStaleGeneratedOutputFiles(
	current []string,
	marked []string,
) []string {
	currentPaths := make(map[string]bool, len(current))
	for _, path := range current {
		currentPaths[path] = true
	}
	stalePaths := map[string]bool{}
	for _, path := range marked {
		if !currentPaths[path] {
			stalePaths[path] = true
		}
	}
	return slices.Sorted(maps.Keys(stalePaths))
}

func removeGeneratedOutputFiles(targetDir string, paths []string) error {
	for _, relative := range paths {
		relative, err := cleanGeneratedOutputPath(filepath.FromSlash(relative))
		if err != nil {
			return err
		}
		if err := rejectOutputSymlink(targetDir, relative); err != nil {
			return err
		}
		path := filepath.Join(targetDir, relative)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove stale generated output %s: %w", path, err)
		}
		removeEmptyOutputParents(filepath.Dir(path), targetDir)
	}
	return nil
}

func rejectOutputSymlink(root, relative string) error {
	path := root
	parts := []string{}
	if relative != "" {
		parts = strings.Split(filepath.Clean(relative), string(filepath.Separator))
	}
	for index := -1; index < len(parts); index++ {
		if index >= 0 {
			path = filepath.Join(path, parts[index])
		}
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect output path %s: %w", path, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output path contains symlink %s", path)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("output parent %s is not a directory", path)
		}
	}
	return nil
}

func removeEmptyOutputParents(dir, root string) {
	for dir != root && strings.HasPrefix(dir, root+string(filepath.Separator)) {
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func cleanGeneratedOutputPath(path string) (string, error) {
	cleaned := filepath.Clean(path)
	if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe generated output path %q", path)
	}
	return cleaned, nil
}

func atomicWriteFile(path string, content []byte, mode fs.FileMode) error {
	return fileutil.Replace(fileutil.Replacement{Path: path, Content: content, Mode: mode})
}
