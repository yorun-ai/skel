package source

import (
	"embed"
	"fmt"
	"go/format"
	"strings"

	"go.yorun.ai/skel/internal/codegen/binding"
)

//go:embed tpl
var templateFS embed.FS

type _Template struct {
	content string
	err     error
}

func loadTemplate(name string) _Template {
	content, err := templateFS.ReadFile("tpl/" + name)
	if err != nil {
		return _Template{err: fmt.Errorf("read Go template %s: %w", name, err)}
	}
	return _Template{content: string(content)}
}

func loadGoTemplate(name string) _Template {
	return joinTemplates("imports.go.tpl", name)
}

func joinTemplates(names ...string) _Template {
	templates := make([]string, 0, len(names))
	for _, name := range names {
		tpl := loadTemplate(name)
		if tpl.err != nil {
			return tpl
		}
		templates = append(templates, tpl.content)
	}
	return _Template{content: strings.Join(templates, "\n")}
}

func (g *_Gen) renderGo(file string, tpl _Template, data any) {
	if g.Renderer.Err() != nil {
		return
	}
	if tpl.err != nil {
		g.Renderer.Fail(tpl.err)
		return
	}
	content, err := binding.RenderTemplate(tpl.content, data)
	if err != nil {
		g.Renderer.Fail(fmt.Errorf("render generated %s: %w", file, err))
		return
	}
	formatted, err := format.Source([]byte(content))
	if err != nil {
		g.Renderer.Fail(fmt.Errorf("format generated %s: %w", file, err))
		return
	}
	g.Renderer.Write(file, string(formatted))
}
