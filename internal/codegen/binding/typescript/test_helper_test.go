package typescript

import (
	"sort"
	"testing"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
)

func buildModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()
	for _, data := range spec.Data {
		data.Kind = model.DataKindData
		data.Domain = spec.Name
	}
	for _, enum := range spec.Enums {
		if enum.UnspecifiedItem == nil {
			enum.UnspecifiedItem = &model.EnumItem{Name: "UNSPECIFIED"}
		}
	}
	sort.Slice(spec.Enums, func(i, j int) bool { return spec.Enums[i].Name < spec.Enums[j].Name })
	sort.Slice(spec.Data, func(i, j int) bool { return spec.Data[i].Name < spec.Data[j].Name })
	sort.Slice(spec.Services, func(i, j int) bool { return spec.Services[i].Name < spec.Services[j].Name })

	domain := model.NewDomainFromSpec(spec)
	roots := (codegen.Declarations{Data: domain.Data(), Services: domain.Services()}).TypeRoots(false)
	codegen.VisitTypeGraphs(roots, func(kind *model.Type) {
		if kind.Data != nil && kind.Data.Kind == "" {
			kind.Data.Kind = model.DataKindData
		}
	})
	return domain
}

func externalDataTypeForTest(data *model.Data, domainName string, alias string, explicitAlias bool) *model.Type {
	type_ := codegentest.DataType(data)
	type_.ExternalDomain = domainName
	type_.ExternalAlias = alias
	type_.ExternalAliasExplicit = explicitAlias
	return type_
}

func renderTemplate(t *testing.T, template string, payload any) string {
	t.Helper()
	content, err := binding.RenderTemplate(template, payload)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func newTestGen(domain *model.Domain, outputDir string, options ...Option) *_Gen {
	option := Option{}
	if len(options) > 0 {
		option = options[0]
	}
	input, err := codegen.Prepare(domain, codegen.Selection{Surface: codegen.SurfaceAPI, API: option.ApiFilter})
	if err != nil {
		return &_Gen{err: err}
	}
	return newGen(input, outputDir, option, nil)
}
