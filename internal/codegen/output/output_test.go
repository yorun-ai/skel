package output

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.yorun.ai/skel/internal/codegen/common"
)

type countingReader struct {
	reader    io.Reader
	bytesRead int
}

func (r *countingReader) Read(buffer []byte) (int, error) {
	count, err := r.reader.Read(buffer)
	r.bytesRead += count
	return count, err
}

func TestRunManagedOutputsStagesAlignedOutputsAndCommits(t *testing.T) {
	root := t.TempDir()
	firstTarget := filepath.Join(root, "regular")
	secondTarget := filepath.Join(root, "public")
	if err := RunManagedOutputs([]string{firstTarget, "", secondTarget}, func(staged []string) error {
		if len(staged) != 3 || staged[0] == "" || staged[1] != "" || staged[2] == "" {
			t.Fatalf("unexpected staged paths: %q", staged)
		}
		if staged[0] == firstTarget || staged[2] == secondTarget {
			t.Fatal("generator received target path instead of staging path")
		}
		writeGeneratedOutputTestFile(t, filepath.Join(staged[0], "generated.go"), "regular")
		writeGeneratedOutputTestFile(t, filepath.Join(staged[2], "generated.go"), "public")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	assertOutputTestContent(t, filepath.Join(firstTarget, "generated.go"), "regular")
	assertOutputTestContent(t, filepath.Join(secondTarget, "generated.go"), "public")
}

func TestRunManagedOutputsAbortsOnGenerationFailure(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	writeOutputTestFile(t, filepath.Join(target, "existing.go"), "existing")
	expected := errors.New("injected generation failure")
	err := RunManagedOutputs([]string{target}, func(staged []string) error {
		writeGeneratedOutputTestFile(t, filepath.Join(staged[0], "new.go"), "new")
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected generation failure, got %v", err)
	}
	assertOutputTestContent(t, filepath.Join(target, "existing.go"), "existing")
	assertOutputTestMissing(t, filepath.Join(target, "new.go"))
}

func TestRunManagedOutputsClassifiesOutputOperations(t *testing.T) {
	blockedParent := filepath.Join(t.TempDir(), "blocked")
	writeOutputTestFile(t, blockedParent, "not a directory")
	err := RunManagedOutputs([]string{filepath.Join(blockedParent, "generated")}, func([]string) error {
		return nil
	})
	if !errors.Is(err, ErrOutputOperation) {
		t.Fatalf("expected output operation failure, got %v", err)
	}
}

func TestManagedOutputPreservesUnmanagedFilesAndRemovesMarkedStaleFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	writeOutputTestFile(t, filepath.Join(target, "user.go"), "user")

	first := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "old.go"), "old")
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "nested", "keep.go"), "first")
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}

	second := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(second.StageDir(), "nested", "keep.go"), "second")
	if err := second.Commit(); err != nil {
		t.Fatal(err)
	}

	assertOutputTestContent(t, filepath.Join(target, "user.go"), "user")
	assertOutputTestContent(t, filepath.Join(target, "nested", "keep.go"), "second")
	assertOutputTestMissing(t, filepath.Join(target, "old.go"))
}

func TestManagedOutputPreservesStaleFileWhenMarkerIsRemoved(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	first := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "old.go"), "generated")
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}
	writeOutputTestFile(t, filepath.Join(target, "old.go"), "user modified")

	second := newOutputTestTransaction(t, target)
	if err := second.Commit(); err != nil {
		t.Fatal(err)
	}

	assertOutputTestContent(t, filepath.Join(target, "old.go"), "user modified")
}

func TestManagedOutputRemovesModifiedStaleFileThatRetainsMarker(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	first := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "old.go"), "generated")
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}
	writeGeneratedOutputTestFile(t, filepath.Join(target, "old.go"), "modified generated content")

	second := newOutputTestTransaction(t, target)
	if err := second.Commit(); err != nil {
		t.Fatal(err)
	}

	assertOutputTestMissing(t, filepath.Join(target, "old.go"))
}

func TestGeneratedFileMarkerScanReadsOnlyPrefix(t *testing.T) {
	content := "// " + common.GeneratedFileMarker + "\n\n" + strings.Repeat("x", generatedFileMarkerScanLimit*2)
	reader := &countingReader{reader: strings.NewReader(content)}
	marked, err := generatedFileMarkerInReader(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !marked {
		t.Fatal("expected generated marker in file prefix")
	}
	if reader.bytesRead > generatedFileMarkerScanLimit {
		t.Fatalf("marker scan read %d bytes, limit is %d", reader.bytesRead, generatedFileMarkerScanLimit)
	}
}

func TestManagedOutputRejectsStagedFileWithoutMarker(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	output := newOutputTestTransaction(t, target)
	writeOutputTestFile(t, filepath.Join(output.StageDir(), "new.go"), "new")
	if err := output.Commit(); err == nil || !strings.Contains(err.Error(), "missing the skelc ownership marker") {
		t.Fatalf("expected missing marker error, got %v", err)
	}
	assertOutputTestMissing(t, target)
}

func TestManagedOutputAbortLeavesTargetUnchanged(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	writeOutputTestFile(t, filepath.Join(target, "existing.go"), "existing")
	output := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(output.StageDir(), "new.go"), "new")
	output.Abort()

	assertOutputTestContent(t, filepath.Join(target, "existing.go"), "existing")
	assertOutputTestMissing(t, filepath.Join(target, "new.go"))
}

func TestManagedOutputRejectsSymlinkedTargetParent(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "generated")
	outside := filepath.Join(root, "outside")
	writeOutputTestFile(t, filepath.Join(outside, "keep.go"), "outside")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, "nested")); err != nil {
		t.Skipf("create symlink: %v", err)
	}

	output := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(output.StageDir(), "nested", "keep.go"), "generated")
	if err := output.Commit(); err == nil || !strings.Contains(err.Error(), "contains symlink") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
	assertOutputTestContent(t, filepath.Join(outside, "keep.go"), "outside")
}

func TestManagedOutputRollsBackPartialCommit(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	initial := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(initial.StageDir(), "a.go"), "old a")
	writeGeneratedOutputTestFile(t, filepath.Join(initial.StageDir(), "b.go"), "old b")
	if err := initial.Commit(); err != nil {
		t.Fatal(err)
	}

	output := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(output.StageDir(), "a.go"), "new a")
	writeGeneratedOutputTestFile(t, filepath.Join(output.StageDir(), "b.go"), "new b")
	writes := 0
	output.writeFile = func(path string, content []byte, mode fs.FileMode) error {
		writes++
		if writes == 2 {
			return errors.New("injected write failure")
		}
		return atomicWriteFile(path, content, mode)
	}
	if err := output.Commit(); err == nil || !strings.Contains(err.Error(), "injected write failure") {
		t.Fatalf("expected injected commit failure, got %v", err)
	}

	assertOutputTestContent(t, filepath.Join(target, "a.go"), "old a")
	assertOutputTestContent(t, filepath.Join(target, "b.go"), "old b")
}

func TestManagedOutputRollsBackNewTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	output := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(output.StageDir(), "a.go"), "new a")
	writeGeneratedOutputTestFile(t, filepath.Join(output.StageDir(), "b.go"), "new b")
	writes := 0
	output.writeFile = func(path string, content []byte, mode fs.FileMode) error {
		writes++
		if writes == 2 {
			return errors.New("injected write failure")
		}
		return atomicWriteFile(path, content, mode)
	}
	if err := output.Commit(); err == nil {
		t.Fatal("expected commit failure")
	}
	assertOutputTestMissing(t, target)
}

func TestCommitManagedOutputsRollsBackEarlierTargets(t *testing.T) {
	root := t.TempDir()
	firstTarget := filepath.Join(root, "regular")
	secondTarget := filepath.Join(root, "public")
	for _, target := range []string{firstTarget, secondTarget} {
		initial := newOutputTestTransaction(t, target)
		writeGeneratedOutputTestFile(t, filepath.Join(initial.StageDir(), "generated.go"), "old")
		if err := initial.Commit(); err != nil {
			t.Fatal(err)
		}
	}

	first := newOutputTestTransaction(t, firstTarget)
	second := newOutputTestTransaction(t, secondTarget)
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "generated.go"), "new regular")
	writeGeneratedOutputTestFile(t, filepath.Join(second.StageDir(), "generated.go"), "new public")
	second.writeFile = func(string, []byte, fs.FileMode) error {
		return errors.New("injected second-output failure")
	}
	if err := CommitManagedOutputs([]*ManagedOutput{first, second}); err == nil || !strings.Contains(err.Error(), "injected second-output failure") {
		t.Fatalf("expected second output failure, got %v", err)
	}

	assertOutputTestContent(t, filepath.Join(firstTarget, "generated.go"), "old")
	assertOutputTestContent(t, filepath.Join(secondTarget, "generated.go"), "old")
}

func newOutputTestTransaction(t *testing.T, target string) *ManagedOutput {
	t.Helper()
	output, err := NewManagedOutput(target)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(output.Abort)
	return output
}

func writeGeneratedOutputTestFile(t *testing.T, path, content string) {
	t.Helper()
	marked, err := common.MarkGeneratedFile(filepath.Base(path), content)
	if err != nil {
		t.Fatal(err)
	}
	writeOutputTestFile(t, path, marked)
}

func writeOutputTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertOutputTestContent(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), expected) {
		t.Fatalf("expected %s to contain %q, got %q", path, expected, content)
	}
}

func assertOutputTestMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be missing, err=%v", path, err)
	}
}

func TestIdenticalOutputPreservesFilesAndOnlyBacksUpChanges(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	first := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "keep.go"), "same")
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "change.go"), "old")
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(target, "keep.go")
	before, err := os.Stat(keep)
	if err != nil {
		t.Fatal(err)
	}
	second := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(second.StageDir(), "keep.go"), "same")
	writeGeneratedOutputTestFile(t, filepath.Join(second.StageDir(), "change.go"), "new")
	commit, err := second.prepareCommit()
	if err != nil {
		t.Fatal(err)
	}
	if len(commit.files) != 1 || commit.files[0] != "change.go" || len(commit.snapshots) != 1 {
		t.Fatalf("unexpected changes: %+v", commit)
	}
	second.writeFile = func(string, []byte, fs.FileMode) error { return errors.New("injected write failure") }
	if err := commit.apply(); err == nil {
		t.Fatal("expected failure")
	}
	if err := commit.rollback(); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(keep)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("identical file was replaced")
	}
	assertOutputTestContent(t, filepath.Join(target, "change.go"), "old")
}

func TestIdenticalOutputCommitDoesNotWriteOrRemoveCurrentFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	first := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(first.StageDir(), "keep.go"), "same")
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(target, "keep.go")
	timestamp := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	second := newOutputTestTransaction(t, target)
	writeGeneratedOutputTestFile(t, filepath.Join(second.StageDir(), "keep.go"), "same")
	second.writeFile = func(string, []byte, fs.FileMode) error { t.Error("identical output was written"); return nil }
	if err := second.Commit(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !timestamp.Equal(info.ModTime()) {
		t.Fatal("mtime changed")
	}
}
