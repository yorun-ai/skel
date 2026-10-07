package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

const webGoFilename = "web.go"

var webImports = []*Import{
	{Path: "reflect"},
	{Path: "go.yorun.ai/vine/core/web"},
}

var webGoTemplate = loadGoTemplate("web.go.tpl")

type WebGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Webs          []*Web
}

type Web struct {
	Name              string
	SkelName          string
	Hash              string
	MountPath         string
	CommentLines      []string
	SpecName          string
	ServerName        string
	DefaultServerName string
}

func (g *_Gen) genWebGo() {
	payload := g.buildWebGoPayload()
	if len(payload.Webs) == 0 {
		return
	}
	g.renderGo(webGoFilename, webGoTemplate, payload)
}

func (g *_Gen) buildWebGoPayload() *WebGoPayload {
	payload := &WebGoPayload{
		PackageName: g.pkgName,
		Webs:        make([]*Web, 0, len(g.view.Webs)),
	}
	for _, tokenWeb := range g.view.Webs {
		payload.Webs = append(payload.Webs, g.castWeb(tokenWeb))
	}
	imports := newImportSet()
	imports.addMany(webImports)
	payload.StdImports, payload.ModuleImports = splitImports(imports.sortedValues())
	return payload
}

func (g *_Gen) castWeb(p *schema.Web) *Web {
	name := nameutil.ToCamel(p.Name)
	serverName := fmt.Sprintf("%sServer", name)
	return &Web{
		Name:              name,
		SkelName:          p.SkelName,
		Hash:              p.Hash,
		MountPath:         p.MountPath,
		CommentLines:      deprecatedGoDocLines(goDocLines(serverName, p.Description), serverName, p.DeprecatedReason),
		SpecName:          fmt.Sprintf("_%sSpec", name),
		ServerName:        serverName,
		DefaultServerName: fmt.Sprintf("Default%s", serverName),
	}
}
