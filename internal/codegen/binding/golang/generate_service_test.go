package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestGeneratorRendersNullableMapAndServiceHooks(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")

	profile := &schema.Data{
		Name: "Profile",
		Members: []*schema.DataMember{
			{Name: "aliases", Type: codegentest.ListType(codegentest.StringType())},
		},
	}
	user := &schema.Data{
		Name: "User",
		Members: []*schema.DataMember{
			{Name: "profile", Type: codegentest.DataType(profile)},
			{Name: "labels", Type: codegentest.NullableType(codegentest.MapType(codegentest.StringType(), codegentest.StringType()))},
			{Name: "friends", Type: codegentest.ListType(codegentest.NullableType(codegentest.DataType(profile)))},
			{Name: "profilesByName", Type: codegentest.MapType(codegentest.StringType(), codegentest.NullableType(codegentest.DataType(profile)))},
		},
	}
	pkg := newSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{profile, user},
		Services: []*schema.Service{
			{
				Name: "UserService",
				Methods: []*schema.Method{
					methodForTest("UserService", &schema.Method{
						Name:       "listUsers",
						ResultType: codegentest.ListType(codegentest.DataType(user)),
						Arguments: []*schema.Argument{
							{Name: "friends", Type: codegentest.ListType(codegentest.NullableType(codegentest.DataType(profile)))},
							{
								Name: "profilesByName",
								Type: codegentest.MapType(codegentest.StringType(), codegentest.NullableType(codegentest.DataType(profile))),
							},
						},
					}),
				}, Pub: true},
		},
	})

	if err := generateFixture(pkg, golang.Option{Out: goOutDir}); err != nil {
		t.Fatal(err)
	}

	goDataContent, err := os.ReadFile(filepath.Join(goOutDir, "data.go"))
	if err != nil {
		t.Fatalf("read go data file: %v", err)
	}
	if !strings.Contains(string(goDataContent), "Labels         *map[string]string") {
		t.Fatalf("expected nullable map pointer, got:\n%s", string(goDataContent))
	}

	goServiceContent, err := os.ReadFile(filepath.Join(goOutDir, "service.go"))
	if err != nil {
		t.Fatalf("read go service file: %v", err)
	}
	for _, tag := range []string{`json:"friends" skel:"index(0)"`, `json:"profilesByName" skel:"index(1)"`} {
		if !strings.Contains(string(goServiceContent), tag) {
			t.Fatalf("expected service argument tag %s, got:\n%s", tag, goServiceContent)
		}
	}
	for _, fragment := range []string{"CloneArguments", "CloneResult"} {
		if strings.Contains(string(goServiceContent), fragment) {
			t.Fatalf("service must not declare %s, got:\n%s", fragment, string(goServiceContent))
		}
	}
}
