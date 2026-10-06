package typescript

import (
	_ "embed"
)

const indexFilename = "index.ts"

//go:embed tpl/index.ts.tpl
var indexTemplate string

func (g *_Gen) genIndex() {
	g.renderTs(indexFilename, indexTemplate, g.hasApiDeclarations())
}
