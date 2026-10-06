package typescript

import (
	"sort"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
)

func buildModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()
	for _, enum := range spec.Enums {
		if enum.UnspecifiedItem == nil {
			enum.UnspecifiedItem = &model.EnumItem{Name: "UNSPECIFIED"}
		}
	}
	sort.Slice(spec.Enums, func(i, j int) bool { return spec.Enums[i].Name < spec.Enums[j].Name })
	sort.Slice(spec.Data, func(i, j int) bool { return spec.Data[i].Name < spec.Data[j].Name })
	sort.Slice(spec.Services, func(i, j int) bool { return spec.Services[i].Name < spec.Services[j].Name })

	return model.NewDomainFromSpec(spec)
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
	content, err := common.RenderTemplate(template, payload)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
