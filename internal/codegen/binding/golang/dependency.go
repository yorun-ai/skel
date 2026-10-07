package golang

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"go.yorun.ai/skel/internal/optionvalidation"
	gomodule "golang.org/x/mod/module"
)

type _GoImportDependency struct {
	Module  string
	Version string
}

func goModDependencies(imports map[string]string, extraDependencies []string) (result []_GoImportDependency, err error) {
	defer func() {
		if err != nil {
			err = optionvalidation.NewValidationError(optionvalidation.FieldGoImport, optionvalidation.RuleInvalid, err.Error())
		}
	}()
	dependencies := map[string]string{}
	paths := make([]string, 0, len(imports)+len(extraDependencies))
	for _, path := range imports {
		paths = append(paths, path)
	}
	paths = append(paths, extraDependencies...)
	slices.Sort(paths)
	for _, path := range paths {
		dependency, err := parseGoImportDependency(path)
		if err != nil {
			return nil, err
		}
		dependency.fillDefaultVersion()
		if err := gomodule.Check(dependency.Module, dependency.Version); err != nil {
			return nil, err
		}
		if version, exists := dependencies[dependency.Module]; exists && version != dependency.Version {
			return nil, fmt.Errorf("conflicting Go dependency versions for %s: %s and %s", dependency.Module, version, dependency.Version)
		}
		dependencies[dependency.Module] = dependency.Version
	}
	return sortedGoImportDependencies(dependencies), nil
}

func (d *_GoImportDependency) fillDefaultVersion() {
	if d.Version == "" {
		d.Version = defaultGoImportVersion
	}
}

func parseGoImportDependency(path string) (_GoImportDependency, error) {
	index := strings.LastIndex(path, "@")
	if index < 0 {
		if path == "" {
			return _GoImportDependency{}, fmt.Errorf("invalid Go import %q: missing module", path)
		}
		if err := gomodule.CheckImportPath(path); err != nil {
			return _GoImportDependency{}, err
		}
		return _GoImportDependency{Module: path}, nil
	}
	module := path[:index]
	version := path[index+1:]
	if module == "" {
		return _GoImportDependency{}, fmt.Errorf("invalid Go import %q: missing module", path)
	}
	if version == "" {
		return _GoImportDependency{}, fmt.Errorf("invalid Go import %q: missing version", path)
	}
	if !strings.HasPrefix(version, "v") {
		return _GoImportDependency{}, fmt.Errorf("invalid Go import %q: version must be a v-prefixed semantic version", path)
	}
	if _, err := semver.StrictNewVersion(strings.TrimPrefix(version, "v")); err != nil {
		return _GoImportDependency{}, fmt.Errorf("invalid Go import %q: %w", path, err)
	}
	if err := gomodule.Check(module, version); err != nil {
		return _GoImportDependency{}, err
	}
	return _GoImportDependency{Module: module, Version: version}, nil
}

func ImportPath(path string) (string, error) {
	dependency, err := parseGoImportDependency(path)
	return dependency.Module, err
}

func sortedGoImportDependencies(dependencies map[string]string) []_GoImportDependency {
	ordered := make([]_GoImportDependency, 0, len(dependencies))
	for module, version := range dependencies {
		ordered = append(ordered, _GoImportDependency{Module: module, Version: version})
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Module < ordered[j].Module
	})
	return ordered
}
