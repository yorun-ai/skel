package api_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/schema"
	"go.yorun.ai/skel/schema/diff"
)

func TestImpersonatorContractAcrossBindings(t *testing.T) {
	root := t.TempDir()
	entry, shared := filepath.Join(root, "app.skel"), filepath.Join(root, "shared.skel")
	writeTestFile(t, shared, `domain shared
pub actor OpsActor { auth { credential { token: string } info { id: string } } }`)
	source := `domain demo
import shared as ops
pub actor UserActor { auth {
    impersonator ops.OpsActor
    credential { session: string }
    info { @identifier id: int }
} }
api service UserApiService { for UserActor auth required method ping { output string } }
`
	writeTestFile(t, entry, source)
	input := api.Input{SkelIn: entry, SkelImports: map[string]string{"shared": shared}}
	parsed, err := api.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	dependencies := schema.Dependencies(parsed.Domain)
	found := false
	for _, dep := range dependencies.Dependencies {
		if dep.Kind == "actor" && dep.Domain == "shared" && dep.Name == "OpsActor" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing operator dependency: %+v", dependencies)
	}
	pub := filepath.Join(root, "contract")
	if _, err := api.CompileSkeleton(input, api.SkeletonOption{Out: pub, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	exported, err := api.Parse(api.Input{SkelIn: pub, SkelImports: input.SkelImports})
	if err != nil {
		t.Fatal(err)
	}
	if exported.Domain.Actors()[0].Auth.Impersonators[0].ActorSkelName != "shared.OpsActor" {
		t.Fatal("public Skel lost impersonator")
	}
	if parsed.Domain.Actors()[0].Hash != exported.Domain.Actors()[0].Hash {
		t.Fatal("public Skel changed actor contract")
	}
	goOut := filepath.Join(root, "regular")
	if _, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: goOut, PubOut: goOut + "pub", AsModule: true, Module: "example.com/demo", Imports: map[string]string{"shared": "example.com/shared@v1.0.0"}}); err != nil {
		t.Fatal(err)
	}
	// API-only TypeScript must remain independent of backend auth callbacks.
	if _, err := api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, Out: filepath.Join(root, "typescript")}); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, entry, strings.Replace(source, "impersonator ops.OpsActor", "", 1))
	plain, err := api.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if plain.Domain.Actors()[0].Hash == parsed.Domain.Actors()[0].Hash {
		t.Fatal("operator declaration did not change actor hash")
	}
	report, err := diff.Compare(plain.Domain, parsed.Domain)
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, change := range report.Changes {
		if change.Code == "actor.impersonators.changed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing impersonator diff: %+v", report)
	}
}
