package golang

import (
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/codegen/golang/view"
	"go.yorun.ai/skel/internal/model"
)

// Option configures Go generation. NormalizeOption resolves paths and validates settings.
type Option struct {
	ApiFilter common.ApiFilter
	// CompilerVersion identifies the actual skelc version embedded in generated metadata.
	// Required for backend output; v0.0.0-dev identifies development builds.
	CompilerVersion string
	AsModule        bool
	PubOnly         bool
	ApiOnly         bool
	Out             string
	Module          string
	PubOut          string
	PubModule       string
	Imports         map[string]string
	ModulePrefix    string
	// VineVersion defaults to DefaultVineVersion and must satisfy MinimumVineVersion.
	VineVersion string
	VrpcVersion string
}

// ResolvedOption carries generation options with validated compiler and runtime versions.
type ResolvedOption struct{ option Option }

// Options returns a copy of the resolved generation settings.
func (o ResolvedOption) Options() Option { return o.option }

// WithOutputs redirects generated files to managed staging directories.
func (o ResolvedOption) WithOutputs(out, pubOut string) ResolvedOption {
	o.option.Out, o.option.PubOut = out, pubOut
	return o
}

type _GenOption struct {
	ApiFilter common.ApiFilter
	AsModule  bool
	Out       string
	Module    string

	CompilerVersion string
	Imports         map[string]string
	ModulePrefix    string
	VineVersion     string
	VrpcVersion     string

	Mode              view.Mode
	PubImportPath     string
	ExtraDependencies []string

	Domain *model.Domain
}
