package source

import (
	goparser "go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/schema"
)

func TestGeneratedGoParsesWithMultilineDeprecatedReasons(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Services: []*schema.Service{{
			Name:             "UserService",
			Deprecated:       true,
			DeprecatedReason: "Use ProfileService instead\nComplete migration first",
			Methods: []*schema.Method{{
				Name:             "getUser",
				Deprecated:       true,
				DeprecatedReason: "Use getProfile instead\nThe old response will be removed",
			}}, Pub: true}},
		Tasks: []*schema.Task{{
			Name:             "RefreshTask",
			Deprecated:       true,
			DeprecatedReason: "Use RebuildTask instead\nThe old schedule will be removed",
			Triggers: []*schema.TaskTrigger{{
				Name:             "legacy",
				Deprecated:       true,
				DeprecatedReason: "Use manually instead\nLegacy scheduling is disabled",
			}},
		}},
	})
	outputDir := t.TempDir()
	gen := newGen(Option{
		Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull,
		PackageName: "skeled", Out: outputDir,
	})
	gen.genServiceGo()
	gen.genTaskGo()

	for _, filename := range []string{serviceGoFilename, taskGoFilename} {
		path := filepath.Join(outputDir, filename)
		file, err := goparser.ParseFile(token.NewFileSet(), path, nil, goparser.ParseComments)
		if err != nil {
			t.Fatalf("generated Go should parse: %v", err)
		}
		if file.Name.Name != "skeled" {
			t.Fatalf("expected package skeled in %s, got %s", path, file.Name.Name)
		}
	}
}
