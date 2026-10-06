package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
)

func TestFacadeGoRendersActorAuthService(t *testing.T) {
	credential := &model.Data{
		Name: "PublicActorCredential",
		Members: []*model.DataMember{
			{Name: "token", Type: codegentest.StringType()},
		},
	}
	info := &model.Data{
		Name: "PublicActorInfo",
		Members: []*model.DataMember{
			{Name: "userId", Type: codegentest.StringType()},
		},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.auth",
		Actors: []*model.Actor{
			{
				Pub:            true,
				Name:           "PublicActor",
				Vias:           []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
				AuthEnabled:    true,
				AuthCredential: credential,
				AuthInfo:       info,
			},
		},
	})
	outputDir := filepath.Join(t.TempDir(), "auth")
	gen := newGen(Option{
		Domain:        pkg,
		View:          mustView(t, view.ModeRegular, pkg),
		Mode:          view.ModeRegular,
		PackageName:   "auth",
		Out:           outputDir,
		PubImportPath: "github.com/acme/skel/demo/authpub",
	})

	gen.genFacadeGo()

	content := readFacadeGoForTest(t, outputDir)
	for _, expected := range []string{
		`import "github.com/acme/skel/demo/authpub"`,
		"type PublicActor = authpub.PublicActor",
		"type PublicActorCredential = authpub.PublicActorCredential",
		"type PublicActorInfo = authpub.PublicActorInfo",
		"type PublicActorAuthServiceServer = authpub.PublicActorAuthServiceServer",
		"type DefaultPublicActorAuthServiceServer = authpub.DefaultPublicActorAuthServiceServer",
		"type PublicActorAuthServiceServerER = authpub.PublicActorAuthServiceServerER",
		"type DefaultPublicActorAuthServiceServerER = authpub.DefaultPublicActorAuthServiceServerER",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected pub.go to contain %q, got:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "PublicActorAuthServiceClient") {
		t.Fatalf("did not expect auth service client facade, got:\n%s", content)
	}
}

func TestFacadeGoRendersResourcePermissions(t *testing.T) {
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.app",
		Resources: []*model.Resource{
			{
				Pub:  true,
				Name: "User",
				Actions: []*model.ResourceAction{
					{Name: "read"},
					{Name: "update"},
					{Name: "manage"},
				},
			},
		},
	})
	outputDir := filepath.Join(t.TempDir(), "app")
	gen := newGen(Option{
		Domain:        pkg,
		View:          mustView(t, view.ModeRegular, pkg),
		Mode:          view.ModeRegular,
		PackageName:   "app",
		Out:           outputDir,
		PubImportPath: "github.com/acme/skel/demo/apppub",
	})

	gen.genFacadeGo()

	content := readFacadeGoForTest(t, outputDir)
	normalizedContent := strings.Join(strings.Fields(content), " ")
	for _, expected := range []string{
		"UserReadPermission = apppub.UserReadPermission",
		"UserUpdatePermission = apppub.UserUpdatePermission",
		"UserManagePermission = apppub.UserManagePermission",
	} {
		if !strings.Contains(normalizedContent, expected) {
			t.Fatalf("expected pub.go to contain %q, got:\n%s", expected, content)
		}
	}
}

func readFacadeGoForTest(t *testing.T, outputDir string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(outputDir, facadeGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(content)
}
