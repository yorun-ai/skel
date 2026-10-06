package codegen

import (
	"context"
	"errors"

	"go.yorun.ai/skel/internal/model"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFilesRejectsInvalidSetsBeforePublishing(t *testing.T) {
	cases := map[string][]File{
		"escape":             {{Path: "../escape.go"}},
		"absolute":           {{Path: "/tmp/escape.go"}},
		"unclean":            {{Path: "a/../escape.go"}},
		"windows":            {{Path: `a\escape.go`}},
		"duplicate":          {{Path: "a.go"}, {Path: "a.go"}},
		"file directory":     {{Path: "a.go"}, {Path: "a.go/b.go"}},
		"unknown target":     {{Target: "missing", Path: "a.go"}},
		"unsupported marker": {{Path: "a.py", CommentPrefix: "nonsense"}},
		"unknown format":     {{Path: "a.unknown"}},
	}
	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "output")
			if err := WriteFiles(t.Context(), files, map[string]string{"": out}); err == nil {
				t.Fatal("invalid files accepted")
			}
			if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("invalid files touched output: %v", err)
			}
		})
	}
}

func TestWriteFilesRejectsOverlappingTargets(t *testing.T) {
	root := t.TempDir()
	for _, second := range []string{root, filepath.Join(root, "child")} {
		if err := WriteFiles(t.Context(), nil, map[string]string{"": root, "pub": second}); err == nil {
			t.Fatal("overlapping targets accepted")
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("validation left output: %v %v", entries, err)
	}
}

func TestWriteFilesPreservesBothTargetsOnCommitFailure(t *testing.T) {
	root := t.TempDir()
	first, second := filepath.Join(root, "first"), filepath.Join(root, "second")
	if err := WriteFiles(t.Context(), []File{{Path: "a.go", Content: "package original"}}, map[string]string{"": first}); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(first, "a.go"))
	if err != nil {
		t.Fatal(err)
	}
	// A non-regular destination makes the second target fail commit preparation.
	if err := os.MkdirAll(filepath.Join(second, "b.go"), 0o755); err != nil {
		t.Fatal(err)
	}
	err = WriteFiles(t.Context(), []File{{Path: "a.go", Content: "package changed"}, {Target: "pub", Path: "b.go", Content: "package changed"}}, map[string]string{"": first, "pub": second})
	if err == nil {
		t.Fatal("invalid destination accepted")
	}
	after, _ := os.ReadFile(filepath.Join(first, "a.go"))
	if string(after) != string(original) {
		t.Fatal("first output changed after second failed")
	}
}

func TestGenerateCancellationNeverPublishes(t *testing.T) {
	input, err := Prepare(model.NewDomainFromSpec(model.DomainSpec{Name: "demo"}), Selection{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	out := filepath.Join(t.TempDir(), "out")
	generator := GeneratorFunc(func(context.Context, Input) ([]File, error) {
		cancel()
		return []File{{Path: "a.go", Content: "package demo"}}, nil
	})
	if err := Run(ctx, input, generator, map[string]string{"": out}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled generator wrote output")
	}
	if _, err := Generate(ctx, input, GeneratorFunc(func(context.Context, Input) ([]File, error) {
		t.Fatal("cancelled call reached generator")
		return nil, nil
	})); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
