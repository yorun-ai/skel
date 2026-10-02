package vineschema

import (
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"strings"
	"text/template"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
)

//go:embed tpl
var templateFS embed.FS

const schemaGoFilename = "schema.go"

var schemaGoTemplate, schemaGoTemplateError = loadTemplates()

type _Gen struct {
	Domain *model.Domain

	view            *view.Domain
	mode            view.Mode
	pkgName         string
	compilerVersion string
	Renderer        *common.Renderer
}

type Option struct {
	Domain          *model.Domain
	View            *view.Domain
	Mode            view.Mode
	PackageName     string
	CompilerVersion string
	Out             string
}

// GenerateValidated renders a domain already checked by common.ValidateDomain.
func GenerateValidated(domain common.ValidatedDomain, option Option) error {
	option.Domain = domain.Model()
	if schemaGoTemplateError != nil {
		return schemaGoTemplateError
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
		Renderer:        common.NewRenderer(option.Out),
	}
}

func (g *_Gen) gen() error {
	payload, err := g.buildSchemaGoPayload()
	if err != nil {
		return err
	}
	content, err := common.RenderTemplateWithFuncs(schemaGoTemplate, payload, g.schemaGoTemplateFuncs())
	if err != nil {
		return fmt.Errorf("render generated %s: %w", schemaGoFilename, err)
	}
	formatted, err := format.Source([]byte(content))
	if err != nil {
		return fmt.Errorf("format generated %s: %w", schemaGoFilename, err)
	}
	g.Renderer.Write(schemaGoFilename, string(formatted))
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
		return "", fmt.Errorf("list Go schema templates: %w", err)
	}
	var templates strings.Builder
	for _, name := range names {
		content, readErr := templateFS.ReadFile(name)
		if readErr != nil {
			return "", fmt.Errorf("read Go schema template %s: %w", name, readErr)
		}
		templates.Write(content)
		templates.WriteByte('\n')
	}
	return templates.String(), nil
}

type SchemaGoPayload struct {
	PackageName string
	Schema      *_DomainSchema
}

func (g *_Gen) buildSchemaGoPayload() (*SchemaGoPayload, error) {
	schema, err := g.buildDomainSchema()
	if err != nil {
		return nil, err
	}
	return &SchemaGoPayload{
		PackageName: g.pkgName,
		Schema:      schema,
	}, nil
}

func (g *_Gen) schemaGoTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"authLiteral":              renderAuthModeLiteral,
		"permissionRequireLiteral": renderPermRequireModeLiteral,
		"quote":                    quote,
		"scalarLiteral":            renderScalarLiteral,
		"viaLiteral":               renderActorViaLiteral,
	}
}
