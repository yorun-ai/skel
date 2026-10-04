package parser

import "testing"

func TestExtServiceModifier(t *testing.T) {
	content, err := ParseSource("ext.skel", []byte("domain demo\n@desc(\"contract\")\next service StorageService { method ping {} }\n"))
	if err != nil {
		t.Fatal(err)
	}
	service := content.Entries[0].Service
	if !service.Ext || service.Pub || service.Api || service.Name.Pos.Line != 3 {
		t.Fatalf("wrong ext declaration: %+v", service)
	}
	for _, declaration := range []string{
		"pub ext service StorageService", "ext pub service StorageService", "api ext service StorageApiService", "ext api service StorageApiService", "ext ext service StorageService", "ext data Item", "ext actor UserActor", "ext resource Item", "ext event ItemEvent", "ext task ItemTask", "ext config ItemConfig", "ext enum Item", "open service StorageService",
	} {
		if _, err := ParseSource("invalid.skel", []byte("domain demo\n"+declaration+" {}")); err == nil {
			t.Fatalf("accepted %s", declaration)
		}
	}
}
