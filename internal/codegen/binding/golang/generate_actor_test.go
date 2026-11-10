package golang_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
)

func TestActorAuthResultGeneration(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "actor.skel")
	shared := filepath.Join(root, "ops.skel")
	writeFileForTest(t, shared, `domain shared
pub actor OpsActor { auth { credential { token: string } info { id: string } } }`)
	writeFileForTest(t, entry, `domain demo
import shared as ops
pub actor UserActor {
    auth {
        impersonator ops.OpsActor, UserActor
        credential { token: string }
        info { @identifier id: string }
    }
}`)
	input := api.Input{SkelIn: entry, SkelImports: map[string]string{"shared": shared}}
	out := filepath.Join(root, "generated")
	_, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Imports: map[string]string{"shared": "example.com/shared"}})
	if err != nil {
		t.Fatal(err)
	}
	actor := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "actor.go"))), " ")
	for _, want := range []string{"type UserActorAuthResult struct", "Info UserActorInfo", "ImpersonatorInfo *UserActorImpersonatorInfo", "OpsActorInfo *ops.OpsActorInfo", "UserActorInfo *UserActorInfo", "Auth(credential UserActorCredential) UserActorAuthResult", "reflect.TypeFor[UserActorAuthResult]()"} {
		if !strings.Contains(actor, want) {
			t.Fatalf("missing %s in %s", want, actor)
		}
	}
	descriptor := readFileForTest(t, filepath.Join(out, "descriptor.go"))
	for _, want := range []string{`InfoFieldName: "opsActorInfo"`, `ActorSkelName: "shared.OpsActor"`, `Name: "UserActorAuthResult"`, `Name: "UserActorImpersonatorInfo"`, `Name: "impersonatorInfo"`} {
		if !strings.Contains(descriptor, want) {
			t.Fatalf("missing %s in descriptor", want)
		}
	}
	original := actor
	if _, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Imports: map[string]string{"shared": "example.com/shared"}}); err != nil {
		t.Fatal(err)
	}
	if original != strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "actor.go"))), " ") {
		t.Fatal("nondeterministic actor output")
	}
}

func TestActorImpersonatorImports(t *testing.T) {
	for _, mode := range []string{"full", "public", "split"} {
		for _, alias := range []string{"", "ops"} {
			t.Run(mode+"/"+alias, func(t *testing.T) {
				root := t.TempDir()
				entry := filepath.Join(root, "app.skel")
				shared := filepath.Join(root, "shared.skel")
				writeFileForTest(t, shared, `domain shared
pub actor OpsActor { auth { credential { token: string } info { id: string } } }`)
				importClause, qualifier, goQualifier := "import shared", "shared", "sharedpub"
				if alias != "" {
					importClause += " as " + alias
					qualifier, goQualifier = alias, alias
				}
				// A local namesake must not capture the imported operator's info type.
				writeFileForTest(t, entry, "domain demo\n"+importClause+`
actor OpsActor { auth { credential { token: string } info { id: int } } }
pub actor UserActor { auth {
    impersonator `+qualifier+`.OpsActor, UserActor
    credential { token: string }
    info { id: string }
} }`)
				input := api.Input{SkelIn: entry, SkelImports: map[string]string{"shared": shared}}
				out := filepath.Join(root, "generated")
				option := api.GolangOption{
					CompilerVersion: "v0.0.0-dev",
					Out:             out,
					AsModule:        true,
					Module:          "example.com/demo",
					PubOnly:         mode == "public",
					Imports:         map[string]string{"shared": "example.com/shared@v1.0.0"},
				}
				if mode == "split" {
					option.PubOut = filepath.Join(root, "public")
					out = option.PubOut
				}
				if _, err := api.CompileGolang(input, option); err != nil {
					t.Fatal(err)
				}
				actor := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "actor.go"))), " ")
				for _, want := range []string{
					`"example.com/shared"`,
					"OpsActorInfo *" + goQualifier + ".OpsActorInfo",
					"UserActorInfo *UserActorInfo",
				} {
					if !strings.Contains(actor, want) {
						t.Fatalf("missing %q in actor output: %s", want, actor)
					}
				}
				module := readFileForTest(t, filepath.Join(out, "go.mod"))
				if !strings.Contains(module, "example.com/shared v1.0.0") {
					t.Fatalf("missing operator module dependency: %s", module)
				}
				if mode == "split" {
					facade := readFileForTest(t, filepath.Join(option.Out, "pub.go"))
					for _, want := range []string{
						"type UserActorAuthResult = demopub.UserActorAuthResult",
						"type UserActorImpersonatorInfo = demopub.UserActorImpersonatorInfo",
					} {
						if !strings.Contains(facade, want) {
							t.Fatalf("missing %q in facade: %s", want, facade)
						}
					}
				}
			})
		}
	}
}
