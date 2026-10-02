package typescript

import _ "embed"

const specTsFilename = "spec.ts"

//go:embed tpl/spec.ts.tpl
var specTsTemplate string

type _SpecTsPayload struct {
	HasWire       bool
	WireFactories []*_WireFactory
	Services      []*_Service
}

func (g *_Gen) genSpecTs() {
	payload := g.buildSpecTsPayload()
	g.renderTs(specTsFilename, specTsTemplate, payload)
}

func (g *_Gen) buildSpecTsPayload() *_SpecTsPayload {
	serviceTokens := g.apiView.Services
	services := g.types.castServices(serviceTokens)
	builder := newWireSchemaBuilder()
	for _, service := range serviceTokens {
		for _, method := range service.Methods {
			builder.collectMethod(method)
		}
	}
	builder.prepareFactoryNames()

	hasWire := false
	for serviceIndex, service := range serviceTokens {
		for _, method := range service.Methods {
			if !methodArgumentsContainBinary(method) && !methodResultContainsBinary(method) {
				continue
			}
			hasWire = true
			services[serviceIndex].WireMethods = append(
				services[serviceIndex].WireMethods,
				builder.renderMethod(method),
			)
		}
	}
	if builder.err != nil && g.err == nil {
		g.err = builder.err
	}

	return &_SpecTsPayload{
		HasWire:       hasWire,
		WireFactories: builder.renderFactories(),
		Services:      services,
	}
}
