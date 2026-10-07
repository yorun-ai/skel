package source

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestCastWeb(t *testing.T) {
	web := new(_Gen).castWeb(&schema.Web{
		Name:        "UserPortalWeb",
		SkelName:    "demo.user.UserPortalWeb",
		Description: "User portal entry point",
		Audiences:   []*schema.ActorAudience{{Actor: "ClientActor"}},
	})

	if web.ServerName != "UserPortalWebServer" {
		t.Fatalf("unexpected server name: %s", web.ServerName)
	}
	if web.DefaultServerName != "DefaultUserPortalWebServer" {
		t.Fatalf("unexpected default server name: %s", web.DefaultServerName)
	}
	if web.SpecName != "_UserPortalWebSpec" {
		t.Fatalf("unexpected spec name: %s", web.SpecName)
	}
	if len(web.CommentLines) == 0 || web.CommentLines[0] != "UserPortalWebServer User portal entry point" {
		t.Fatalf("unexpected comment lines: %+v", web.CommentLines)
	}
}
