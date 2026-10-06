package typescript

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestCastEnum(t *testing.T) {
	enum := castEnum(&schema.Enum{
		Name:        "UserStatus",
		Description: "User status",
		UnspecifiedItem: &schema.EnumItem{
			Name: "UNSPECIFIED",
		},
		Items: []*schema.EnumItem{
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
	if len(enum.CommentLines) == 0 || enum.CommentLines[0] != "User status." {
		t.Fatalf("unexpected enum comment lines: %+v", enum.CommentLines)
	}
	if len(enum.Items) != 3 {
		t.Fatalf("unexpected item count: %d", len(enum.Items))
	}
	if enum.Items[0].Literal != `"UNSPECIFIED"` {
		t.Fatalf("unexpected unspecified item literal: %s", enum.Items[0].Literal)
	}
	if enum.Items[1].Value != "ACTIVE" {
		t.Fatalf("unexpected first item value: %s", enum.Items[1].Value)
	}
	if len(enum.Items[1].CommentLines) == 0 || enum.Items[1].CommentLines[0] != "Active." {
		t.Fatalf("unexpected first item comment lines: %+v", enum.Items[1].CommentLines)
	}
}
