package grammar

import (
	"strings"
	"testing"
)

// TODO: Remove this compatibility test when legacy via syntax and warnings are removed.
func TestParseActorAcceptsLegacyViaSections(t *testing.T) {
	content := parseSkelForTest(t, `domain demo.user
actor ClientActor {
    via client {}
    via openapi {}
    auth { credential { subject: string } }
    via Custom {}
}`)
	sections := content.Entries[0].Actor.Sections
	if len(sections) != 4 || sections[2].Auth == nil {
		t.Fatalf("unexpected actor sections: %+v", sections)
	}
	for _, test := range []struct {
		index int
		name  string
	}{{0, "client"}, {1, "openapi"}, {3, "Custom"}} {
		via := sections[test.index].Via
		if via == nil || via.Name.Value != test.name {
			t.Fatalf("unexpected legacy via section %d: %+v", test.index, via)
		}
	}
}

func TestParseActorAuthCredentialAndInfoCanAppearInEitherOrder(t *testing.T) {
	content := parseSkelForTest(t, `
domain demo.user

actor ClientActor {
    auth {
        info {
            userId: int
        }
        credential {
            subject: string
        }
    }
}
	`)
	actor := content.Entries[0].Actor
	if len(actor.Sections) != 1 || actor.Sections[0].Auth == nil || actor.Sections[0].Auth.Info == nil || actor.Sections[0].Auth.Credential == nil {
		t.Fatalf("unexpected actor sections: %+v", actor.Sections)
	}
}

func TestParseActorPermissionMustBeEmpty(t *testing.T) {
	parser := buildSkelParserForTest(t)
	_, err := parser.ParseString("test.skel", strings.TrimSpace(`
domain demo.user

actor ClientActor {
    permission {
        User:read
    }
}
`))
	if err == nil {
		t.Fatal("expected parse error when actor permission block is not empty")
	}
}

func TestParseResourceCheckUsesInputBlock(t *testing.T) {
	content := parseSkelForTest(t, `
domain demo.user

resource User {
    @desc("Lookup")
    check byId {
        @desc("Lookup input")
        input {
            @sensitive
            id: int
        }
    }
    action read
}
`)
	check := content.Entries[0].Resource.Checks[0]
	if check.Name.Value != "byId" || check.Input == nil || len(check.Input.Arguments) != 1 {
		t.Fatalf("unexpected resource check: %+v", check)
	}
	if len(check.Decorators) != 1 || len(check.Input.Decorators) != 1 || len(check.Input.Arguments[0].Decorators) != 1 {
		t.Fatalf("resource check decorators were not preserved: %+v", check)
	}

}

func serviceMethodsForTest(service *Service) []*Method {
	methods := make([]*Method, 0)
	for _, section := range service.Sections {
		if section.Method != nil {
			section.Method.Decorators = append(section.Decorators, section.Method.Decorators...)
			methods = append(methods, section.Method)
		}
	}
	return methods
}

func TestWebMountSyntax(t *testing.T) {
	for _, body := range []string{
		"for ClientActor\nmount /portal/v1-assets/",
		"for ClientActor\nmount /* prefix */ /portal/v1-assets/ // trailing comment",
		"for ClientActor mount /portal/v1-assets/",
	} {
		content := parseSkelForTest(t, "domain demo\nweb PortalWeb {\n"+body+"\n}")
		web := content.Entries[0].Web
		if len(web.Mounts) != 1 || web.Mounts[0].Path.Value != "/portal/v1-assets/" || len(web.Audiences) != 1 {
			t.Fatalf("unexpected web: %+v", web)
		}
	}
	content := parseSkelForTest(t, "domain demo\nweb PortalWeb {\n    mount /\n    for ClientActor\n}")
	path := content.Entries[0].Web.Mounts[0].Path
	if path.Value != "/" || path.Pos.Filename != "test.skel" || path.Pos.Line != 3 || path.Pos.Column != 11 {
		t.Fatalf("unexpected path or position: %+v", path)
	}
}
