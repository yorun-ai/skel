package typescript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/output"
	"go.yorun.ai/skel/internal/model"
)

func TestGenerateModule(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "ts")
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{{
			Name: "User",
			Pub:  true,
			Members: []*model.DataMember{{
				Name: "id",
				Type: &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString},
			}},
		}},
	})

	if err := Generate(domain, Option{Out: outDir, AsModule: true, Module: "@yorun-ai/skeled-demo-user"}); err != nil {
		t.Fatalf("generate typescript: %v", err)
	}

	packageJSON, err := os.ReadFile(filepath.Join(outDir, "package.json"))
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	if !strings.Contains(string(packageJSON), `"types": "./index.ts"`) {
		t.Fatalf("unexpected package.json: %s", packageJSON)
	}
	if !strings.Contains(string(packageJSON), `"@yorun-ai/vrpc": "*"`) {
		t.Fatalf("missing vrpc peer dependency: %s", packageJSON)
	}
	var packageData map[string]any
	if err := json.Unmarshal(packageJSON, &packageData); err != nil {
		t.Fatalf("generated package.json is invalid: %v", err)
	}
	if packageData["_generated"] != output.GeneratedFileMarker {
		t.Fatalf("missing generated marker in package.json: %s", packageJSON)
	}
	if _, err := os.Stat(filepath.Join(outDir, "tsconfig.json")); !os.IsNotExist(err) {
		t.Fatalf("expected tsconfig.json to be missing: %v", err)
	}
	for _, filename := range []string{"data.ts", "service.ts", "spec.ts"} {
		if _, err := os.Stat(filepath.Join(outDir, filename)); err != nil {
			t.Fatalf("expected %s: %v", filename, err)
		}
	}
}
