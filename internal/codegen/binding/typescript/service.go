package typescript

import (
	_ "embed"
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
)

const serviceTsFilename = "service.ts"

//go:embed tpl/service.ts.tpl
var serviceTsTemplate string

type _ServiceTsPayload struct {
	TypeImports         []string
	ExternalTypeImports []*_TypeImport
	Services            []*_Service
}

func (g *_Gen) genServiceTs() {
	payload := g.buildServiceTsPayload()
	g.renderTs(serviceTsFilename, serviceTsTemplate, payload)
}

func (g *_Gen) buildServiceTsPayload() *_ServiceTsPayload {
	clientServices := g.apiView.Services
	typeImports := g.types.buildServiceTypeImports(clientServices)
	externalTypeImports := g.types.buildServiceExternalTypeImports(clientServices)
	payload := &_ServiceTsPayload{
		Services:            make([]*_Service, 0, len(clientServices)),
		ExternalTypeImports: externalTypeImports,
	}
	services := g.types.castServices(clientServices)
	payload.Services = services
	payload.TypeImports = typeImports
	return payload
}

type _Service struct {
	Name         string
	SkelName     string
	SpecName     string
	CommentLines []string
	FactoryName  string
	Methods      []*_ServiceMethod
	WireMethods  []*_WireMethod
}

type _ServiceNames struct {
	Name        string
	FactoryName string
	SpecName    string
}

func buildServiceNames(serviceName string) *_ServiceNames {
	name := nameutil.ToCamel(serviceName)
	return &_ServiceNames{
		Name:        name,
		FactoryName: fmt.Sprintf("create%s", name),
		SpecName:    fmt.Sprintf("%sSpec", name),
	}
}

func (r _Types) castService(p *model.Service) *_Service {
	names := buildServiceNames(p.Name)
	service := &_Service{
		Name:         names.Name,
		SkelName:     p.SkelName,
		SpecName:     names.SpecName,
		CommentLines: deprecatedTsDocLines(tsDocLines(p.Description), p.Deprecated, p.DeprecatedReason),
		FactoryName:  names.FactoryName,
		Methods:      make([]*_ServiceMethod, 0, len(p.Methods)),
	}
	for _, methodToken := range p.Methods {
		castedMethod := r.castServiceMethod(methodToken)
		service.Methods = append(service.Methods, castedMethod)
	}
	return service
}

func (r _Types) castServices(services []*model.Service) []*_Service {
	castedServices := make([]*_Service, 0, len(services))
	for _, serviceToken := range services {
		castedService := r.castService(serviceToken)
		castedServices = append(castedServices, castedService)
	}
	return castedServices
}

type _ServiceMethod struct {
	Name         string
	SkelName     string
	SummaryLines []string
	ParamDocs    []*_MethodParamDoc
	ReturnDoc    *_MethodReturnDoc
	Arguments    []*_MethodArgument
	HasParams    bool
	ResultType   *_Type
	ReturnType   string
	HasWire      bool
}

func (r _Types) castServiceMethod(p *model.Method) *_ServiceMethod {
	resultType := r.castType(p.ResultType)
	method := &_ServiceMethod{
		Name:         nameutil.ToLowerCamel(p.Name),
		SkelName:     p.Name,
		SummaryLines: deprecatedTsDocLines(tsSummaryLines(p.Description, p.Example), p.Deprecated, p.DeprecatedReason),
		ParamDocs:    make([]*_MethodParamDoc, 0, len(p.Arguments)+2),
		Arguments:    make([]*_MethodArgument, 0, len(p.Arguments)),
		HasParams:    len(p.Arguments) > 0,
		ResultType:   resultType,
		ReturnType:   "void",
		HasWire:      methodArgumentsContainBinary(p) || methodResultContainsBinary(p),
	}
	if resultType != nil {
		method.ReturnType = resultType.Plain
	}
	for _, argToken := range p.Arguments {
		castedArg := r.castMethodArgument(argToken)
		method.Arguments = append(method.Arguments, castedArg)
	}
	method.ParamDocs = append(method.ParamDocs, &_MethodParamDoc{
		Name:        "params",
		Description: binding.ChooseString(method.HasParams, "Request parameters", "Must be null"),
	})
	method.ParamDocs = append(method.ParamDocs, &_MethodParamDoc{
		Name:        "options",
		Description: "Call options, optional",
	})
	method.ReturnDoc = tsReturnDoc(resultType, p.OutputDescription, p.OutputExample)
	return method
}

type _MethodArgument struct {
	Name            string
	SkelName        string
	Description     string
	Type            *_Type
	DeprecatedLines []string
}

type _MethodParamDoc struct {
	Name        string
	Description string
}

type _MethodReturnDoc struct {
	TypeName    string
	Description string
}

func (r _Types) castMethodArgument(p *model.Argument) *_MethodArgument {
	argType := r.castType(p.Type)
	return &_MethodArgument{
		Name:            nameutil.ToLowerCamel(p.Name),
		SkelName:        p.Name,
		Description:     tsTagDoc(binding.MergeDescriptionAndExample(p.Description, p.Example)),
		Type:            argType,
		DeprecatedLines: deprecatedTsDocLines(nil, p.Deprecated, p.DeprecatedReason),
	}
}

func (r _Types) buildServiceTypeImports(services []*model.Service) []string {
	imports := make([]string, 0)
	seen := make(map[string]struct{})
	types := serviceTypeRoots(services)
	codegen.VisitTypes(types, func(current *model.Type) {
		binding := r.bindings[current]
		switch current.Kind {
		case model.TypeKindEnum:
			if binding == nil || binding.Path == "" {
				imports = appendUniqueServiceTypeImport(imports, seen, transEnumName(current.Enum))
			}
		case model.TypeKindData:
			if binding == nil || binding.Path == "" {
				imports = appendUniqueServiceTypeImport(imports, seen, transDataName(current.Data))
			}
		}
	})
	return imports
}

func serviceTypeRoots(services []*model.Service) []*model.Type {
	types := make([]*model.Type, 0)
	for _, service := range services {
		for _, method := range service.Methods {
			types = append(types, method.ResultType)
			for _, argument := range method.Arguments {
				types = append(types, argument.Type)
			}
		}
	}
	return types
}

func appendUniqueServiceTypeImport(imports []string, seen map[string]struct{}, name string) []string {
	if _, ok := seen[name]; ok {
		return imports
	}
	seen[name] = struct{}{}
	return append(imports, name)
}

func (r _Types) buildServiceExternalTypeImports(services []*model.Service) []*_TypeImport {
	return r.buildExternalTypeImports(serviceTypeRoots(services))
}

func tsDocLines(description string) []string {
	return binding.SplitDocLines(description)
}

func tsCommentLines(description string, example string) []string {
	docLines := tsDocLines(binding.MergeDescriptionAndExample(description, example))
	if len(docLines) == 0 {
		return nil
	}
	docLines[0] = binding.EnsureSentence(docLines[0])
	return docLines
}

func tsSummaryLines(description string, example string) []string {
	return tsCommentLines(description, example)
}

func tsReturnDoc(resultType *_Type, description string, example string) *_MethodReturnDoc {
	if resultType == nil {
		return nil
	}
	return &_MethodReturnDoc{
		TypeName:    resultType.Plain,
		Description: tsTagDoc(binding.MergeDescriptionAndExample(description, example)),
	}
}

func tsTagDoc(description string) string {
	lines := tsDocLines(description)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, " ")
}

func deprecatedTsDocLines(lines []string, deprecated bool, reason string) []string {
	if !deprecated {
		return lines
	}
	reasonLines := tsDocLines(strings.TrimSpace(reason))
	if len(reasonLines) == 0 {
		return lines
	}
	reasonLines[0] = "@deprecated " + reasonLines[0]
	return append(lines, reasonLines...)
}
