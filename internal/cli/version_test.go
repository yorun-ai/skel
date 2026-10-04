package cli

import (
	"encoding/json"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/command"
)

func TestRunSkelcVersion(t *testing.T) {
	result := Run([]string{"version"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	info := new(command.VersionResult)
	if err := json.Unmarshal([]byte(result.Stdout), info); err != nil {
		t.Fatalf("decode version result: %v\n%s", err, result.Stdout)
	}
	if info.Name != cliName || info.Version == "" || info.Platform == "" {
		t.Fatalf("unexpected version result: %+v", info)
	}
	if result.Stderr != "" {
		t.Fatalf("unexpected stderr: %q", result.Stderr)
	}
}

func TestGoVineVersions(t *testing.T) {
	info, err := versionInfo()
	if err != nil {
		t.Fatal(err)
	}
	if version := info.GolangCodeGen.MinimumVineVersion; version != "v0.27.0" {
		t.Fatalf("unexpected minimum Go Vine version: %q", version)
	}
	if version := info.GolangCodeGen.DefaultVineVersion; version != "v0.27.0" {
		t.Fatalf("unexpected default Go Vine version: %q", version)
	}
}

func TestRunSkelcVersionRejectsFeatures(t *testing.T) {
	for _, flag := range []string{"--features", "--features=false"} {
		t.Run(flag, func(t *testing.T) {
			result := Run([]string{"version", flag})
			var failure command.Error
			if err := json.Unmarshal([]byte(result.Stdout), &failure); err != nil {
				t.Fatal(err)
			}
			if result.ExitCode != command.ExitCodeError || failure.Code != command.ErrorCodeInvalidArgument {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}

func TestRunSkelcVersionHelpOmitsFeatures(t *testing.T) {
	result := Run([]string{"version", "--help"})
	if result.ExitCode != ExitCodeSuccess || result.Stderr != "" || strings.Contains(result.Stdout, "--features") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestModuleVersion(t *testing.T) {
	for _, test := range []struct {
		name     string
		raw      string
		expected string
	}{
		{name: "empty", expected: devVersion},
		{name: "devel", raw: "(devel)", expected: devVersion},
		{name: "without v prefix", raw: "1.2.3", expected: "v1.2.3"},
		{name: "module version", raw: "v2.3.4", expected: "v2.3.4"},
		{name: "dirty", raw: "v1.1.0-alpha3+dirty", expected: "v1.1.0-alpha3"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := moduleVersion(test.raw)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.expected {
				t.Fatalf("unexpected module version: got %q want %q", got, test.expected)
			}
		})
	}
}

func TestDebugBuildInfo(t *testing.T) {
	setReadBuildInfoForTest(t, func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			GoVersion: "go1.26.0",
			Main:      debug.Module{Version: "v1.1.3"},
		}, true
	})

	info, err := debugBuildInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "v1.1.3" {
		t.Fatalf("unexpected version: %q", info.Version)
	}
	if info.Platform != runtime.GOOS+"/"+runtime.GOARCH {
		t.Fatalf("unexpected platform: %q", info.Platform)
	}
	if info.GoVersion != "go1.26.0" {
		t.Fatalf("unexpected go version: %q", info.GoVersion)
	}
}

func TestDebugBuildInfoUsesLinkerVersion(t *testing.T) {
	original := ldModuleVersion
	t.Cleanup(func() {
		ldModuleVersion = original
	})
	ldModuleVersion = "v1.2.3"
	setReadBuildInfoForTest(t, func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			GoVersion: "go1.26.0",
			Main:      debug.Module{Version: "(devel)"},
		}, true
	})

	info, err := debugBuildInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "v1.2.3" {
		t.Fatalf("unexpected version: %q", info.Version)
	}
}

func TestDebugBuildInfoRejectsMissingBuildInfo(t *testing.T) {
	setReadBuildInfoForTest(t, func() (*debug.BuildInfo, bool) {
		return nil, false
	})

	_, err := debugBuildInfo()
	if err == nil {
		t.Fatal("expected error")
	}
}

func setReadBuildInfoForTest(t *testing.T, read func() (*debug.BuildInfo, bool)) {
	t.Helper()
	original := readBuildInfo
	t.Cleanup(func() {
		readBuildInfo = original
	})
	readBuildInfo = read
}
