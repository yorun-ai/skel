// Package testutil holds shared helpers for tests that drive external
// command-line tools. Only test files import it; it never ships in a build.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// RequireGit skips t when git is not available on PATH.
func RequireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

// InitRepository initializes a repository at directory with a deterministic
// identity and commit signing disabled, so committing does not depend on the
// git configuration of the machine running the tests.
func InitRepository(t *testing.T, directory string) {
	t.Helper()
	RequireGit(t)
	gitCommand(t, directory, "init", "--quiet")
	gitCommand(t, directory, "config", "user.name", "Skel Test")
	gitCommand(t, directory, "config", "user.email", "skel@example.com")
	gitCommand(t, directory, "config", "commit.gpgsign", "false")
}

// Commit stages paths, or every change when paths is empty, and commits them.
func Commit(t *testing.T, directory string, message string, paths ...string) {
	t.Helper()
	if len(paths) == 0 {
		gitCommand(t, directory, "add", "--all")
	} else {
		gitCommand(t, directory, append([]string{"add", "--"}, paths...)...)
	}
	gitCommand(t, directory, "commit", "--quiet", "-m", message)
}

func gitCommand(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", directory}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", args, err, output)
	}
}

// RequireToolchain guards tests that compile a generated module end to end.
// They need a writable build cache, a populated module cache and network access
// to the module proxy, so they are skipped in short mode.
func RequireToolchain(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping generated-module compilation in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
}

// UseLocalSkel makes a temporary generated module test the runtime types from
// this checkout. Other dependencies continue to use their published versions.
func UseLocalSkel(t *testing.T, directory string) {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate Skel test source")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(source)))
	Go(t, directory, "mod", "edit", "-replace=go.yorun.ai/skel="+root)
}

// Go runs the go command in directory with a self-contained module graph and
// returns its combined output.
func Go(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("go", args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", args, err, output)
	}
	return string(output)
}
