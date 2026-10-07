package source

const docGoFilename = "doc.go"

var docGoTemplate = loadTemplate("doc.go.tpl")

type DocGoPayload struct {
	PackageName  string
	CommentLines []string
}

func (g *_Gen) genDocGo() {
	payload := g.buildDocGoPayload()
	g.renderGo(docGoFilename, docGoTemplate, payload)
}

func (g *_Gen) buildDocGoPayload() *DocGoPayload {
	commentLines := goDocLines("Package "+g.pkgName, g.Domain.Description())
	if len(commentLines) == 0 {
		commentLines = []string{"Package " + g.pkgName}
	}
	return &DocGoPayload{
		PackageName:  g.pkgName,
		CommentLines: commentLines,
	}
}
