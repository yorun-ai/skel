package typescript

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/optionvalidation"
)

const (
	packageJSONFilename    = "package.json"
	defaultTSImportVersion = "*"
)

type _ModuleOption struct {
	Sink            binding.FileSink
	Out             string
	PackageName     string
	Imports         map[string]string
	ResolvedImports map[string]string
}

type _PackageJSON struct {
	Name             string                    `json:"name"`
	Private          bool                      `json:"private"`
	Type             string                    `json:"type"`
	Exports          map[string]_PackageExport `json:"exports"`
	PeerDependencies map[string]string         `json:"peerDependencies"`
}

type _PackageExport struct {
	Types   string `json:"types"`
	Default string `json:"default"`
}

type _PackageJSONDependency struct {
	Package string
	Version string
}

func generateModule(option _ModuleOption) error {
	payload, err := buildPackageJSONPayload(option)
	if err != nil {
		return err
	}
	renderer := binding.NewRendererWithSink(option.Out, option.Sink)
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("render %s: %w", packageJSONFilename, err)
	}
	renderer.Write(packageJSONFilename, string(content))
	return renderer.Err()
}

func buildPackageJSONPayload(option _ModuleOption) (*_PackageJSON, error) {
	if err := validatePackageName(option.PackageName); err != nil {
		return nil, optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptModule, optionvalidation.RuleInvalid, err.Error())
	}
	dependencies, err := packageJSONDependencies(option)
	if err != nil {
		return nil, err
	}
	peers := make(map[string]string, len(dependencies))
	for _, dependency := range dependencies {
		peers[dependency.Package] = dependency.Version
	}
	return &_PackageJSON{Name: option.PackageName, Private: true, Type: "module", Exports: map[string]_PackageExport{".": {Types: "./index.ts", Default: "./index.ts"}}, PeerDependencies: peers}, nil
}

func packageJSONDependencies(option _ModuleOption) (result []_PackageJSONDependency, err error) {
	defer func() {
		if err != nil {
			err = optionvalidation.NewValidationError(optionvalidation.FieldTypeScriptImport, optionvalidation.RuleInvalid, err.Error())
		}
	}()
	dependencies := map[string]string{"@yorun-ai/vrpc": ""}
	paths := make([]string, 0, len(option.Imports)+len(option.ResolvedImports))
	for _, path := range option.Imports {
		paths = append(paths, path)
	}
	for domain, path := range option.ResolvedImports {
		if option.Imports[domain] == "" {
			paths = append(paths, path)
		}
	}
	slices.Sort(paths)
	for _, path := range paths {
		dependency, err := parseTSImportDependency(path)
		if err != nil {
			return nil, err
		}
		if err := validatePackageName(dependency.Package); err != nil {
			return nil, err
		}
		previous, exists := dependencies[dependency.Package]
		if exists && previous != "" && dependency.Version != "" && previous != dependency.Version {
			return nil, fmt.Errorf("conflicting TypeScript dependency versions for %s: %s and %s", dependency.Package, previous, dependency.Version)
		}
		if !exists || dependency.Version != "" {
			dependencies[dependency.Package] = dependency.Version
		}
	}
	for pkg, version := range dependencies {
		if version == "" {
			dependencies[pkg] = defaultTSImportVersion
		}
	}
	return sortedPackageJSONDependencies(dependencies), nil
}

func importPath(path string) (string, error) {
	dependency, err := parseTSImportDependency(path)
	return dependency.Package, err
}

func parseTSImportDependency(path string) (_PackageJSONDependency, error) {
	index := strings.LastIndex(path, "@")
	if index <= 0 {
		if path == "" {
			return _PackageJSONDependency{}, fmt.Errorf("invalid TypeScript import %q: missing package", path)
		}
		return _PackageJSONDependency{Package: path}, nil
	}
	pkg := path[:index]
	version := path[index+1:]
	if version == "" {
		return _PackageJSONDependency{}, fmt.Errorf("invalid TypeScript import %q: missing version", path)
	}
	return _PackageJSONDependency{Package: pkg, Version: version}, nil
}

func sortedPackageJSONDependencies(dependencies map[string]string) []_PackageJSONDependency {
	ordered := make([]_PackageJSONDependency, 0, len(dependencies))
	for dependency, version := range dependencies {
		ordered = append(ordered, _PackageJSONDependency{Package: dependency, Version: version})
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Package < ordered[j].Package
	})
	return ordered
}
