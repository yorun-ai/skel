package source

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func buildSchemaDomainForTest(t *testing.T, spec schema.DomainSpec) *schema.Domain {
	t.Helper()
	for _, data := range spec.Data {
		data.Kind = schema.DataKindData
		data.Domain = spec.Name
	}
	for _, actor := range spec.Actors {
		for _, data := range []*schema.Data{actor.AuthInfo, actor.AuthCredential} {
			if data != nil {
				data.Kind = schema.DataKindData
				data.Domain = spec.Name
			}
		}
		if actor.AuthEnabled && actor.AuthService == nil {
			method := &schema.Method{
				Name:       "auth",
				SkelName:   "auth",
				Auth:       schema.AuthModeNoAuth,
				ResultType: codegentest.DataType(actor.AuthInfo),
				Arguments: []*schema.Argument{
					{Name: "credential", Type: codegentest.DataType(actor.AuthCredential)},
				},
			}
			actor.AuthMethod = method
			actor.AuthService = &schema.Service{
				Name:     actor.Name + "AuthService",
				SkelName: spec.Name + "." + actor.Name + "AuthService",
				Auth:     schema.AuthModeNoAuth,
				Methods:  []*schema.Method{method},
			}
		}
	}

	return schema.NewDomainFromSpec(spec)
}

func importPaths(imports []*Import) []string {
	paths := make([]string, 0, len(imports))
	for _, import_ := range imports {
		paths = append(paths, import_.Path)
	}
	return paths
}

func mustView(t *testing.T, mode view.Mode, domain *schema.Domain) *view.Domain {
	t.Helper()
	result, err := view.New(mode, domain)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
