package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/golang/view"
	"go.yorun.ai/skel/internal/model"
)

func TestActorInfoIdentifierTag(t *testing.T) {
	for _, sensitive := range []bool{false, true} {
		pkg := buildModelDomainForTest(t, model.DomainSpec{
			Name: "demo.auth",
			Actors: []*model.Actor{{
				Name:            "UserActor",
				AuthEnabled:     true,
				IdentifierField: "userId",
				AuthCredential:  &model.Data{Name: "UserActorCredential"},
				AuthInfo: &model.Data{Name: "UserActorInfo", Members: []*model.DataMember{
					{Name: "userId", Type: codegentest.StringType(), Sensitive: sensitive},
				}},
			}},
		})
		outputDir := t.TempDir()
		gen := newGen(Option{Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull, PackageName: "skeled", Out: outputDir})
		gen.genActorGo()
		content, err := os.ReadFile(filepath.Join(outputDir, actorGoFilename))
		if err != nil {
			t.Fatal(err)
		}
		tag := `json:"userId" skel:"identifier"`
		if sensitive {
			tag = `json:"userId" skel:"sensitive,identifier"`
		}
		if !strings.Contains(string(content), tag) {
			t.Fatalf("missing tag %s in generated actor:\n%s", tag, content)
		}
	}
}
