package skeleton

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"go.yorun.ai/skel/internal/formatter"
)

//go:embed tpl/*.skel.tpl
var skelTemplateFS embed.FS

func newSkelTemplate() (*template.Template, error) {
	tpl, err := template.New("skel").Funcs(skelTemplateFuncs()).ParseFS(skelTemplateFS, "tpl/*.skel.tpl")
	if err != nil {
		return nil, fmt.Errorf("parse Skel templates: %w", err)
	}
	return tpl, nil
}

func (g *_Gen) renderSkel(templateName string, payload *_SkelPayload) (string, error) {
	var rendered bytes.Buffer
	if err := g.template.ExecuteTemplate(&rendered, templateName, payload); err != nil {
		return "", fmt.Errorf("execute Skel template %s: %w", templateName, err)
	}
	formatted, err := formatter.Source(rendered.Bytes())
	if err != nil {
		return "", fmt.Errorf("format Skel template %s: %w", templateName, err)
	}
	return string(formatted), nil
}

func skelTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"description":   descriptionView,
		"deprecated":    deprecatedView,
		"example":       exampleView,
		"sensitive":     sensitiveView,
		"importAlias":   importAlias,
		"emptyMethod":   emptyMethod,
		"authMarker":    authMarker,
		"methodAuth":    methodAuthMarker,
		"typeParams":    typeParameterNames,
		"typeRef":       typeView,
		"resourceCheck": resourceCheckView,
		"configSuffix":  configLifecycle,
		"spaces":        strings.Repeat,
	}
}
