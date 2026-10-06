package typescript

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/codegen/binding"
)

func (g *_Gen) renderTs(file string, tpl string, data any) {
	if g.renderer.Err() != nil {
		return
	}
	content, err := binding.RenderTemplate(tpl, data)
	if err != nil {
		g.renderer.Fail(fmt.Errorf("render generated %s: %w", file, err))
		return
	}
	g.renderer.Write(file, trimTrailingWhitespace(content))
}

func trimTrailingWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}
