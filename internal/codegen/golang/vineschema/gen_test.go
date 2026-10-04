package vineschema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/codegentest"
	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
)

func TestGenSchemaGoRendersActorAuthEnabled(t *testing.T) {
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Actors: []*model.Actor{
			{
				Name:            "ClientActor",
				Vias:            []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
				AuthEnabled:     true,
				IdentifierField: "userId",
				AuthCredential: &model.Data{
					Name: "ClientActorCredential",
					Members: []*model.DataMember{
						{Name: "token", Type: codegentest.StringType()},
					},
				},
				AuthInfo: &model.Data{
					Name: "ClientActorInfo",
					Members: []*model.DataMember{
						{Name: "userId", Type: codegentest.IntType()},
					},
				},
			},
			{
				Name: "AnonymousActor",
				Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
			},
		},
	})

	outputDir := filepath.Join(t.TempDir(), "skeled")
	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         outputDir,
	})
	gen.gen()

	content, err := os.ReadFile(filepath.Join(outputDir, schemaGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	codegentest.AssertGoSourceContains(t, string(content), `IdentifierField: "userId"`)
	codegentest.AssertGoSourceContains(t, string(content), "AuthEnabled: true")
	codegentest.AssertGoSourceContains(t, string(content), "AuthEnabled: false")
	codegentest.AssertGoSourceContains(t, string(content), "PermEnabled: false")
}

func TestGenSchemaGoRendersDeprecatedFields(t *testing.T) {
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{{
			Name:             "User",
			Deprecated:       true,
			DeprecatedReason: "Use Profile instead",
			Members: []*model.DataMember{{
				Name:             "legacyId",
				Type:             codegentest.StringType(),
				Deprecated:       true,
				DeprecatedReason: "Use id instead",
			}},
		}},
	})
	outputDir := filepath.Join(t.TempDir(), "skeled")
	newGen(Option{
		Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull,
		PackageName: "skeled", Out: outputDir,
	}).gen()

	content, err := os.ReadFile(filepath.Join(outputDir, schemaGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	generated := string(content)
	if strings.Count(generated, "Deprecated:") < 2 {
		t.Fatalf("expected data and member deprecation flags, got:\n%s", generated)
	}
	for _, reason := range []string{"Use Profile instead", "Use id instead"} {
		if !strings.Contains(generated, `"`+reason+`"`) {
			t.Fatalf("expected reason %q, got:\n%s", reason, generated)
		}
	}
}

func TestGenSchemaGoRendersWebAuthModes(t *testing.T) {
	for _, mode := range []model.AuthMode{model.AuthModeUnset, model.AuthModeRequired, model.AuthModeOptional, model.AuthModeGuest, model.AuthModeOff, model.AuthModeAuth, model.AuthModeNoAuth} {
		t.Run(string(mode), func(t *testing.T) {
			pkg := buildModelDomainForTest(t, model.DomainSpec{Name: "demo.user", Webs: []*model.Web{{Name: "ConsoleWeb", Auth: mode}}})
			out := filepath.Join(t.TempDir(), "skeled")
			newGen(Option{Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull, PackageName: "skeled", Out: out}).gen()
			content, err := os.ReadFile(filepath.Join(out, schemaGoFilename))
			if err != nil {
				t.Fatal(err)
			}
			if mode == model.AuthModeUnset {
				if strings.Contains(string(content), "AuthMode:") {
					t.Fatalf("unset web mode should preserve legacy field omission: %s", content)
				}
			} else {
				literal, err := renderAuthModeLiteral(_AuthMode(mode))
				if err != nil {
					t.Fatal(err)
				}
				codegentest.AssertGoSourceContains(t, string(content), "AuthMode: "+literal)
			}
		})
	}
}
