package golang

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
)

// Option configures Go generation. NormalizeOption resolves paths and validates settings.
type Option struct {
	ApiFilter codegen.ApiFilter
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

type _GenOption struct {
	Input     codegen.Input
	Sink      binding.FileSink
	ApiFilter codegen.ApiFilter
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
}
