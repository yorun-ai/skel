package output

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

type _OutputSnapshot struct {
	path       string
	backupPath string
	mode       fs.FileMode
	existed    bool
}

type _OutputCommit struct {
	output        *ManagedOutput
	files         []string
	staleFiles    []string
	snapshots     []_OutputSnapshot
	targetExisted bool
}

// ErrOutputOperation identifies staging or commit failures outside generator
// compilation.
var ErrOutputOperation = errors.New("generated output operation failed")

func outputOperationError(err error) error {
	if err == nil || errors.Is(err, ErrOutputOperation) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrOutputOperation, err)
}

// RunManagedOutputs stages every non-empty target, runs generate with matching
// staging paths, and commits all outputs as one transaction. A generation or
// commit failure leaves target outputs unchanged.
func RunManagedOutputs(targets []string, generate func([]string) error) error {
	stagePaths := make([]string, len(targets))
	outputs := make([]*ManagedOutput, 0, len(targets))
	defer func() {
		for _, output := range outputs {
			output.Abort()
		}
	}()
	for index, target := range targets {
		if target == "" {
			continue
		}
		output, err := NewManagedOutput(target)
		if err != nil {
			return outputOperationError(err)
		}
		outputs = append(outputs, output)
		stagePaths[index] = output.StageDir()
	}
	if err := generate(stagePaths); err != nil {
		return err
	}
	return outputOperationError(CommitManagedOutputs(outputs))
}

func (o *ManagedOutput) Commit() error {
	if o == nil || o.stageRoot == "" {
		return errors.New("output staging transaction is closed")
	}
	defer o.Abort()
	commit, err := o.prepareCommit()
	if err != nil {
		return err
	}
	if err := commit.apply(); err != nil {
		if rollbackErr := commit.rollback(); rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("roll back generated output: %w", rollbackErr))
		}
		return err
	}
	return nil
}

// CommitManagedOutputs commits several staged outputs as one transaction. If
// any target fails, every target already modified by the batch is restored.
func CommitManagedOutputs(outputs []*ManagedOutput) error {
	commits := make([]*_OutputCommit, 0, len(outputs))
	targets := map[string]bool{}
	defer func() {
		for _, output := range outputs {
			output.Abort()
		}
	}()
	for _, output := range outputs {
		if output == nil || output.stageRoot == "" {
			return errors.New("output staging transaction is closed")
		}
		if targets[output.targetDir] {
			return fmt.Errorf("cannot commit output directory %s more than once", output.targetDir)
		}
		targets[output.targetDir] = true
		commit, err := output.prepareCommit()
		if err != nil {
			rollbackErrors := []error{}
			for index := len(commits) - 1; index >= 0; index-- {
				if rollbackErr := commits[index].rollback(); rollbackErr != nil {
					rollbackErrors = append(rollbackErrors, rollbackErr)
				}
			}
			if len(rollbackErrors) > 0 {
				return errors.Join(err, fmt.Errorf("clean up prepared outputs: %w", errors.Join(rollbackErrors...)))
			}
			return err
		}
		commits = append(commits, commit)
	}
	for index, commit := range commits {
		if err := commit.apply(); err != nil {
			rollbackErrors := []error{}
			for rollbackIndex := index; rollbackIndex >= 0; rollbackIndex-- {
				if rollbackErr := commits[rollbackIndex].rollback(); rollbackErr != nil {
					rollbackErrors = append(rollbackErrors, rollbackErr)
				}
			}
			if len(rollbackErrors) > 0 {
				return errors.Join(err, fmt.Errorf("roll back generated outputs: %w", errors.Join(rollbackErrors...)))
			}
			return err
		}
	}
	return nil
}

func (o *ManagedOutput) prepareCommit() (*_OutputCommit, error) {
	files, err := collectStagedGeneratedOutputFiles(o.stageDir)
	if err != nil {
		return nil, err
	}
	targetExisted := true
	if _, err := os.Lstat(o.targetDir); errors.Is(err, os.ErrNotExist) {
		targetExisted = false
	} else if err != nil {
		return nil, fmt.Errorf("inspect output directory %s: %w", o.targetDir, err)
	}
	if err := os.MkdirAll(o.targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory %s: %w", o.targetDir, err)
	}
	cleanupNewTarget := func() {
		if !targetExisted {
			_ = os.Remove(o.targetDir)
		}
	}
	if err := rejectOutputSymlink(o.targetDir, ""); err != nil {
		cleanupNewTarget()
		return nil, err
	}
	markedFiles, err := collectMarkedGeneratedOutputFiles(o.targetDir)
	if err != nil {
		cleanupNewTarget()
		return nil, err
	}
	staleFiles := collectStaleGeneratedOutputFiles(files, markedFiles)
	changedFiles := make([]string, 0, len(files))
	paths := map[string]bool{}
	for _, file := range append(slices.Clone(files), staleFiles...) {
		if err := rejectOutputSymlink(o.targetDir, filepath.FromSlash(file)); err != nil {
			cleanupNewTarget()
			return nil, err
		}
	}
	for _, file := range files {
		target := filepath.Join(o.targetDir, filepath.FromSlash(file))
		info, err := os.Lstat(target)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupNewTarget()
			return nil, fmt.Errorf("inspect output %s: %w", target, err)
		}
		if err == nil {
			if !info.Mode().IsRegular() {
				cleanupNewTarget()
				return nil, fmt.Errorf("output path %s is not a regular file", target)
			}
			current, readErr := os.ReadFile(target)
			if readErr != nil {
				cleanupNewTarget()
				return nil, fmt.Errorf("read output %s: %w", target, readErr)
			}
			staged, readErr := os.ReadFile(filepath.Join(o.stageDir, filepath.FromSlash(file)))
			if readErr != nil {
				cleanupNewTarget()
				return nil, fmt.Errorf("read staged output %s: %w", file, readErr)
			}
			if bytes.Equal(current, staged) {
				continue
			}
		}
		changedFiles = append(changedFiles, file)
		paths[file] = true
	}
	for _, file := range staleFiles {
		paths[file] = true
	}
	files = changedFiles
	snapshots, err := snapshotOutputFiles(o.targetDir, o.stageRoot, paths)
	if err != nil {
		cleanupNewTarget()
		return nil, err
	}
	return &_OutputCommit{
		output: o, files: files, staleFiles: staleFiles,
		snapshots:     snapshots,
		targetExisted: targetExisted,
	}, nil
}

func (commit *_OutputCommit) apply() error {
	for _, file := range commit.files {
		if err := commit.output.commitOutputFile(file); err != nil {
			return err
		}
	}
	return removeGeneratedOutputFiles(commit.output.targetDir, commit.staleFiles)
}

func (commit *_OutputCommit) rollback() error {
	errs := []error{}
	for index := len(commit.snapshots) - 1; index >= 0; index-- {
		snapshot := commit.snapshots[index]
		target := filepath.Join(commit.output.targetDir, filepath.FromSlash(snapshot.path))
		if !snapshot.existed {
			if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
				errs = append(errs, fmt.Errorf("remove new output %s: %w", target, err))
				continue
			}
			removeEmptyOutputParents(filepath.Dir(target), commit.output.targetDir)
			continue
		}
		content, err := os.ReadFile(snapshot.backupPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("read output backup %s: %w", snapshot.backupPath, err))
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			errs = append(errs, fmt.Errorf("restore output parent %s: %w", target, err))
			continue
		}
		if err := atomicWriteFile(target, content, snapshot.mode); err != nil {
			errs = append(errs, fmt.Errorf("restore output %s: %w", target, err))
		}
	}
	if !commit.targetExisted {
		_ = os.Remove(commit.output.targetDir)
	}
	return errors.Join(errs...)
}

func snapshotOutputFiles(targetDir, stageRoot string, paths map[string]bool) ([]_OutputSnapshot, error) {
	relatives := slices.Sorted(maps.Keys(paths))
	backupRoot := filepath.Join(stageRoot, "backup")
	snapshots := make([]_OutputSnapshot, 0, len(relatives))
	for index, relative := range relatives {
		target := filepath.Join(targetDir, filepath.FromSlash(relative))
		info, err := os.Lstat(target)
		if errors.Is(err, os.ErrNotExist) {
			snapshots = append(snapshots, _OutputSnapshot{path: relative})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect output %s: %w", target, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("output path %s is not a regular file", target)
		}
		content, err := os.ReadFile(target)
		if err != nil {
			return nil, fmt.Errorf("read output %s: %w", target, err)
		}
		backupPath := filepath.Join(backupRoot, fmt.Sprintf("%d", index))
		if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
			return nil, fmt.Errorf("create output backup directory: %w", err)
		}
		if err := os.WriteFile(backupPath, content, info.Mode().Perm()); err != nil {
			return nil, fmt.Errorf("back up output %s: %w", target, err)
		}
		snapshots = append(snapshots, _OutputSnapshot{
			path: relative, backupPath: backupPath, mode: info.Mode().Perm(), existed: true,
		})
	}
	return snapshots, nil
}
