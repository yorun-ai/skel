package skelc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSkelcGenSkelRequiresPub(t *testing.T) {
	dir := t.TempDir()
	skelOut := filepath.Join(t.TempDir(), "pubskel")
	writeCLIFile(t, dir+"/domain.skel", `domain demo.user`)

	result := Run([]string{"gen", "skel", "--skel-in", dir, "--skel-out", skelOut})

	assertCommandErrorMessage(t, result, "flag pub is required for gen skel")
}

func TestRunSkelcGenSkel(t *testing.T) {
	dir := t.TempDir()
	skelOut := filepath.Join(t.TempDir(), "pubskel")
	writeCLIFile(t, dir+"/domain.skel", `@desc("User domain")
domain demo.user
`)
	writeCLIFile(t, dir+"/actor.skel", `domain demo.user

pub actor ClientActor {
    via client {}
}
`)
	writeCLIFile(t, dir+"/types.skel", `domain demo.user

pub enum UserStatus {
    ACTIVE
}

pub data User {
    status: UserStatus
}

data InternalUser {
    id: int
}
`)
	writeCLIFile(t, dir+"/service.skel", `domain demo.user

pub service UserService {

    method getUser {
        output User
    }
}
`)

	result := Run([]string{"gen", "skel", "--pub", "--skel-in", dir, "--skel-out", skelOut})
	assertGenerationResult(t, result)

	typesContent, err := os.ReadFile(filepath.Join(skelOut, "types.skel"))
	if err != nil {
		t.Fatalf("read generated types.skel: %v", err)
	}
	if !strings.Contains(string(typesContent), "pub data User") {
		t.Fatalf("expected pub data in generated skel: %s", string(typesContent))
	}
	if strings.Contains(string(typesContent), "InternalUser") {
		t.Fatalf("did not expect internal data in generated skel: %s", string(typesContent))
	}

	checkResult := Run([]string{"check", "--skel-in", skelOut})
	if checkResult.ExitCode != ExitCodeSuccess {
		t.Fatalf("generated skel should check: exit=%d stderr=%q", checkResult.ExitCode, checkResult.Stderr)
	}
}
