package compiler

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/schema"
)

func TestCompileAuthModes(t *testing.T) {
	for _, mode := range []string{"required", "optional", "guest", "auth", "noauth", ""} {
		t.Run(mode, func(t *testing.T) {
			marker := mode
			if mode != "auth" && mode != "noauth" && mode != "" {
				marker = "auth " + mode
			}
			path := filepath.Join(t.TempDir(), "domain.skel")
			writeFile(t, path, "domain demo.user\nactor ClientActor { via client {} }\napi service UserApiService { for ClientActor via client "+marker+" method ping { auth guest } }\n")
			result, err := Compile(Option{SkelIn: path})
			if err != nil {
				t.Fatal(err)
			}
			wantCode := ""
			if mode == "auth" || mode == "noauth" {
				wantCode = diagnostic.CodeAuthLegacy
			}
			if mode == "" {
				wantCode = diagnostic.CodeApiAuthMissing
			}
			if wantCode == "" && len(result.Diagnostics) != 0 {
				t.Fatal(result.Diagnostics)
			}
			if wantCode != "" && (len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != wantCode || result.Diagnostics[0].Severity != DiagnosticSeverityWarning) {
				t.Fatal(result.Diagnostics)
			}
			document, err := schema.Project(result.Domain, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range document.Declarations {
				if d.Service == nil {
					continue
				}
				expected := mode
				if expected == "" {
					expected = "required"
				}
				if string(d.Service.Auth) != expected || d.Service.Methods[0].Auth != schema.AuthModeGuest {
					t.Fatalf("unexpected projected auth: %+v", d.Service)
				}
			}
			_, err = Compile(Option{SkelIn: path, Strict: true})
			if (err != nil) != (wantCode != "") {
				t.Fatalf("strict %s: %v", mode, err)
			}
		})
	}
}

func TestCompileWebAuthModes(t *testing.T) {
	for _, mode := range []string{"required", "optional", "guest", "off", "auth", "noauth", ""} {
		t.Run(mode, func(t *testing.T) {
			marker := mode
			if mode != "auth" && mode != "noauth" && mode != "" {
				marker = "auth " + mode
			}
			path := filepath.Join(t.TempDir(), "domain.skel")
			writeFile(t, path, "domain demo.user\nactor ClientActor { via client {} }\nweb ConsoleWeb { for ClientActor via client "+marker+" mount /console }\n")
			result, err := Compile(Option{SkelIn: path})
			if err != nil {
				t.Fatal(err)
			}
			document, err := schema.Project(result.Domain, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range document.Declarations {
				if d.Web != nil && string(d.Web.Auth) != mode {
					t.Fatalf("web auth = %q, want %q", d.Web.Auth, mode)
				}
			}
			_, err = Compile(Option{SkelIn: path, Strict: true})
			if (err != nil) != (mode == "auth" || mode == "noauth") {
				t.Fatalf("strict: %v", err)
			}
		})
	}
}

func TestCompileRejectsInvalidAuth(t *testing.T) {
	for _, declaration := range []string{
		"api service UserApiService { for ClientActor via client auth off method ping {} }",
		"pub service UserService { auth off method ping { auth required } }",
		"pub service UserService { method ping { auth off } }",
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

func TestCompileLegacyServiceAuthKeepsClientAPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeFile(t, path, `domain demo.user
pub service LegacyService { noauth method ping { auth } }
`)
	result, err := Compile(Option{SkelIn: path})
	if err != nil {
		t.Fatal(err)
	}
	service := result.Domain.Services()[0]
	if !service.ClientApi() || service.Auth != model.AuthModeNoAuth || service.Methods[0].Auth != model.AuthModeAuth {
		t.Fatalf("lost legacy rules: %+v", service)
	}
	if len(result.Diagnostics) != 3 {
		t.Fatal(result.Diagnostics)
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
 method ping { auth /* method policy */ guest }
}
`)
	result, err := Compile(Option{SkelIn: path, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 || !result.Domain.Actors()[0].AuthEnabled {
		t.Fatalf("actor auth deprecated: %+v", result)
	}
}
