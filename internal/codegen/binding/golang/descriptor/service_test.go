package descriptor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/schema"
)

func TestExtensionDescriptorsPreserveDeclaredVisibility(t *testing.T) {
	domain := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name:     "demo.storage",
		Services: []*schema.Service{{Name: "StorageService", Ext: true}},
		Events:   []*schema.Data{{Name: "StoredEvent", Ext: true}},
	})
	for _, mode := range []view.Mode{view.ModeFull, view.ModePub, view.ModeRegular} {
		t.Run(string(mode), func(t *testing.T) {
			out := t.TempDir()
			gen := newGen(Option{Domain: domain, View: mustView(t, mode, domain), Mode: mode, PackageName: "storage", Out: out})
			schema := gen.buildDomainDescriptor()
			if len(schema.Services) != 1 || !schema.Services[0].Ext || schema.Services[0].Pub || schema.Services[0].Api {
				t.Fatalf("extension direction lost: %+v", schema.Services)
			}
			if len(schema.Events) != 1 || !schema.Events[0].Ext || schema.Events[0].Pub {
				t.Fatalf("event declaration modifiers lost: %+v", schema.Events)
			}
			if err := gen.gen(); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join(out, descriptorGoFilename))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(content), "Ext: true") != 2 || strings.Count(string(content), "Pub: false") != 2 || strings.Contains(string(content), "Pub: true") {
				t.Fatalf("generated descriptors lost declaration modifiers: %s", content)
			}
		})
	}
}
