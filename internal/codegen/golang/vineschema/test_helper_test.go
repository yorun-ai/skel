package vineschema

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/golang/view"
	"go.yorun.ai/skel/internal/model"
)

func fillModelHashesForTest(pkg *model.Domain) {
	pkg.SetHash("domain-hash")
	for _, data := range pkg.Data() {
		data.Hash = "data-hash"
	}
	for _, actor := range pkg.Actors() {
		actor.Hash = "actor-hash"
	}
	for _, service := range pkg.Services() {
		service.Hash = "service-hash"
		for _, method := range service.Methods {
			method.Hash = "method-hash"
		}
	}
}

func buildModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()
	for _, data := range spec.Data {
		data.Kind = model.DataKindData
		data.Domain = spec.Name
	}
	for _, config := range spec.Configs {
		config.Kind = model.DataKindConfig
		config.Domain = spec.Name
	}
	for _, event := range spec.Events {
		event.Kind = model.DataKindEvent
		event.Domain = spec.Name
	}
	for _, actor := range spec.Actors {
		if actor.AuthEnabled && actor.AuthService == nil {
			for _, data := range []*model.Data{actor.AuthCredential, actor.AuthInfo} {
				data.Kind = model.DataKindData
				data.Domain = spec.Name
				if data.SkelName == "" {
					data.SkelName = spec.Name + "." + data.Name
				}
			}
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

func mustView(t *testing.T, mode view.Mode, domain *model.Domain) *view.Domain {
	t.Helper()
	result, err := view.New(mode, domain)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustBuildDomainSchema(t *testing.T, gen *_Gen) *_DomainSchema {
	t.Helper()
	result, err := gen.buildDomainSchema()
	if err != nil {
		t.Fatal(err)
	}
	return result
}
