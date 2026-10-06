package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestParseWeb(t *testing.T) {
	web := parseWebTest(t, &grammar.Web{
		Decorators: []*grammar.Decorator{
			{Name: ident("desc"), Value: decoratorValue(`"User portal entry point"`)},
		},
		Name:      ident("UserPortalWeb"),
		Audiences: []*grammar.WebAudience{webAllow("ClientActor", "client"), webAllow("OpenAPIActor")},
	}, false)

	if web.Name != "UserPortalWeb" {
		t.Fatalf("unexpected web name: %s", web.Name)
	}
	if web.Description != "User portal entry point" {
		t.Fatalf("unexpected web description: %q", web.Description)
	}
	if len(web.Audiences) != 2 || web.Audiences[0].Actor != "ClientActor" || web.Audiences[1].Actor != "OpenAPIActor" {
		t.Fatalf("unexpected web audiences: %+v", web.Audiences)
	}
	if web.Audiences[0].Via != "client" {
		t.Fatalf("unexpected web for via: %+v", web.Audiences)
	}
}

func TestParseWebRejectsPub(t *testing.T) {
	expectWebDiagnostic(t, "does not support pub", &grammar.Web{
		Name:      ident("UserPortalWeb"),
		Audiences: []*grammar.WebAudience{webAllow("ClientActor")},
	}, true)
}

func TestParseWebRejectsMissingActors(t *testing.T) {
	expectWebDiagnostic(t, "must declare at least one actor", &grammar.Web{
		Name: ident("UserPortalWeb"),
	}, false)
}
