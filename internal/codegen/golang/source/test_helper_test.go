package source

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/golang/view"
	"go.yorun.ai/skel/internal/model"
)

func buildModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()
	for _, data := range spec.Data {
		data.Kind = model.DataKindData
		data.Domain = spec.Name
	}
	for _, actor := range spec.Actors {
		if actor.AuthEnabled && actor.AuthService == nil {
			method := &model.Method{
				Name:       "auth",
				SkelName:   "auth",
				Auth:       model.AuthModeNoAuth,
				ResultType: codegentest.DataType(actor.AuthInfo),
				Arguments: []*model.Argument{
					{Name: "credential", Type: codegentest.DataType(actor.AuthCredential)},
				},
			}
			actor.AuthMethod = method
			actor.AuthService = &model.Service{
				Name:     actor.Name + "AuthService",
				SkelName: spec.Name + "." + actor.Name + "AuthService",
				Auth:     model.AuthModeNoAuth,
				Methods:  []*model.Method{method},
			}
		}
	}

	return model.NewDomainFromSpec(spec)
}

func importPaths(imports []*Import) []string {
	paths := make([]string, 0, len(imports))
	for _, import_ := range imports {
		paths = append(paths, import_.Path)
	}
	return paths
}

func mustView(t *testing.T, mode view.Mode, domain *model.Domain) *view.Domain {
	t.Helper()
	result, err := view.New(mode, domain)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
