package compiler

import (
	"go.yorun.ai/skel/schema"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileExplicitAuthModes(t *testing.T) {
	for _, owner := range []string{"service", "web"} {
		for _, mode := range []string{"required", "optional", "anonymous", "off"} {
			if owner == "service" && mode == "off" {
				continue
			}
			t.Run(owner+"/"+mode, func(t *testing.T) {
				declaration := "api service UserApiService { for ClientActor via client auth " + mode + " method ping {} }"
				if owner == "web" {
					declaration = "web ConsoleWeb { for ClientActor via client auth " + mode + " }"
				}
				path := filepath.Join(t.TempDir(), "input.skel")
				writeFile(t, path, "domain demo.user\nactor ClientActor { via client {} }\n"+declaration)
				for _, strict := range []bool{false, true} {
					result, err := Compile(Option{SkelIn: path, Strict: strict})
					if err != nil {
						t.Fatal(err)
					}
					if len(result.Diagnostics) != 0 {
						t.Fatal(result.Diagnostics)
					}
					if owner == "web" {
						if result.Domain.Webs()[0].NormalizedAuth() != schema.AuthMode(mode) {
							t.Fatal(result.Domain.Webs()[0])
						}
					} else if method := result.Domain.Services()[0].Methods[0]; method.NormalizedAuth() != schema.AuthModeInherit || method.EffectiveAuthMode != schema.AuthMode(mode) {
						t.Fatal(method)
					}
				}
			})
		}
	}
}

func TestCompileRejectsUnsupportedAuthSyntax(t *testing.T) {
	for _, marker := range []string{"auth", "noauth", "auth unknown"} {
		for _, declaration := range []string{
			"api service EntryApiService { for ClientActor " + marker + " method ping {} }",
			"api service EntryApiService { for ClientActor auth required method ping { " + marker + " } }",
			"web EntryWeb { for ClientActor " + marker + " }",
		} {
			path := filepath.Join(t.TempDir(), "input.skel")
			writeFile(t, path, "domain demo\nactor ClientActor { via client {} }\n"+declaration)
			for _, strict := range []bool{false, true} {
				if _, err := Compile(Option{SkelIn: path, Strict: strict}); err == nil {
					t.Fatal("accepted unsupported auth syntax")
				}
				checked, err := Check(Option{SkelIn: path, Strict: strict})
				if err != nil || !checked.Diagnostics.HasErrors() {
					t.Fatalf("%+v %v", checked, err)
				}
				for _, item := range checked.Diagnostics {
					if item.Suggestion != nil && (strings.Contains(item.Suggestion.Replacement, "auth") || strings.Contains(item.Suggestion.Message, "auth")) {
						t.Fatalf("unexpected replacement suggestion: %+v", item)
					}
				}
			}
		}
	}
}

func TestCompileRejectsInvalidAuth(t *testing.T) {
	for _, declaration := range []string{
		"api service UserApiService { for ClientActor via client auth off method ping {} }",

		"web ConsoleWeb { for ClientActor via client auth required auth off }",
		"api service UserApiService { for ClientActor via client auth optional auth required }",
		"api service UserApiService { for ClientActor via client auth unknown }",
		"web ConsoleWeb { for ClientActor via client auth unknown }",
	} {
		t.Run(declaration, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "domain.skel")
			writeFile(t, path, "domain demo.user\nactor ClientActor { via client {} }\n"+declaration)
			_, err := Compile(Option{SkelIn: path})
			if err == nil {
				t.Fatal("accepted invalid auth")
			}
			if strings.Contains(declaration, "service") && strings.Contains(declaration, "off") && !strings.Contains(err.Error(), "only supported on web") {
				t.Fatal(err)
			}
		})
	}
}

func TestCompileAuthCommentsAndActorContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeFile(t, path, `domain demo.user
actor ClientActor {
 via client {}
 auth { credential { token: string } info { userId: string } }
}
api service UserApiService {
 for ClientActor via client
 auth /* service policy */ required
 method ping { auth /* method policy */ anonymous }
}
`)
	result, err := Compile(Option{SkelIn: path, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 || result.Domain.Actors()[0].Auth == nil {
		t.Fatalf("actor auth deprecated: %+v", result)
	}
}

func TestCompileNormalizesAuthModes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeFile(t, path, `domain demo.auth
actor ClientActor { via client {} }
api service SessionApiService {
 for ClientActor via client
 auth optional
 method omitted {}
 method protected { auth required }
 method public { auth optional }
}
pub service BackendService { method call {} }
ext service ExtensionService { method call {} }
web ConsoleWeb { for ClientActor via client auth off }
`)
	result, err := Compile(Option{SkelIn: path})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatal(result.Diagnostics)
	}
	for _, declaration := range result.Domain.Declarations() {
		if declaration.Web != nil && declaration.Web.NormalizedAuth() != schema.AuthModeOff {
			t.Fatal(declaration.Web.NormalizedAuth())
		}
		if declaration.Service == nil {
			continue
		}
		want := schema.AuthModeRequired
		if declaration.Service.Api {
			want = schema.AuthModeOptional
		}
		if declaration.Service.NormalizedAuth() != want {
			t.Fatalf("%s: %s", declaration.Name, declaration.Service.NormalizedAuth())
		}
		for _, method := range declaration.Service.Methods {
			want := schema.AuthModeInherit
			switch method.Name {
			case "protected":
				want = schema.AuthModeRequired
			case "public":
				want = schema.AuthModeOptional
			}
			if method.NormalizedAuth() != want {
				t.Fatalf("%s/%s: %s", declaration.Name, method.Name, method.NormalizedAuth())
			}
		}
	}
}
