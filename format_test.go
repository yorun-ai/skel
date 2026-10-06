package skelc_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.yorun.ai/skelc"
)

func TestFormatSourceAndPlanDoNotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	original := []byte("domain demo\npub data Value{value:string}\n")
	writeTestFile(t, path, string(original))
	formatted, err := skelc.FormatSource(original)
	if err != nil || bytes.Equal(original, formatted) {
		t.Fatalf("format failed: %s, %v", formatted, err)
	}
	again, err := skelc.FormatSource(formatted)
	if err != nil || !bytes.Equal(formatted, again) {
		t.Fatalf("format not idempotent: %v", err)
	}
	planned, err := skelc.FormatFiles(skelc.FormatOption{SkelIn: path})
	if err != nil || !planned.Changed || len(planned.Files) != 1 {
		t.Fatalf("plan=%+v, err=%v", planned, err)
	}
	if !bytes.Equal(planned.Files[0].Original, original) || !bytes.Equal(planned.Files[0].Content, formatted) {
		t.Fatal("incorrect replacement")
	}
	disk, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(disk, original) {
		t.Fatal("format plan modified disk")
	}
	planned, err = skelc.FormatFilesContext(t.Context(), skelc.FormatOption{SkelIn: path, Sources: map[string][]byte{path: formatted}})
	if err != nil || planned.Changed || planned.Files == nil {
		t.Fatalf("clean frozen plan: %+v, %v", planned, err)
	}
	if _, err := skelc.FormatSource([]byte("domain demo\npub data Broken {")); err == nil {
		t.Fatal("invalid source accepted")
	}
}

func TestFormatPlanValidatesEveryInput(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		filepath.Join(dir, "domain.skel"): []byte("domain demo\n"),
		filepath.Join(dir, "a.skel"):      []byte("domain demo\npub data Value{value:string}\n"),
		filepath.Join(dir, "b.skel"):      []byte("domain demo\npub data Broken {"),
	}
	result, err := skelc.FormatFiles(skelc.FormatOption{SkelIn: dir, Sources: files})
	if !errors.Is(err, skelc.ErrFormatCompilation) || len(result.Files) != 0 {
		t.Fatalf("partial plan returned: %+v, %v", result, err)
	}
	files[filepath.Join(dir, "b.skel")] = []byte("domain demo\nservice LegacyService { method ping {} }\n")
	if _, err := skelc.FormatFiles(skelc.FormatOption{SkelIn: dir, Sources: files, Strict: true}); !errors.Is(err, skelc.ErrFormatCompilation) {
		t.Fatalf("strict format error lost: %v", err)
	}
}

func TestFormatPlanPreservesDiscoveryWarnings(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		filepath.Join(dir, "domain.skel"):  []byte("domain demo\n"),
		filepath.Join(dir, ".hidden.skel"): []byte("ignored"),
	}
	for _, strict := range []bool{false, true} {
		result, err := skelc.FormatFiles(skelc.FormatOption{SkelIn: dir, Sources: files, Strict: strict})
		if err != nil || len(result.Diagnostics) != 1 {
			t.Fatalf("strict=%v: diagnostics=%+v, err=%v", strict, result.Diagnostics, err)
		}
		if result.Diagnostics[0].Position.File != filepath.Join(dir, ".hidden.skel") {
			t.Fatalf("warning lost source location: %+v", result.Diagnostics)
		}
	}
}

func TestFormatCancellationIsNotCompilationFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	files := map[string][]byte{path: []byte("domain demo\npub data Value{value:string}\n")}
	for _, strict := range []bool{false, true} {
		for _, deadline := range []bool{false, true} {
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			if deadline {
				ctx, cancel = context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
				defer cancel()
			}
			result, err := skelc.FormatFilesContext(ctx, skelc.FormatOption{SkelIn: path, Sources: files, Strict: strict})
			if !errors.Is(err, ctx.Err()) || errors.Is(err, skelc.ErrFormatCompilation) {
				t.Fatalf("strict=%v deadline=%v: unexpected error classification: %v", strict, deadline, err)
			}
			if result.Changed || len(result.Files) != 0 {
				t.Fatalf("canceled format returned a plan: %+v", result)
			}
		}
	}
}
