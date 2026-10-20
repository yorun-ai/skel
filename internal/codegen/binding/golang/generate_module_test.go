package golang_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestGeneratorRendersGoPubAndRegularModules(t *testing.T) {
	tmpDir := t.TempDir()
	goOutDir := filepath.Join(tmpDir, "user")
	goPubOutDir := filepath.Join(tmpDir, "userpub")

	user := &schema.Data{
		Pub:  true,
		Name: "User",
		Members: []*schema.DataMember{
			{Name: "id", Type: codegentest.StringType()},
		},
	}
	pkg := newSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{user},
		Services: []*schema.Service{
			{
				Pub:  true,
				Name: "UserService",
				Methods: []*schema.Method{
					methodForTest("UserService", &schema.Method{Name: "getUser", ResultType: codegentest.DataType(user)}),
				},
			},
		},
		Events: []*schema.Data{
			{
				Pub:  true,
				Name: "UserChangedEvent",
				Members: []*schema.DataMember{
					{Name: "user", Type: codegentest.DataType(user)},
				},
			},
		},
	})

	if err := generateFixture(pkg, golang.Option{
		Out:          goOutDir,
		AsModule:     true,
		PubOut:       goPubOutDir,
		ModulePrefix: "github.com/acme/skel",
		VineVersion:  "v1.8.7",
	}); err != nil {
		t.Fatal(err)
	}

	pubServiceContent := readFileForTest(t, filepath.Join(goPubOutDir, "service.go"))
	if !strings.Contains(pubServiceContent, "rpc.ServiceSpecTypeClient") {
		t.Fatalf("expected pub service client spec, got:\n%s", pubServiceContent)
	}
	if strings.Contains(pubServiceContent, "type UserServiceServer interface") {
		t.Fatalf("did not expect pub service server, got:\n%s", pubServiceContent)
	}
	pubEventContent := readFileForTest(t, filepath.Join(goPubOutDir, "event.go"))
	if !strings.Contains(pubEventContent, "event.EventSpecTypeListener") {
		t.Fatalf("expected pub event listener spec, got:\n%s", pubEventContent)
	}
	if strings.Contains(pubEventContent, "type UserChangedEventEmitter interface") {
		t.Fatalf("did not expect pub event emitter, got:\n%s", pubEventContent)
	}

	regularServiceContent := readFileForTest(t, filepath.Join(goOutDir, "service.go"))
	if !strings.Contains(regularServiceContent, "rpc.ServiceSpecTypeServer") {
		t.Fatalf("expected regular pub service server spec, got:\n%s", regularServiceContent)
	}
	if strings.Contains(regularServiceContent, "type UserServiceClient interface") {
		t.Fatalf("did not expect regular pub service client, got:\n%s", regularServiceContent)
	}
	regularEventContent := readFileForTest(t, filepath.Join(goOutDir, "event.go"))
	if !strings.Contains(regularEventContent, "event.EventSpecTypeEmitter") {
		t.Fatalf("expected regular pub event emitter spec, got:\n%s", regularEventContent)
	}
	if strings.Contains(regularEventContent, "type UserChangedEventListener interface") {
		t.Fatalf("did not expect regular pub event listener, got:\n%s", regularEventContent)
	}

	pubDescriptorContent := readFileForTest(t, filepath.Join(goPubOutDir, "descriptor.go"))
	for _, name := range []string{"User", "UserService", "UserChangedEvent"} {
		codegentest.AssertGoSourceContains(t, pubDescriptorContent, `Name: "`+name+`"`)
	}
	codegentest.AssertGoSourceContains(t, pubDescriptorContent, "Full: false")
	regularDescriptorContent := readFileForTest(t, filepath.Join(goOutDir, "descriptor.go"))
	codegentest.AssertGoSourceContains(t, regularDescriptorContent, "Full: true")
	for _, name := range []string{"User", "UserService", "UserChangedEvent"} {
		codegentest.AssertGoSourceContains(t, regularDescriptorContent, `Name: "`+name+`"`)
	}

	regularGoModContent := readFileForTest(t, filepath.Join(goOutDir, "go.mod"))
	if !strings.Contains(regularGoModContent, "github.com/acme/skel/demo/userpub") {
		t.Fatalf("expected regular module to require pub module, got:\n%s", regularGoModContent)
	}
	facadeContent := readFileForTest(t, filepath.Join(goOutDir, "pub.go"))
	if !strings.Contains(facadeContent, `import "github.com/acme/skel/demo/userpub"`) ||
		!strings.Contains(facadeContent, "type User = userpub.User") ||
		!strings.Contains(facadeContent, "type UserServiceClient = userpub.UserServiceClient") ||
		!strings.Contains(facadeContent, "type UserChangedEventListener = userpub.UserChangedEventListener") {
		t.Fatalf("expected regular pub facade, got:\n%s", facadeContent)
	}
}
