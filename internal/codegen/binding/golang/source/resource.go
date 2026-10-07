package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

const resourceGoFilename = "resource.go"

var resourceGoTemplate = joinTemplates(
	"imports.go.tpl",
	"resource.go.tpl",
	"service/info.go.tpl",
	"service/arguments.go.tpl",
	"service/server.go.tpl",
	"service/er_server.go.tpl",
	"service/client.go.tpl",
	"service/er_client.go.tpl",
)

type ResourceGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Resources     []*Resource
	Services      []*Service
}

type Resource struct {
	Name    string
	Actions []*ResourceAction
}

type ResourceAction struct {
	PermissionName string
	PermissionCode string
	CommentLines   []string
}

func (g *_Gen) genResourceGo() {
	payload := g.buildResourceGoPayload()
	if len(payload.Resources) > 0 || len(payload.Services) > 0 {
		g.renderGo(resourceGoFilename, resourceGoTemplate, payload)
	}
}

func (g *_Gen) buildResourceGoPayload() *ResourceGoPayload {
	payload := &ResourceGoPayload{
		PackageName: g.pkgName,
		Resources:   make([]*Resource, 0, len(g.view.Resources)),
		Services:    make([]*Service, 0, len(g.view.Resources)),
	}
	for _, resource := range g.view.Resources {
		payload.Resources = append(payload.Resources, castResource(resource))
		if resource.CheckService != nil {
			payload.Services = append(payload.Services, g.castService(resource.CheckService, false, true))
		}
	}
	imports := []*Import{}
	if len(payload.Services) > 0 {
		imports = buildServiceImports(payload.Services)
	}
	payload.StdImports, payload.ModuleImports = splitImports(imports)
	return payload
}

func castResource(resource *schema.Resource) *Resource {
	casted := &Resource{
		Name:    nameutil.ToCamel(resource.Name),
		Actions: make([]*ResourceAction, 0, len(resource.Actions)),
	}
	for _, action := range resource.Actions {
		casted.Actions = append(casted.Actions, &ResourceAction{
			PermissionName: fmt.Sprintf("%s%sPermission", casted.Name, nameutil.ToCamel(action.Name)),
			PermissionCode: action.PermissionCode,
			CommentLines: deprecatedGoDocLines(
				goDocLines(fmt.Sprintf("%s%sPermission", casted.Name, nameutil.ToCamel(action.Name)), action.Description),
				fmt.Sprintf("%s%sPermission", casted.Name, nameutil.ToCamel(action.Name)),
				action.DeprecatedReason,
			),
		})
	}
	return casted
}
