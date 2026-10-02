package golang

import (
	"errors"
	"testing"

	"go.yorun.ai/skelc/internal/optionvalidation"
)

func TestImportPathStripsVersion(t *testing.T) {
	path := "go.yorun.ai/app/vine/demo/skeled/userpub@v0.0.0-00010101000000-000000000000"
	got, err := ImportPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "go.yorun.ai/app/vine/demo/skeled/userpub" {
		t.Fatalf("unexpected go import path: %s", got)
	}
}

func TestParseGoImportDependencyRejectsMissingVersion(t *testing.T) {
	if _, err := parseGoImportDependency("go.yorun.ai/app/userpub@"); err == nil {
		t.Fatal("expected missing version error")
	}
}

func TestGoModDependenciesRejectConflicts(t *testing.T) {
	for _, test := range []struct {
		name    string
		imports map[string]string
		extra   []string
	}{
		{"explicit", map[string]string{"a": "example.com/shared@v1.0.0", "b": "example.com/shared@v1.1.0"}, nil},
		{"implicit", map[string]string{"a": "example.com/shared", "b": "example.com/shared@v1.1.0"}, nil},
		{"extra", map[string]string{"a": "example.com/shared@v1.1.0"}, []string{"example.com/shared@v1.0.0"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var message string
			for range 20 {
				_, err := goModDependencies(test.imports, test.extra)
				if err == nil {
					t.Fatal("accepted conflicting versions")
				}
				if message != "" && message != err.Error() {
					t.Fatalf("unstable error: %s != %s", message, err)
				}
				message = err.Error()
			}
		})
	}
}

func TestGoModDependenciesDeduplicatesMatchingVersions(t *testing.T) {
	deps, err := goModDependencies(map[string]string{"a": "example.com/shared@v1.0.0", "b": "example.com/shared@v1.0.0"}, []string{"example.com/shared@v1.0.0"})
	if err != nil || len(deps) != 1 || deps[0].Version != "v1.0.0" {
		t.Fatalf("unexpected dependencies: %+v, %v", deps, err)
	}
}

func TestGoDependencyVersionsRejectMalformedRequirements(t *testing.T) {
	for _, path := range []string{"example.com/pkg@v01.2.3", "example.com/pkg@v1.2", "example.com/pkg@main", "example.com/pkg@v2.0.0", "example.com/pkg/v2@v1.0.0", "bad path@v1.0.0", "local"} {
		t.Run(path, func(t *testing.T) {
			_, err := goModDependencies(map[string]string{"sample": path}, nil)
			var validation *optionvalidation.ValidationError
			if !errors.As(err, &validation) || validation.Field != optionvalidation.FieldGoImport {
				t.Fatalf("expected typed invalid import, got %v", err)
			}
		})
	}
	for _, path := range []string{"example.com/pkg@v1.2.3", "example.com/pkg/v2@v2.0.0", "example.com/pkg@v0.0.0-20260901000000-0123456789ab", "example.com/pkg"} {
		if _, err := goModDependencies(map[string]string{"sample": path}, nil); err != nil {
			t.Fatalf("rejected valid dependency %s: %v", path, err)
		}
	}
}
