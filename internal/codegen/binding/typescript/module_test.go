package typescript

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/optionvalidation"
)

func TestBuildPackageJSONPayload(t *testing.T) {
	payload, err := buildPackageJSONPayload(_ModuleOption{PackageName: "@yorun-ai/skeled-example"})
	if err != nil {
		t.Fatal(err)
	}
	if payload.Name != "@yorun-ai/skeled-example" {
		t.Fatalf("unexpected package name: %s", payload.Name)
	}
	if got := joinPackageJSONDependencies(sortedPackageJSONDependencies(payload.PeerDependencies)); got != "@yorun-ai/vrpc@*" {
		t.Fatalf("unexpected peer dependencies: %#v", payload.PeerDependencies)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, `"@yorun-ai/vrpc": "*"`) {
		t.Fatalf("expected rendered package.json to include @yorun-ai/vrpc, got:\n%s", output)
	}
	if !strings.Contains(output, `"peerDependencies": {`) || strings.Contains(output, `"dependencies": {`) {
		t.Fatalf("expected rendered package.json to put deps in peerDependencies, got:\n%s", output)
	}
}

func TestBuildPackageJSONPayloadIncludesConfiguredAndResolvedImports(t *testing.T) {
	payload, err := buildPackageJSONPayload(_ModuleOption{
		PackageName: "@yorun-ai/skeled-example",
		Imports: map[string]string{
			"demo.user": "@vine-demo/skeled-user@workspace:*",
			"app":       "@vine-demo/skeled-app",
		},
		ResolvedImports: map[string]string{
			"demo.user": "@vine-demo/skeled-user",
			"inventory": "@vine-demo/skeled-inventory",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "@vine-demo/skeled-app@*,@vine-demo/skeled-inventory@*,@vine-demo/skeled-user@workspace:*,@yorun-ai/vrpc@*"
	if got := joinPackageJSONDependencies(sortedPackageJSONDependencies(payload.PeerDependencies)); got != want {
		t.Fatalf("unexpected peer dependencies: got=%s want=%s", got, want)
	}
}

func TestImportPathStripsVersion(t *testing.T) {
	got, err := importPath("@vine-demo/skeled-user@workspace:*")
	if err != nil {
		t.Fatal(err)
	}
	if got != "@vine-demo/skeled-user" {
		t.Fatalf("unexpected import path: %s", got)
	}
}

func TestImportPathRejectsMissingVersion(t *testing.T) {
	if _, err := importPath("@vine-demo/skeled-user@"); err == nil {
		t.Fatal("expected missing version error")
	}
}

func TestPackageJSONUsesPureTypeScriptEntryAndEscapesConstraints(t *testing.T) {
	out := t.TempDir()
	constraint := "file:../some\"directory\\path"
	if err := generateModule(_ModuleOption{Out: out, PackageName: "@yorun-ai/skeled-example", Imports: map[string]string{"user": "@vine-demo/skeled-user@" + constraint}}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var payload _PackageJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	if payload.Name != "@yorun-ai/skeled-example" || !payload.Private || payload.Type != "module" || payload.Exports["."].Types != "./index.ts" || payload.Exports["."].Default != "./index.ts" || payload.PeerDependencies["@vine-demo/skeled-user"] != constraint {
		t.Fatalf("unexpected package metadata: %+v", payload)
	}
}

func joinPackageJSONDependencies(dependencies []_PackageJSONDependency) string {
	values := make([]string, 0, len(dependencies))
	for _, dependency := range dependencies {
		values = append(values, dependency.Package+"@"+dependency.Version)
	}
	return strings.Join(values, ",")
}

func TestDependencyConstraintsRejectExplicitConflicts(t *testing.T) {
	var previous string
	for range 40 {
		_, err := packageJSONDependencies(_ModuleOption{Imports: map[string]string{
			"first": "@example/shared@1.0.0", "second": "@example/shared@2.0.0",
		}})
		var validation *optionvalidation.ValidationError
		if !errors.As(err, &validation) || validation.Field != optionvalidation.FieldTypeScriptImport {
			t.Fatalf("expected typed dependency conflict, got %v", err)
		}
		if previous != "" && err.Error() != previous {
			t.Fatalf("non-deterministic conflict: %s / %s", previous, err)
		}
		previous = err.Error()
	}
}

func TestImplicitDependenciesPreserveExplicitConstraints(t *testing.T) {
	for _, version := range []string{"1.0.0", "^1.0.0", "workspace:*", "*"} {
		t.Run(version, func(t *testing.T) {
			dependencies, err := packageJSONDependencies(_ModuleOption{
				Imports:         map[string]string{"explicit": "@example/shared@" + version, "implicit": "@example/shared", "same": "@example/shared@" + version},
				ResolvedImports: map[string]string{"derived": "@example/shared"},
			})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, dependency := range dependencies {
				if dependency.Package == "@example/shared" {
					found = true
					if dependency.Version != version {
						t.Fatalf("lost explicit constraint: %+v", dependency)
					}
				}
			}
			if !found {
				t.Fatal("missing shared dependency")
			}
		})
	}
}
