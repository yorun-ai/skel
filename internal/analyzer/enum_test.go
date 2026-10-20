package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestParseEnumReturnsErrorForDuplicatedItem(t *testing.T) {
	expectEnumDiagnostic(t, "duplicated EnumItem ACTIVE", &grammar.Enum{
		Name: ident("UserStatus"),
		Items: []*grammar.EnumItem{
			{Name: ident("ACTIVE")},
			{Name: ident("ACTIVE")},
		},
	})
}

func TestParseEnumReturnsErrorForReservedKindSuffix(t *testing.T) {
	for _, name := range []string{"UserConfig", "UserEvent", "UserActor", "UserService", "UserWeb"} {
		t.Run(name, func(t *testing.T) {
			expectEnumDiagnostic(t, "Enum name must not end with", &grammar.Enum{
				Name:  ident(name),
				Items: []*grammar.EnumItem{{Name: ident("ACTIVE")}},
			})
		})
	}
}

func TestParseEnumReturnsErrorForReservedUnspecifiedItem(t *testing.T) {
	expectEnumDiagnostic(t, "reversed EnumItem value UNSPECIFIED", &grammar.Enum{
		Name: ident("UserStatus"),
		Items: []*grammar.EnumItem{
			{Name: ident("UNSPECIFIED")},
		},
	})
}
