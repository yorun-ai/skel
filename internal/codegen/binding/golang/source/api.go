package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
)

const apiSkelImport = "go.yorun.ai/vrpc/skel"

var apiGoTemplate = joinTemplates(
	"imports.go.tpl",
	"api/root.go.tpl",
	"api/info.go.tpl",
	"service/arguments.go.tpl",
	"api/client.go.tpl",
)

type _ApiPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Services      []*_ApiService
}

type _ApiService struct {
	*_ServiceNames
	HasMethodArguments bool
	SkelName           string
	CommentLines       []string
	Methods            []*_ApiMethod
}

type _ApiMethod struct {
	*_ClientMethodNames
	Name               string
	SkelName           string
	CommentLines       []string
	ArgumentsData      *Data
	Arguments          []*MethodArgument
	ResultType         *Type
	ArgumentsBinary    bool
	ResultBinary       bool
	ArgumentsSensitive bool
	ResultSensitive    bool
}

func (g *_Gen) apiImports(imports []*Import) []*Import {
	if g.mode != view.ModeApi {
		return imports
	}
	imports = cloneImports(imports)
	for _, item := range imports {
		if item.Path == skelImport {
			item.Path = apiSkelImport
		}
	}
	return imports
}

func (g *_Gen) genApiGo() {
	payload := &_ApiPayload{PackageName: g.pkgName}
	imports := newImportSet()
	for _, service := range g.view.Services {
		names := buildServiceNames(service.Name)
		item := &_ApiService{
			_ServiceNames: names,
			SkelName:      service.SkelName,
			CommentLines:  deprecatedGoDocLines(goDocLines(names.ClientName, service.Description), names.ClientName, service.DeprecatedReason),
		}
		for _, method := range service.Methods {
			m := &_ApiMethod{
				Name:               nameutil.ToCamel(method.Name),
				SkelName:           method.SkelName,
				ArgumentsSensitive: method.ArgumentsSensitive,
				ResultSensitive:    method.ResultSensitive,
				ResultBinary:       methodResultContainsBinaryType(method),
			}
			for _, arg := range method.Arguments {
				argument := g.types.castMethodArgument(arg)
				m.Arguments = append(m.Arguments, argument)
				imports.addMany(g.apiImports(argument.Type.Imports))
			}
			m.ArgumentsBinary = methodArgumentsContainBinaryType(method)
			if method.ArgumentsData != nil {
				m.ArgumentsData = g.types.castData(method.ArgumentsData)
				m.ArgumentsData.Name = "_" + m.ArgumentsData.Name
				for _, arg := range m.Arguments {
					member, ok := sliceutil.Find(m.ArgumentsData.Members, func(member *DataMember) bool {
						return member.SkelName == arg.SkelName
					})
					if ok {
						arg.MemberName = member.Name
					}
				}
				item.HasMethodArguments = true
			}
			if method.ResultType != nil {
				kind := g.types.castType(method.ResultType)
				if kind == nil {
					g.Renderer.Fail(fmt.Errorf("unsupported API result type for %s.%s", service.Name, method.Name))
					return
				}
				m.ResultType = kind
				imports.addMany(g.apiImports(kind.Imports))
			}
			m._ClientMethodNames = buildClientMethodNames(m.Arguments)
			m.CommentLines = goMethodDocLines(
				m.Name, method.Description, method.Example, m.Arguments, m.ResultType,
				method.OutputDescription, method.OutputExample, method.DeprecatedReason,
			)
			item.Methods = append(item.Methods, m)
		}
		payload.Services = append(payload.Services, item)
	}
	g.genEnumGo()
	g.genDataGo()
	if len(payload.Services) == 0 {
		return
	}
	imports.add(&Import{Path: "context"})
	imports.add(&Import{Path: "go.yorun.ai/vrpc"})
	payload.StdImports, payload.ModuleImports = splitImports(imports.sortedValues())
	g.renderGo("service.go", apiGoTemplate, payload)
}
