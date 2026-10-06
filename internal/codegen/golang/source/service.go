package source

import (
	"strings"

	"go.yorun.ai/skel/internal/model"
)

const (
	serviceGoFilename = "service.go"
)

var serviceImports = []*Import{
	{Path: "reflect"},
	{Path: "go.yorun.ai/vine/core/ex"},
	{Path: "go.yorun.ai/vine/core/rpc"},
}

var serviceGoTemplate = joinTemplates(
	"imports.go.tpl",
	"service/root.go.tpl",
	"service/info.go.tpl",
	"service/arguments.go.tpl",
	"service/server.go.tpl",
	"service/er_server.go.tpl",
	"service/client.go.tpl",
	"service/er_client.go.tpl",
)

type ServiceGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Services      []*Service
}

func (g *_Gen) genServiceGo() {
	payload := g.buildServiceGoPayload()
	if len(payload.Services) > 0 {
		g.renderGo(serviceGoFilename, serviceGoTemplate, payload)
	}
}

func (g *_Gen) buildServiceGoPayload() *ServiceGoPayload {
	payload := &ServiceGoPayload{
		PackageName: g.pkgName,
		Services:    make([]*Service, 0, len(g.view.Services)),
	}
	for _, service := range g.view.Services {
		castedService := g.castService(service, g.serviceClientOnly(service), g.serviceServerOnly(service))
		payload.Services = append(payload.Services, castedService)
	}
	imports := buildServiceImports(payload.Services)
	payload.StdImports, payload.ModuleImports = splitImports(imports)

	return payload
}

type Service struct {
	Name                   string
	SkelName               string
	Hash                   string
	CommentLines           []string
	DeprecatedCommentLines []string
	Ext                    bool
	ClientOnly             bool
	ServerOnly             bool

	SpecName string

	BaseName          string
	ServerName        string
	DefaultServerName string
	ClientName        string
	ClientImplName    string
	ClientCtorName    string

	ERBaseName              string
	ERServerName            string
	WrapperERServerName     string
	WrapperERServerCtorName string
	DefaultERServerName     string
	ERClientName            string
	ERClientImplName        string
	ERClientCtorName        string

	Methods            []*ServiceMethod
	HasMethodArguments bool
}

func (g *_Gen) serviceClientOnly(service *model.Service) bool {
	return (g.isSplitPub() && service.Pub) || (g.isSplitRegular() && service.Ext)
}

func (g *_Gen) serviceServerOnly(service *model.Service) bool {
	return service.Api || (g.isSplitRegular() && service.Pub) || (g.isSplitPub() && service.Ext)
}

func (g *_Gen) castService(p *model.Service, clientOnly bool, serverOnly bool) *Service {
	return g.types.castService(p, clientOnly, serverOnly)
}

func (r _Types) castService(p *model.Service, clientOnly bool, serverOnly bool) *Service {
	names := buildServiceNames(p.Name)
	service := &Service{
		Name:                    names.Name,
		Ext:                     p.Ext,
		SkelName:                p.SkelName,
		Hash:                    p.Hash,
		ClientOnly:              clientOnly,
		ServerOnly:              serverOnly,
		SpecName:                names.SpecName,
		BaseName:                names.BaseName,
		ServerName:              names.ServerName,
		DefaultServerName:       names.DefaultServerName,
		ClientName:              names.ClientName,
		ClientImplName:          names.ClientImplName,
		ClientCtorName:          names.ClientCtorName,
		ERBaseName:              names.ERBaseName,
		ERServerName:            names.ERServerName,
		WrapperERServerName:     names.WrapperERServerName,
		WrapperERServerCtorName: names.WrapperERServerCtorName,
		DefaultERServerName:     names.DefaultERServerName,
		ERClientName:            names.ERClientName,
		ERClientImplName:        names.ERClientImplName,
		ERClientCtorName:        names.ERClientCtorName,
		CommentLines: deprecatedGoDocLines(
			goDocLines(names.ServerName, p.Description),
			names.ServerName,
			p.DeprecatedReason,
		),
		DeprecatedCommentLines: deprecatedGoDocParagraph(p.DeprecatedReason),
	}
	if clientOnly {
		service.CommentLines = deprecatedGoDocLines(
			goDocLines(names.ClientName, p.Description),
			names.ClientName,
			p.DeprecatedReason,
		)
	}
	service.Methods = make([]*ServiceMethod, 0, len(p.Methods))
	for _, method := range p.Methods {
		castedMethod := r.castServiceMethod(p, method)
		service.Methods = append(service.Methods, castedMethod)
	}

	service.HasMethodArguments = false
	for _, m := range service.Methods {
		if m.ArgumentsData != nil {
			service.HasMethodArguments = true
		}
	}

	return service
}

func (r _Types) castActorAuthService(p *model.Service) *Service {
	return r.castService(p, false, true)
}

func buildServiceImports(services []*Service) []*Import {
	imports := newImportSet()
	imports.addMany(serviceImports)
	for _, service := range services {
		for _, method := range service.Methods {
			imports.addMany(collectTypeImports(method.ResultType))
			for _, argument := range method.Arguments {
				imports.addMany(collectTypeImports(argument.Type))
			}
		}
	}
	return imports.sortedValues()
}

func splitImports(imports []*Import) ([]*Import, []*Import) {
	stdImports := make([]*Import, 0, len(imports))
	moduleImports := make([]*Import, 0, len(imports))
	for _, import_ := range imports {
		if strings.Contains(import_.Path, ".") {
			moduleImports = append(moduleImports, import_)
			continue
		}
		stdImports = append(stdImports, import_)
	}
	return stdImports, moduleImports
}
