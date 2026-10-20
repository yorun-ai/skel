package typescript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestGenRendersTypesWithoutClientServices(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "ts")
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{{
			Name: "User", Pub: true,
			Members: []*schema.DataMember{{Name: "id", Type: codegentest.IntType()}},
		}},
		Services: []*schema.Service{{
			Name: "BackendService", Pub: true, Methods: []*schema.Method{{Name: "ping"}},
		}},
	})

	gen := newTestGen(pkg, outDir)
	if gen.err != nil {
		t.Fatal(gen.err)
	}
	if len(gen.apiView.Services) != 0 {
		t.Fatalf("fixture must not expose client services: %+v", gen.apiView.Services)
	}
	gen.generate()
	if err := gen.renderer.Err(); err != nil {
		t.Fatal(err)
	}

	for _, filename := range []string{indexFilename, dataTsFilename, serviceTsFilename, specTsFilename} {
		content, err := os.ReadFile(filepath.Join(outDir, filename))
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}
		if strings.Contains(string(content), "BackendService") {
			t.Fatalf("backend service leaked into %s: %s", filename, content)
		}
		if filename == dataTsFilename && !strings.Contains(string(content), "export type User =") {
			t.Fatalf("public type missing from data.ts: %s", content)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "package.json")); !os.IsNotExist(err) {
		t.Fatalf("expected package.json to be missing, err=%v", err)
	}
}
