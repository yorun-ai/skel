package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc"
)

func TestApiActorFilterAcrossTargets(t *testing.T) {
	root := t.TempDir()
	actors, backend, entry := filepath.Join(root, "actors.skel"), filepath.Join(root, "backend.skel"), filepath.Join(root, "entry.skel")
	writeFileForTest(t, actors, `domain identity.user
pub actor UserActor { via client {} via agent {} }
`)
	writeFileForTest(t, backend, `domain private.backend
pub data Secret { value: string }
`)
	writeFileForTest(t, entry, `domain demo.order
import identity.user as identity
import private.backend as backend
actor AdminActor { via client {} }
actor IdleActor { via client {} }
pub data PublicValue { value: string }
data UserValue { value: string }
api service UserApiService {
 for identity.UserActor via agent
 noauth
 method get { output UserValue }
}
api service SharedApiService {
 for identity.UserActor via client
 for AdminActor via client
 method ping {}
}
api service AdminApiService {
 for AdminActor via client
 method get { output backend.Secret }
}
`)
	input := skelc.Input{SkelIn: entry, SkelImports: map[string]string{"identity.user": actors, "private.backend": backend}}
	for _, target := range []string{"go", "ts"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			generate := func(names []string) error {
				selection := skelc.ApiFilter{Actors: names}
				if target == "go" {
					_, err := skelc.CompileGolang(input, skelc.GolangOption{ApiOnly: true, ApiFilter: selection, Out: out, Imports: map[string]string{"private.backend": "example.com/backendapi"}})
					return err
				}
				_, err := skelc.CompileTypeScript(input, skelc.TypeScriptOption{ApiOnly: true, ApiFilter: selection, Out: out, Imports: map[string]string{"private.backend": "@demo/backendapi"}})
				return err
			}
			ext := "." + target
			if target == "ts" {
				ext = ".ts"
			}
			servicePath, dataPath := filepath.Join(out, "service"+ext), filepath.Join(out, "data"+ext)
			for _, names := range [][]string{nil, {"identity.user.UserActor"}, {"demo.order.AdminActor", "identity.user.UserActor"}, {"demo.order.IdleActor"}} {
				if err := generate(names); err != nil {
					t.Fatal(err)
				}
				service, err := os.ReadFile(servicePath)
				if len(names) == 1 && names[0] == "demo.order.IdleActor" {
					if target == "ts" {
						if err != nil || strings.Contains(string(service), "ApiService") {
							t.Fatalf("stale TS service: %s, %v", service, err)
						}
					} else if !os.IsNotExist(err) {
						t.Fatalf("stale service output: %s, %v", service, err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					for _, name := range []string{"UserApiService", "SharedApiService"} {
						if !strings.Contains(string(service), name) {
							t.Fatalf("missing %s: %s", name, service)
						}
					}
					wantAdmin := len(names) != 1
					if strings.Contains(string(service), "AdminApiService") != wantAdmin {
						t.Fatalf("wrong selection: %s", service)
					}
				}
				data, err := os.ReadFile(dataPath)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "PublicValue") || strings.Contains(string(data), "UserValue") != (len(names) != 1 || names[0] != "demo.order.IdleActor") {
					t.Fatalf("wrong data closure: %s", data)
				}
			}
			before := generatedFiles(t, out)
			for _, name := range []string{"UserActor", "identity.UserActor", "identity.user.MissingActor", ""} {
				if err := generate([]string{name}); err == nil {
					t.Fatalf("accepted invalid actor %q", name)
				}
				after := generatedFiles(t, out)
				if len(before) != len(after) {
					t.Fatal("failed generation changed output")
				}
				for path, content := range before {
					if after[path] != content {
						t.Fatalf("changed %s", path)
					}
				}
			}
			// The unselected admin service must not require an API import mapping.
			if target == "go" {
				_, err := skelc.CompileGolang(input, skelc.GolangOption{ApiOnly: true, ApiFilter: skelc.ApiFilter{Actors: []string{"identity.user.UserActor"}}, Out: out})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				_, err := skelc.CompileTypeScript(input, skelc.TypeScriptOption{ApiOnly: true, ApiFilter: skelc.ApiFilter{Actors: []string{"identity.user.UserActor"}}, Out: out})
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
