package golang_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/testutil"
)

func TestRuntimeClientLocalNameCollisions(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	input := filepath.Join(root, "service.skel")
	source := `domain demo.names
pub service NameService {
    method get {
        input {
            client: string
            ret: string
            err: string
            retI: string
            errI: string
        }
        output string
    }
    method find {
        output string?
    }
    method ping {
        input { err: string }
    }
}
`
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "generated")
	if _, err := api.CompileGolang(api.Input{SkelIn: input}, api.GolangOption{
		CompilerVersion: "v0.0.0-dev",
		PubOnly:         true,
		AsModule:        true,
		Module:          "example.com/names",
		Out:             output,
	}); err != nil {
		t.Fatal(err)
	}
	testutil.Go(t, output, "build", "-mod=mod", "./...")
}
