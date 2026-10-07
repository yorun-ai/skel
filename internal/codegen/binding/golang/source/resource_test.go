package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestResourceCheckServiceUsesResourceGoPayload(t *testing.T) {
	checkService := &schema.Service{
		Name:     "UserCheckService",
		SkelName: "demo.user.UserCheckService",
		Methods: []*schema.Method{{
			Name:     "checkById",
			SkelName: "checkById",
			Arguments: []*schema.Argument{{
				Name: "id",
				Type: codegentest.IntType(),
			}},
			ArgumentsData: &schema.Data{
				Name: "UserCheckServiceCheckByIdArguments",
				Members: []*schema.DataMember{{
					Name: "id",
					Type: codegentest.IntType(),
				}},
			},
		}},
	}
	gen := &_Gen{
		pkgName: "user",
		view: &view.Domain{
			Resources: []*schema.Resource{{
				Name:         "User",
				CheckService: checkService,
			}},
		},
	}

	servicePayload := gen.buildServiceGoPayload()
	if len(servicePayload.Services) != 0 {
		t.Fatalf("expected service.go payload to ignore resource check services, got %+v", servicePayload.Services)
	}

	resourcePayload := gen.buildResourceGoPayload()
	if len(resourcePayload.Services) != 1 || resourcePayload.Services[0].Name != "UserCheckService" {
		t.Fatalf("unexpected resource.go services: %+v", resourcePayload.Services)
	}
}

func TestResourceGoRegistersCheckServices(t *testing.T) {
	checkService := &schema.Service{
		Name:     "UserCheckService",
		SkelName: "demo.user.UserCheckService",
		Methods: []*schema.Method{{
			Name:     "checkById",
			SkelName: "checkById",
			Arguments: []*schema.Argument{{
				Name: "id",
				Type: codegentest.IntType(),
			}},
			ArgumentsData: &schema.Data{
				Name: "UserCheckServiceCheckByIdArguments",
				Members: []*schema.DataMember{{
					Name:      "id",
					Sensitive: true,
					Type:      codegentest.IntType(),
				}},
			},
		}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "user",
		Resources: []*schema.Resource{{
			Name:         "User",
			CheckService: checkService,
		}},
	})
	outputDir := filepath.Join(t.TempDir(), "skeled")
	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         outputDir,
	})

	gen.genResourceGo()

	content, err := os.ReadFile(filepath.Join(outputDir, resourceGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "rpc.Register(_UserCheckServiceSpec)") {
		t.Fatalf("expected resource.go to register check service, got:\n%s", string(content))
	}
	if !strings.Contains(string(content), `json:"id" skel:"index(0),sensitive"`) {
		t.Fatalf("expected resource check argument to preserve sensitive tag, got:\n%s", string(content))
	}
}

func TestResourceGoPayloadIncludesPermissionCodes(t *testing.T) {
	gen := &_Gen{
		pkgName: "user",
		view: &view.Domain{
			Resources: []*schema.Resource{{
				Name: "User",
				Actions: []*schema.ResourceAction{
					{Name: "read", PermissionCode: "app.User:read"},
					{Name: "update", PermissionCode: "app.User:update"},
				},
			}},
		},
	}

	payload := gen.buildResourceGoPayload()
	if len(payload.Resources) != 1 {
		t.Fatalf("unexpected resource count: %d", len(payload.Resources))
	}
	resource := payload.Resources[0]
	if len(resource.Actions) != 2 {
		t.Fatalf("unexpected action count: %d", len(resource.Actions))
	}
	if resource.Actions[0].PermissionName != "UserReadPermission" || resource.Actions[0].PermissionCode != "app.User:read" {
		t.Fatalf("unexpected first permission: %+v", resource.Actions[0])
	}
	if resource.Actions[1].PermissionName != "UserUpdatePermission" || resource.Actions[1].PermissionCode != "app.User:update" {
		t.Fatalf("unexpected second permission: %+v", resource.Actions[1])
	}
}
