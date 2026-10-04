package vineschema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
)

func TestExtensionServiceSchemaPreservesDirection(t *testing.T) {
	domain := buildModelDomainForTest(t, model.DomainSpec{
		Name:     "demo.storage",
		Services: []*model.Service{{Name: "StorageService", Ext: true}},
	})
	for _, mode := range []view.Mode{view.ModeFull, view.ModePub, view.ModeRegular} {
		t.Run(string(mode), func(t *testing.T) {
			out := t.TempDir()
			gen := newGen(Option{Domain: domain, View: mustView(t, mode, domain), Mode: mode, PackageName: "storage", Out: out})
			schema := mustBuildDomainSchema(t, gen)
			if len(schema.Services) != 1 || !schema.Services[0].Ext || !schema.Services[0].Pub || schema.Services[0].Api {
				t.Fatalf("extension direction lost: %+v", schema.Services)
			}
			if err := gen.gen(); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join(out, schemaGoFilename))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(content), "Ext: true") {
				t.Fatalf("generated runtime schema lost extension flag: %s", content)
			}
		})
	}
}
