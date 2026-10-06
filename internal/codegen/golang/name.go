package golang

import (
	"fmt"
	gotoken "go/token"
	"path/filepath"
	"strings"

	"go.yorun.ai/skel/internal/util/nameutil"
)

func buildModuleName(modulePrefix string, domainParts []string, usePubPackage bool) string {
	parts := append([]string{modulePrefix}, domainParts...)
	if usePubPackage {
		parts[len(parts)-1] = pubPackageName(domainParts)
	}
	return strings.Join(parts, "/")
}

func DefaultPubModuleName(modulePrefix string, moduleName string, domainName string) string {
	if moduleName != "" {
		return moduleName + "pub"
	}
	return buildModuleName(modulePrefix, strings.Split(domainName, "."), true)
}

func packageNameFallback(domainParts []string, usePubPackage bool) string {
	if usePubPackage {
		return pubPackageName(domainParts)
	}
	return domainParts[len(domainParts)-1]
}

func pubPackageName(domainParts []string) string {
	return domainParts[len(domainParts)-1] + "pub"
}

func importPackageName(domainName string, usePubPackage bool) string {
	return packageNameFallback(strings.Split(domainName, "."), usePubPackage)
}

func inferPackageName(outputDir string, fallback string, asModule bool) (string, error) {
	if asModule {
		return fallback, nil
	}
	dirName := filepath.Base(outputDir)
	if dirName == "." || dirName == string(filepath.Separator) || dirName == "" {
		return fallback, nil
	}
	if !nameutil.IsSnakeCase(dirName) || gotoken.Lookup(dirName).IsKeyword() {
		return "", fmt.Errorf("go output directory name %q is not a valid package name", dirName)
	}
	return dirName, nil
}
