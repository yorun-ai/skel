package golang_test

import (
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/testutil"
)

func TestGeneratedAuthModesWithPublishedVine(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	input := filepath.Join(root, "domain.skel")
	writeFileForTest(t, input, `domain demo.auth
actor ClientActor { via client {} }
api service SessionApiService {
 for ClientActor via client
 auth required
 method profile {}
 method browse { auth optional }
 method login { auth anonymous }
}
web RequiredWeb { for ClientActor via client auth required }
web OptionalWeb { for ClientActor via client auth optional }
web AnonymousWeb { for ClientActor via client auth anonymous }
web OffWeb { for ClientActor via client auth off }
`)
	out := filepath.Join(root, "auth")
	if _, err := api.CompileGolang(api.Input{SkelIn: input, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Module: "example.com/auth", AsModule: true}); err != nil {
		t.Fatal(err)
	}
	writeFileForTest(t, filepath.Join(out, "auth_test.go"), `package auth
import (
 "testing"
 "go.yorun.ai/vine/core/skel"
)
func TestAuthModes(t *testing.T) {
 if _DomainSchema.Services[0].AuthMode != skel.AuthModeRequired { t.Fatal("wrong service auth") }
 methods:=_DomainSchema.Services[0].Methods
 want:=map[string]skel.AuthMode{"profile":skel.AuthModeInherit,"browse":skel.AuthModeOptional,"login":skel.AuthModeAnonymous}
 for _,method:=range methods { if method.AuthMode!=want[method.Name] { t.Fatalf("method %s: %s",method.Name,method.AuthMode) } }
 webs:=map[string]skel.AuthMode{"RequiredWeb":skel.AuthModeRequired,"OptionalWeb":skel.AuthModeOptional,"AnonymousWeb":skel.AuthModeAnonymous,"OffWeb":skel.AuthModeOff}
 for _,web:=range _DomainSchema.Webs { if web.AuthMode!=webs[web.Name] { t.Fatalf("web %s: %s",web.Name,web.AuthMode) } }
}
`)
	testutil.Go(t, out, "test", "-mod=mod", "./...")
}
