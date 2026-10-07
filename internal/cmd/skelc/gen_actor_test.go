package skelc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerationActorFlags(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.skel")
	writeCLIFile(t, source, `domain demo.user
actor UserActor { via client {} }
actor AdminActor { via client {} }
api service UserApiService { for UserActor via client auth required method ping {} }
api service AdminApiService { for AdminActor via client auth required method ping {} }
`)
	for _, target := range []string{"go", "go-module", "ts"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			args := []string{"gen", target, "--api", "--skel-in", source}
			filename := "service.go"
			if target == "ts" {
				args = append(args, "--ts-out", out)
				filename = "service.ts"
			} else {
				args = append(args, "--go-out", out)
			}
			if target == "go-module" {
				args = append(args, "--go-module", "example.com/userapi")
			}
			for _, names := range [][]string{{"demo.user.UserActor"}, {"demo.user.UserActor", "demo.user.AdminActor"}} {
				selected := append([]string{}, args...)
				for _, name := range names {
					selected = append(selected, "--actor", name)
				}
				assertGenerationResult(t, Run(selected))
				data, err := os.ReadFile(filepath.Join(out, filename))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "UserApiService") || strings.Contains(string(data), "AdminApiService") != (len(names) == 2) {
					t.Fatalf("wrong selection: %s", data)
				}
			}
			for _, name := range []string{"UserActor", "demo.user.MissingActor", ""} {
				result := Run(append(append([]string{}, args...), "--actor", name))
				if result.ExitCode == ExitCodeSuccess {
					t.Fatalf("accepted actor %q", name)
				}
			}
			withoutAPI := append([]string{}, args[:2]...)
			withoutAPI = append(withoutAPI, args[3:]...)
			result := Run(append(withoutAPI, "--actor", "demo.user.UserActor"))
			assertCommandErrorMessage(t, result, "flag actor requires api")
		})
	}
}
