package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestParseWebRejectsPub(t *testing.T) {
	expectWebDiagnostic(t, "does not support pub", &grammar.Web{
		Sections:  []*grammar.WebSection{{Auth: &grammar.AuthMarker{Value: "required"}}},
		Name:      ident("UserPortalWeb"),
		Audiences: []*grammar.ActorAudience{actorAudience("ClientActor")},
	}, true)
}

func TestParseWebRejectsMissingActors(t *testing.T) {
	expectWebDiagnostic(t, "must declare at least one actor", &grammar.Web{
		Sections: []*grammar.WebSection{{Auth: &grammar.AuthMarker{Value: "required"}}},
		Name:     ident("UserPortalWeb"),
	}, false)
}
