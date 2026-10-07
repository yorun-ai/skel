package descriptor

import (
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"strings"
	"text/template"

	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/schema"
)

//go:embed tpl
var templateFS embed.FS

const descriptorGoFilename = "descriptor.go"

var descriptorGoTemplate, descriptorGoTemplateError = loadTemplates()

type _Gen struct {
	Domain *schema.Domain

	view            *view.Domain
	mode            view.Mode
	pkgName         string
	compilerVersion string
	Renderer        *binding.Renderer
}

type Option struct {
	Sink            binding.FileSink
	Domain          *schema.Domain
	View            *view.Domain
	Mode            view.Mode
	PackageName     string
	CompilerVersion string
	Out             string
}

// GenerateValidated renders a domain already checked by codegen.ValidateDomain.
func GenerateValidated(domain codegen.Input, option Option) error {
	option.Domain = domain.Schema()
	if descriptorGoTemplateError != nil {
		return descriptorGoTemplateError
	}
	gen := newGen(option)
	if err := gen.gen(); err != nil {
		return err
	}
	return gen.Renderer.Err()
}

func newGen(option Option) *_Gen {
	return &_Gen{
		Domain:          option.Domain,
		view:            option.View,
		mode:            option.Mode,
		pkgName:         option.PackageName,
		compilerVersion: option.CompilerVersion,
		Renderer:        binding.NewRendererWithSink(option.Out, option.Sink),
	}
}

func (g *_Gen) gen() error {
	payload := g.buildDescriptorGoPayload()
	if err := descriptor.ValidateEffectivePolicy(payload.Descriptor); err != nil {
		return fmt.Errorf("generated descriptor policy: %w", err)
	}
	content, err := binding.RenderTemplateWithFuncs(descriptorGoTemplate, payload, g.descriptorGoTemplateFuncs())
	if err != nil {
		return fmt.Errorf("render generated %s: %w", descriptorGoFilename, err)
	}
	formatted, err := format.Source([]byte(content))
	if err != nil {
		return fmt.Errorf("format generated %s: %w", descriptorGoFilename, err)
	}
	g.Renderer.Write(descriptorGoFilename, string(formatted))
	return g.Renderer.Err()
}

func (g *_Gen) isSplitPub() bool {
	return g.mode == view.ModePub
}

func (g *_Gen) isSplitRegular() bool {
	return g.mode == view.ModeRegular
}

func loadTemplates() (string, error) {
	names, err := fs.Glob(templateFS, "tpl/*.go.tpl")
	if err != nil {
		return "", fmt.Errorf("list Go descriptor templates: %w", err)
	}
	var templates strings.Builder
	for _, name := range names {
		content, readErr := templateFS.ReadFile(name)
		if readErr != nil {
			return "", fmt.Errorf("read Go descriptor template %s: %w", name, readErr)
		}
		templates.Write(content)
		templates.WriteByte('\n')
	}
	return templates.String(), nil
}

type DescriptorGoPayload struct {
	PackageName string
	Descriptor  *descriptor.Domain
}

func (g *_Gen) buildDescriptorGoPayload() *DescriptorGoPayload {
	value := g.buildDomainDescriptor()
	return &DescriptorGoPayload{
		PackageName: g.pkgName,
		Descriptor:  value,
	}
}

func (g *_Gen) descriptorGoTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"authLiteral":              renderAuthModeLiteral,
		"configLifecycleLiteral":   renderConfigLifecycleLiteral,
		"permissionRequireLiteral": renderPermissionRequireModeLiteral,
		"quote":                    quote,
		"scalarLiteral":            renderScalarLiteral,
		"viaLiteral":               renderActorViaLiteral,
	}
}
