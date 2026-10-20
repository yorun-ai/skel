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
