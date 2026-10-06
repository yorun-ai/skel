package golang

import (
	"fmt"
	"runtime/debug"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/optionvalidation"
	"golang.org/x/mod/modfile"
	gomodule "golang.org/x/mod/module"
)

const goModFilename = "go.mod"

const (
	goVersion      = "1.27.0"
	decimalModule  = "github.com/shopspring/decimal"
	decimalVersion = "v1.4.0"
	vineModule     = "go.yorun.ai/vine"
	skelModule     = "go.yorun.ai/skel"

	defaultGoImportVersion = "v0.0.0-00010101000000-000000000000"
)

type _ModuleOption struct {
	Sink              binding.FileSink
	Out               string
	Module            string
	Api               bool
	VineVersion       string
	VrpcVersion       string
	CompilerVersion   string
	Imports           map[string]string
	ExtraDependencies []string
}

// ValidateModulePath checks module identities before writing generated metadata.
func ValidateModulePath(path string, field optionvalidation.Field) error {
	if err := gomodule.CheckPath(path); err != nil {
		return optionvalidation.NewValidationError(field, optionvalidation.RuleInvalid, err.Error())
	}
	return nil
}

func generateModule(option _ModuleOption) error {
	if err := ValidateModulePath(option.Module, optionvalidation.FieldGoModule); err != nil {
		return err
	}
	file := new(modfile.File)
	if err := file.AddModuleStmt(option.Module); err != nil {
		return fmt.Errorf("add Go module statement: %w", err)
	}
	if err := file.AddGoStmt(goVersion); err != nil {
		return fmt.Errorf("add Go version statement: %w", err)
	}

	dependencies, err := moduleDependencies(option)
	if err != nil {
		return err
	}
	for _, dependency := range dependencies {
		if err := file.AddRequire(dependency.Module, dependency.Version); err != nil {
			return fmt.Errorf("add Go requirement %s: %w", dependency.Module, err)
		}
	}

	content, err := file.Format()
	if err != nil {
		return fmt.Errorf("format go.mod: %w", err)
	}
	renderer := binding.NewRendererWithSink(option.Out, option.Sink)
	renderer.Write(goModFilename, string(content))
	return renderer.Err()
}

func moduleDependencies(option _ModuleOption) ([]_GoImportDependency, error) {
	runtimeModule, runtimeVersion := vineModule, option.VineVersion
	if option.Api {
		runtimeModule, runtimeVersion = "go.yorun.ai/vrpc", option.VrpcVersion
	}
	info, _ := debug.ReadBuildInfo()
	extra := append([]string{
		decimalModule + "@" + decimalVersion,
		runtimeModule + "@" + runtimeVersion,
		skelModule + "@" + skelModuleVersion(option.CompilerVersion, info),
	}, option.ExtraDependencies...)
	return goModDependencies(option.Imports, extra)
}

// Released tools pin their own module version. Library callers may omit the
// compiler version; in that case use the linked Skel module. Development builds
// require the caller to provide a local workspace or replacement.
func skelModuleVersion(compilerVersion string, info *debug.BuildInfo) string {
	if compilerVersion != "" {
		return compilerVersion
	}
	if info != nil {
		modules := append([]*debug.Module{&info.Main}, info.Deps...)
		for _, module := range modules {
			if module.Path == skelModule && module.Replace == nil && gomodule.Check(skelModule, module.Version) == nil {
				return module.Version
			}
		}
	}
	return developmentCompilerVersion
}
