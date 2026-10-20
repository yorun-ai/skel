package analyzer

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

func TestParseDataRejectsInvalidDeprecatedDecorator(t *testing.T) {
	for name, test := range map[string]struct {
		decorators []*grammar.Decorator
		message    string
	}{
		"missing argument": {
			decorators: []*grammar.Decorator{{Name: ident("deprecated")}},
			message:    "decorator @deprecated requires a deprecation reason",
		},
		"empty reason": {
			decorators: []*grammar.Decorator{{Name: ident("deprecated"), Value: decoratorValue(`"  "`)}},
			message:    "decorator @deprecated requires a deprecation reason",
		},
		"non-string": {
			decorators: []*grammar.Decorator{{Name: ident("deprecated"), Value: decoratorValue(`true`)}},
			message:    "invalid decorator @deprecated",
		},
		"duplicated": {
			decorators: []*grammar.Decorator{
				{Name: ident("deprecated"), Value: decoratorValue(`"first"`)},
				{Name: ident("deprecated"), Value: decoratorValue(`"second"`)},
			},
			message: "duplicated decorator @deprecated",
		},
	} {
		t.Run(name, func(t *testing.T) {
			expectDataDiagnostic(t, test.message, &grammar.Data{
				Decorators: test.decorators,
				Name:       ident("User"),
			})
		})
	}
}

func TestDomainRejectsDeprecatedDecorator(t *testing.T) {
	domain := &grammar.DomainContent{
		Decorators: []*grammar.Decorator{
			{Name: ident("deprecated"), Value: decoratorValue(`"Domains are not deprecable"`)},
		},
		Name: qualifiedName("demo"),
	}
	err := domain.Finalize()
	if err == nil || !strings.Contains(err.Error(), "unexpected decorator @deprecated") {
		t.Fatalf("expected domain deprecation diagnostic, got %v", err)
	}
}

func TestParseDataAcceptsGoMarkerFieldName(t *testing.T) {
	data := parseDataTest(t, &grammar.Data{
		Decorators: []*grammar.Decorator{{Name: ident("sensitive")}},
		Name:       ident("Credential"),
		Members: []*grammar.DataMember{{
			Name: ident("skelSensitive"),
			Type: plainType(grammar.String),
		}},
	})
	if data.Members[0].Name != "skelSensitive" || !data.Sensitive {
		t.Fatalf("unexpected sensitive data: %+v", data)
	}
}

func TestParseDataRejectsInvalidSensitiveDecorator(t *testing.T) {
	t.Run("argument", func(t *testing.T) {
		expectDataDiagnostic(t, "decorator @sensitive does not accept an argument", &grammar.Data{
			Name: ident("User"),
			Members: []*grammar.DataMember{{
				Decorators: []*grammar.Decorator{
					{Name: ident("sensitive"), Value: decoratorValue(`true`)},
				},
				Name: ident("password"),
				Type: plainType(grammar.String),
			}},
		})
	})
	t.Run("duplicated", func(t *testing.T) {
		expectDataDiagnostic(t, "duplicated decorator @sensitive", &grammar.Data{
			Name: ident("User"),
			Members: []*grammar.DataMember{{
				Decorators: []*grammar.Decorator{
					{Name: ident("sensitive")},
					{Name: ident("sensitive")},
				},
				Name: ident("password"),
				Type: plainType(grammar.String),
			}},
		})
	})
}

func TestParseDataReturnsErrorForDuplicatedMember(t *testing.T) {
	expectDataDiagnostic(t, "duplicated DataMember id", &grammar.Data{
		Name: ident("User"),
		Members: []*grammar.DataMember{
			{Name: ident("id"), Type: plainType(grammar.Int)},
			{Name: ident("id"), Type: plainType(grammar.String)},
		},
	})
}

func TestParseDataReturnsErrorForReservedKindSuffix(t *testing.T) {
	for _, name := range []string{"UserConfig", "UserEvent", "UserActor", "UserService", "UserWeb"} {
		t.Run(name, func(t *testing.T) {
			expectDataDiagnostic(t, "Data name must not end with", &grammar.Data{
				Name: ident(name),
			})
		})
	}
}

func TestParseDataReturnsErrorForNullableTypeParameter(t *testing.T) {
	expectDataDiagnostic(t, "TypeParameter TItem cannot be nullable", &grammar.Data{
		Name: ident("Page"),
		TypeParameters: []*grammar.TypeParameter{
			{Name: ident("TItem"), Nullable: true},
		},
	})
}

func TestParseDataReturnsErrorWhenMemberExampleHasNoDescription(t *testing.T) {
	expectDataDiagnostic(t, "decorator @example must be used with @desc", &grammar.Data{
		Name: ident("UserProfile"),
		Members: []*grammar.DataMember{
			{
				Decorators: []*grammar.Decorator{
					{Name: ident("example"), Value: decoratorValue(`"https://xxx.com/a.png"`)},
				},
				Name: ident("avatarUrl"),
				Type: nullableType(plainType(grammar.String)),
			},
		},
	})
}

func TestParseConfigReturnsErrorWhenNameDoesNotEndWithConfig(t *testing.T) {
	expectConfigDiagnostic(t, "Config name must end with Config", &grammar.Data{
		Name:      ident("Database"),
		Qualifier: ident("eternal"),
	})
}

func TestParseConfigReturnsErrorForTypeParameters(t *testing.T) {
	expectConfigDiagnostic(t, "does not support type parameters", &grammar.Data{
		Name:      ident("PageConfig"),
		Qualifier: ident("instant"),
		TypeParameters: []*grammar.TypeParameter{
			{Name: ident("TItem")},
		},
	})
}

func TestParseConfigReturnsErrorWithoutLifecycleQualifier(t *testing.T) {
	expectConfigDiagnostic(t, "requires lifecycle qualifier eternal/instant", &grammar.Data{
		Name: ident("DatabaseConfig"),
	})
}

func TestParseConfigReturnsErrorForInvalidLifecycleQualifier(t *testing.T) {
	expectConfigDiagnostic(t, "invalid lifecycle qualifier", &grammar.Data{
		Name:      ident("DatabaseConfig"),
		Qualifier: ident("cached"),
	})
}

func TestParseDataReturnsErrorForQualifier(t *testing.T) {
	expectDataDiagnostic(t, "does not support qualifier", &grammar.Data{
		Name:      ident("User"),
		Qualifier: ident("eternal"),
	})
}

func TestParseEvent(t *testing.T) {
	event := parseEventTest(t, &grammar.Event{
		Decorators: []*grammar.Decorator{
			{Name: ident("desc"), Value: decoratorValue(`"User created event"`)},
		},
		Pub:  true,
		Name: ident("UserCreatedEvent"),
		Payload: &grammar.EventPayload{
			Decorators: []*grammar.Decorator{{Name: ident("sensitive")}},
			Members: []*grammar.DataMember{
				{
					Decorators: []*grammar.Decorator{
						{Name: ident("desc"), Value: decoratorValue(`"User ID"`)},
					},
					Name: ident("userId"),
					Type: plainType(grammar.Int),
				},
			},
		},
	})
	if event.Kind != schema.DataKindEvent {
		t.Fatalf("unexpected event kind: %v", event.Kind)
	}
	if !event.Pub {
		t.Fatal("expected pub event")
	}
	if !event.Sensitive {
		t.Fatal("expected event payload to be sensitive as a whole")
	}
}

func TestParseEventRejectsSensitiveDeclaration(t *testing.T) {
	expectEventDiagnostic(t, "unexpected decorator @sensitive", &grammar.Event{
		Decorators: []*grammar.Decorator{{Name: ident("sensitive")}},
		Name:       ident("UserCreatedEvent"),
		Payload:    &grammar.EventPayload{},
	})
}

func TestParseConfigAllowsPub(t *testing.T) {
	config := parseConfigTest(t, &grammar.Data{
		Decorators: []*grammar.Decorator{{Name: ident("sensitive")}},
		Pub:        true,
		Name:       ident("DatabaseConfig"),
		Qualifier:  ident("instant"),
	})
	if !config.Pub {
		t.Fatal("expected pub config")
	}
	if !config.Sensitive {
		t.Fatal("expected config to be sensitive as a whole")
	}
}

func TestParseConfigRejectsUnknownDecorator(t *testing.T) {
	for _, lifecycle := range []string{"eternal", "instant"} {
		t.Run(lifecycle, func(t *testing.T) {
			for _, decorators := range [][]*grammar.Decorator{
				{{Name: ident("unknownDecorator")}},
				{{Name: ident("sensitive")}, {Name: ident("unknownDecorator")}},
				{{Name: ident("unknownDecorator"), Value: decoratorValue("true")}},
			} {
				expectConfigDiagnostic(t, "unexpected decorator @unknownDecorator", &grammar.Data{
					Name: ident("TextConfig"), Qualifier: ident(lifecycle),
					Members: []*grammar.DataMember{{Name: ident("text"), Type: plainType(grammar.String), Decorators: decorators}},
				})
			}
		})
	}
}

func TestParseAbstractConfigName(t *testing.T) {
	for _, name := range []string{"DatabaseAbsConfig", "DatabaseAbsConfigV2"} {
		t.Run(name, func(t *testing.T) {
			parseConfigTest(t, &grammar.Data{Abs: true, Name: ident(name), Qualifier: ident("eternal")})
		})
	}
	for _, name := range []string{"DatabaseConfig", "DatabaseConfigV2", "DatabaseAbsConfigV0", "DatabaseAbsConfigV01", "DatabaseAbsConfigV", "DatabaseAbsConfigV-1"} {
		t.Run(name, func(t *testing.T) {
			expectConfigDiagnostic(t, "Config name must end with AbsConfig", &grammar.Data{Abs: true, Name: ident(name), Qualifier: ident("eternal")})
		})
	}
	for _, name := range []string{"AbsConfig", "AbsConfigV2"} {
		t.Run(name, func(t *testing.T) {
			expectConfigDiagnostic(t, "missing body after trimming prefix & suffix", &grammar.Data{Abs: true, Name: ident(name), Qualifier: ident("eternal")})
		})
	}
}

func TestParseConcreteConfigRejectsAbstractSuffix(t *testing.T) {
	for _, pub := range []bool{false, true} {
		for _, name := range []string{"DatabaseAbsConfig", "DatabaseAbsConfigV2"} {
			expectConfigDiagnostic(t, "AbsConfig suffix is reserved for abs config", &grammar.Data{Pub: pub, Name: ident(name), Qualifier: ident("eternal")})
		}
	}
}
