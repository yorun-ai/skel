package golang

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestGenerateIncludesVersionedGoImports(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "golang")
	if err := generateModule(_ModuleOption{
		Out:         outputDir,
		Module:      "go.yorun.ai/app/vine/demo/skeled/booker",
		VineVersion: "v1.2.3",
		Imports: map[string]string{
			"user": "go.yorun.ai/app/vine/demo/skeled/userpub@v0.0.0-00010101000000-000000000000",
		},
	}); err != nil {
		t.Fatal(err)
	}

	content := readGoModForTest(t, outputDir)
	if !strings.Contains(content, "go 1.27.0") {
		t.Fatalf("expected generated Go toolchain requirement:\n%s", content)
	}
	if !strings.Contains(content, "go.yorun.ai/app/vine/demo/skeled/userpub v0.0.0-00010101000000-000000000000") {
		t.Fatalf("expected versioned go import require:\n%s", content)
	}
}

func TestGenerateIncludesDefaultVersionForGoImports(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "golang")
	if err := generateModule(_ModuleOption{
		Out:         outputDir,
		Module:      "go.yorun.ai/app/vine/demo/skeled/booker",
		VineVersion: "v1.2.3",
		Imports: map[string]string{
			"user": "go.yorun.ai/app/vine/demo/skeled/userpub",
		},
	}); err != nil {
		t.Fatal(err)
	}

	content := readGoModForTest(t, outputDir)
	if !strings.Contains(content, "go.yorun.ai/app/vine/demo/skeled/userpub v0.0.0-00010101000000-000000000000") {
		t.Fatalf("expected default go import require:\n%s", content)
	}
}

func readGoModForTest(t *testing.T, outputDir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outputDir, goModFilename))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	return string(content)
}

func TestGenerateRejectsRuntimeDependencyOverride(t *testing.T) {
	out := t.TempDir()
	err := generateModule(_ModuleOption{Out: out, Module: "example.com/generated", VineVersion: DefaultVineVersion, Imports: map[string]string{"runtime": "go.yorun.ai/vine@v0.18.9"}})
	if err == nil || !strings.Contains(err.Error(), "conflicting Go dependency versions") {
		t.Fatalf("runtime override accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "go.mod")); !os.IsNotExist(err) {
		t.Fatalf("wrote output on conflict: %v", err)
	}
}

func TestSkelModuleVersion(t *testing.T) {
	for _, tc := range []struct {
		name     string
		compiler string
		info     *debug.BuildInfo
		want     string
	}{
		{name: "explicit", compiler: "v1.2.3", want: "v1.2.3"},
		{name: "main", info: &debug.BuildInfo{Main: debug.Module{Path: skelModule, Version: "v1.2.3"}}, want: "v1.2.3"},
		{name: "library", info: &debug.BuildInfo{Deps: []*debug.Module{{Path: skelModule, Version: "v1.2.4"}}}, want: "v1.2.4"},
		{name: "replacement", info: &debug.BuildInfo{Deps: []*debug.Module{{Path: skelModule, Version: "v1.2.4", Replace: &debug.Module{Path: "../skel"}}}}, want: developmentCompilerVersion},
		{name: "development", info: &debug.BuildInfo{Main: debug.Module{Path: skelModule, Version: "(devel)"}}, want: developmentCompilerVersion},
		{name: "unavailable", want: developmentCompilerVersion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := skelModuleVersion(tc.compiler, tc.info); got != tc.want {
				t.Fatalf("version = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGeneratePinsSkelTypesDependency(t *testing.T) {
	for _, api := range []bool{false, true} {
		out := t.TempDir()
		if err := generateModule(_ModuleOption{
			Out: out, Module: "example.com/generated", Api: api,
			CompilerVersion: "v1.2.3", VineVersion: DefaultVineVersion, VrpcVersion: defaultVrpcVersion,
		}); err != nil {
			t.Fatal(err)
		}
		if content := readGoModForTest(t, out); !strings.Contains(content, skelModule+" v1.2.3") {
			t.Fatalf("missing Skel types dependency:\n%s", content)
		}
	}
}
