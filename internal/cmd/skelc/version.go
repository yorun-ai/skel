package skelc

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/Masterminds/semver/v3"
	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skel/internal/cmd/skelc/output"
	"go.yorun.ai/skel/internal/codegen/golang"
)

const (
	commandVersion = "version"

	cliName    = "Skelc CLI"
	devVersion = "v0.0.0-dev"
)

var (
	readBuildInfo   = debug.ReadBuildInfo
	ldModuleVersion string
)

func newVersionCommand() *ucli.Command {
	return &ucli.Command{
		Name: commandVersion,
		Action: func(_ context.Context, cmd *ucli.Command) error {
			if cmd.Args().Len() != 0 {
				return commandFailure(output.ErrorCodeInvalidArgument, fmt.Errorf("unexpected args for %s", commandVersion))
			}
			info, err := versionInfo()
			if err != nil {
				return commandFailure(output.ErrorCodeCommandFailed, err)
			}
			if err := writeJSONResult(cmd, info, "version result"); err != nil {
				return commandFailure(output.ErrorCodeCommandFailed, err)
			}
			return nil
		},
	}
}

type _VersionInfo = output.VersionResult

type _DebugBuildInfo struct {
	Version   string `json:"version"`
	Platform  string `json:"platform"`
	GoVersion string `json:"goVersion"`
}

func versionInfo() (_VersionInfo, error) {
	buildInfo, err := debugBuildInfo()
	if err != nil {
		return _VersionInfo{}, err
	}
	return _VersionInfo{
		Name:      cliName,
		Version:   buildInfo.Version,
		Platform:  buildInfo.Platform,
		GoVersion: buildInfo.GoVersion,
		GolangCodeGen: output.VersionGolangCodeGenResult{
			MinimumVineVersion: golang.MinimumVineVersion,
			DefaultVineVersion: golang.DefaultVineVersion,
		},
	}, nil
}

func debugBuildInfo() (_DebugBuildInfo, error) {
	info, ok := readBuildInfo()
	if !ok {
		return _DebugBuildInfo{}, fmt.Errorf("read Go build info failed")
	}
	rawVersion := info.Main.Version
	if ldModuleVersion != "" {
		rawVersion = ldModuleVersion
	}
	version, err := moduleVersion(rawVersion)
	if err != nil {
		return _DebugBuildInfo{}, err
	}
	return _DebugBuildInfo{
		Version:   version,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion: info.GoVersion,
	}, nil
}

func moduleVersion(rawVersion string) (string, error) {
	if rawVersion == "" || rawVersion == "(devel)" {
		return devVersion, nil
	}

	version := strings.TrimSuffix(rawVersion, "+dirty")
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	_, err := semver.NewVersion(version)
	if err != nil {
		return "", fmt.Errorf("parse module version %s failed: %w", version, err)
	}
	return version, nil
}
