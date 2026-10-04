package golang

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/codegentest"
	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
)

func TestNewGenDerivesModuleAndPackageName(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user.profile"))

	gen, err := newGen(_GenOption{
		VineVersion: "v0.27.0",
		Mode:        view.ModeFull,
		Domain:      pkg,
		Out:         filepath.Join(t.TempDir(), "skeled"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if gen.modName != "" {
		t.Fatalf("unexpected module name: %s", gen.modName)
	}
	if gen.pkgName != "skeled" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenKeepsDomainDerivedPackageNameForModuleOutput(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user.profile"))

	gen, err := newGen(_GenOption{
		ModulePrefix: "github.com/acme/skel",
		VineVersion:  "v0.27.0",
		Mode:         view.ModeFull,
		Domain:       pkg,
		Out:          filepath.Join(t.TempDir(), "skeled"),
		AsModule:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if gen.modName != "github.com/acme/skel/demo/user/profile" {
		t.Fatalf("unexpected module name: %s", gen.modName)
	}
	if gen.pkgName != "profile" {
		t.Fatalf("unexpected module package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPubModuleAndPackageName(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user.profile"))

	gen, err := newGen(_GenOption{
		ModulePrefix: "github.com/acme/skel",
		VineVersion:  "v0.27.0",
		Mode:         view.ModePub,
		Domain:       pkg,
		Out:          filepath.Join(t.TempDir(), "skeled"),
		AsModule:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if gen.modName != "github.com/acme/skel/demo/user/profilepub" {
		t.Fatalf("unexpected module name: %s", gen.modName)
	}
	if gen.pkgName != "profilepub" {
		t.Fatalf("unexpected module package name: %s", gen.pkgName)
	}
}

func TestNewGenRejectsInvalidLocalPackageNameFromOutputDir(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user.profile"))

	_, err := newGen(_GenOption{
		VineVersion: "v0.27.0",
		Mode:        view.ModeFull,
		Domain:      pkg,
		Out:         filepath.Join(t.TempDir(), "my-skel go"),
	})
	if err == nil || !strings.Contains(err.Error(), `go output directory name "my-skel go" is not a valid package name`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewGenRejectsKeywordLocalPackageNameFromOutputDir(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user.profile"))

	_, err := newGen(_GenOption{
		VineVersion: "v0.27.0",
		Mode:        view.ModeFull,
		Domain:      pkg,
		Out:         filepath.Join(t.TempDir(), "go"),
	})
	if err == nil || !strings.Contains(err.Error(), `go output directory name "go" is not a valid package name`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func buildModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()
	return model.NewDomainFromSpec(spec)
}
