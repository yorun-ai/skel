package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/golang/view"

	"go.yorun.ai/skel/internal/model"
)

func TestCastEnum(t *testing.T) {
	enum := castEnum(&model.Enum{
		Name:        "UserStatus",
		Description: "User status",
		UnspecifiedItem: &model.EnumItem{
			Name: "UNSPECIFIED",
		},
		Items: []*model.EnumItem{
			{
				Name:        "ACTIVE",
				Description: "Active",
			},
			{
				Name: "DISABLED",
			},
		},
	})

	if enum.Name != "UserStatus" {
		t.Fatalf("unexpected enum name: %s", enum.Name)
	}
	if enum.VarName != "userStatus" {
		t.Fatalf("unexpected enum var name: %s", enum.VarName)
	}
	if len(enum.CommentLines) == 0 || enum.CommentLines[0] != "UserStatus User status" {
		t.Fatalf("unexpected enum comment lines: %+v", enum.CommentLines)
	}
	if enum.UnspecifiedItem.Name != "UserStatusUnspecified" {
		t.Fatalf("unexpected unspecified item name: %s", enum.UnspecifiedItem.Name)
	}
	if enum.Items[0].Name != "UserStatusActive" {
		t.Fatalf("unexpected first item name: %s", enum.Items[0].Name)
	}
	if enum.Items[0].Value != "ACTIVE" {
		t.Fatalf("unexpected first item value: %s", enum.Items[0].Value)
	}
	if len(enum.Items[0].CommentLines) == 0 || enum.Items[0].CommentLines[0] != "UserStatusActive Active" {
		t.Fatalf("unexpected first item comment lines: %+v", enum.Items[0].CommentLines)
	}
}

func TestGenEnumGoAppendsJSONBytes(t *testing.T) {
	out := t.TempDir()
	gen := newGen(Option{
		PackageName: "example", Out: out,
		View: &view.Domain{Enums: []*model.Enum{
			{Name: "UserStatus", UnspecifiedItem: &model.EnumItem{Name: "UNSPECIFIED"}, Items: []*model.EnumItem{{Name: "ACTIVE"}}},
			{Name: "SiteType", UnspecifiedItem: &model.EnumItem{Name: "UNSPECIFIED"}, Items: []*model.EnumItem{{Name: "WEB"}}},
		}},
	})
	gen.genEnumGo()
	if err := gen.Renderer.Err(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(out, enumGoFilename))
	if err != nil {
		t.Fatal(err)
	}
	for _, variable := range []string{"userStatus", "siteType"} {
		want := `return fmt.Appendf(nil, "\"%s\"", ` + variable + `), nil`
		if !strings.Contains(string(content), want) {
			t.Fatalf("missing direct byte formatting %q in generated enum:\n%s", want, content)
		}
	}
	gen.genEnumGo()
	if err := gen.Renderer.Err(); err != nil {
		t.Fatal(err)
	}
	repeated, err := os.ReadFile(filepath.Join(out, enumGoFilename))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(repeated) {
		t.Fatal("enum generation is not deterministic")
	}
}
