package binding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go.yorun.ai/skel/internal/codegen/output"
)

type Renderer struct {
	outputDir string
	sink      FileSink
	err       error
}

func NewRenderer(outputDir string) *Renderer {
	return new(Renderer{outputDir: outputDir})
}

func NewRendererWithSink(outputDir string, sink FileSink) *Renderer {
	return new(Renderer{outputDir: outputDir, sink: sink})
}

func (r *Renderer) Write(file string, content string) {
	if r.err != nil {
		return
	}
	content = normalizeTrailingNewline(content)
	content, err := output.MarkGeneratedFile(file, content)
	if err != nil {
		r.err = err
		return
	}
	content = normalizeTrailingNewline(content)
	if r.sink != nil {
		r.err = r.sink(file, content)
		return
	}
	fullPath := filepath.Join(r.outputDir, file)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		r.err = fmt.Errorf("create output directory for %s: %w", fullPath, err)
		return
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		r.err = fmt.Errorf("write %s: %w", fullPath, err)
	}
}

func (r *Renderer) Err() error { return r.err }

func (r *Renderer) Fail(err error) {
	if r.err == nil {
		r.err = err
	}
}

func RenderTemplate(tplString string, payloadData any) (string, error) {
	return RenderTemplateWithFuncs(tplString, payloadData, nil)
}

func RenderTemplateWithFuncs(tplString string, payloadData any, funcs template.FuncMap) (string, error) {
	tpl := template.New("template")
	if funcs != nil {
		tpl = tpl.Funcs(funcs)
	}
	tpl, err := tpl.Parse(tplString)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var rendered strings.Builder
	if err := tpl.Execute(&rendered, payloadData); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return rendered.String(), nil
}

func normalizeTrailingNewline(content string) string {
	content = strings.TrimLeft(content, "\n")
	return strings.TrimRight(content, "\n") + "\n"
}
