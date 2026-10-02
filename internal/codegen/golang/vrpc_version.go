package golang

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/mod/module"
)

const defaultVrpcVersion = "v0.12.0"
const minimumVrpcVersion = "v0.12.0"

func resolveVrpcVersion(version string) (string, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return defaultVrpcVersion, nil
	}
	parsed, err := semver.StrictNewVersion(strings.TrimPrefix(version, "v"))
	if err != nil || !strings.HasPrefix(version, "v") {
		return "", fmt.Errorf("go-vrpc-version must be a v-prefixed semantic version")
	}
	if parsed.Compare(semver.MustParse(minimumVrpcVersion)) < 0 {
		return "", fmt.Errorf("go-vrpc-version must be at least %s", minimumVrpcVersion)
	}
	if err := module.Check("go.yorun.ai/vrpc", version); err != nil {
		return "", err
	}
	return version, nil
}
