package golang

import (
	"fmt"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/optionvalidation"
	"golang.org/x/mod/modfile"
	gomodule "golang.org/x/mod/module"
)

const goModFilename = "go.mod"

const (
	goVersion      = "1.27.0"
	decimalModule  = "github.com/shopspring/decimal"
	decimalVersion = "v1.4.0"
	vineModule     = "go.yorun.ai/vine"

	defaultGoImportVersion = "v0.0.0-00010101000000-000000000000"
)

type _ModuleOption struct {
	Out               string
	Module            string
	Api               bool
	VineVersion       string
	VrpcVersion       string
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
	renderer := common.NewRenderer(option.Out)
	renderer.Write(goModFilename, string(content))
	return renderer.Err()
}

func moduleDependencies(option _ModuleOption) ([]_GoImportDependency, error) {
	runtimeModule, runtimeVersion := vineModule, option.VineVersion
	if option.Api {
		runtimeModule, runtimeVersion = "go.yorun.ai/vrpc", option.VrpcVersion
	}
	extra := append([]string{decimalModule + "@" + decimalVersion, runtimeModule + "@" + runtimeVersion}, option.ExtraDependencies...)
	return goModDependencies(option.Imports, extra)
}
