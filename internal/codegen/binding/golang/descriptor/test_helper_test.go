package descriptor

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func fillSchemaHashesForTest(pkg *schema.Domain) {
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

func buildDescriptorDomainForTest(t *testing.T, spec schema.DomainSpec) *schema.Domain {
	t.Helper()
	for _, data := range spec.Data {
		data.Kind = schema.DataKindData
		data.Domain = spec.Name
	}
	for _, config := range spec.Configs {
		config.Kind = schema.DataKindConfig
		config.Domain = spec.Name
	}
	for _, event := range spec.Events {
		event.Kind = schema.DataKindEvent
		event.Domain = spec.Name
	}
	for _, actor := range spec.Actors {
		if actor.Auth != nil && actor.Auth.Service == nil {
			for _, data := range []*schema.Data{actor.Auth.Credential, actor.Auth.Info} {
				data.Kind = schema.DataKindData
				data.Domain = spec.Name
				if data.SkelName == "" {
					data.SkelName = spec.Name + "." + data.Name
				}
			}
			method := &schema.Method{
				Name:       "auth",
				SkelName:   "auth",
				AuthMode:   schema.AuthModeNoAuth,
				ResultType: codegentest.DataType(actor.Auth.Info),
				Arguments: []*schema.Argument{
					{Name: "credential", Type: codegentest.DataType(actor.Auth.Credential)},
				},
			}
			actor.Auth.Method = method
			actor.Auth.Service = &schema.Service{
				Name:     actor.Name + "AuthService",
				SkelName: spec.Name + "." + actor.Name + "AuthService",
				AuthMode: schema.AuthModeNoAuth,
				Methods:  []*schema.Method{method},
			}
		}
	}
	spec.Hash = "domain-hash"
	return schema.NewDomainFromSpec(spec)
}

func mustView(t *testing.T, mode view.Mode, domain *schema.Domain) *view.Domain {
	t.Helper()
	result, err := view.New(mode, domain)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

const portableDescriptorSource = `domain demo
pub data Node { children: list<Node> }
pub resource Document {
    check byId { input { id: string } }
    action read { check owner { input { ownerId: string } } }
    action write { check owner { input { editorId: string } } }
}
actor ClientActor {
    via client {}
    auth {
        credential { token: string }
        info { id: string }
    }
    permission {}
}
actor GuestActor { via client {} }
actor PermissionActor { via client {} permission {} }
actor TokenActor {
    via client {}
    auth {
        credential { token: string }
        info { @identifier id: string }
    }
}
service DocumentService {
    for ClientActor
    require Document:read
    method get {
        require Document:read:byId(id)
        input { id: string }
        output Node
    }
}`
