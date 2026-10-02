package parser

import "testing"

func TestOpenServiceModifier(t *testing.T) {
	content, err := ParseSource("open.skel", []byte("domain demo\n@desc(\"contract\")\nopen service StorageService { method ping {} }\n"))
	if err != nil {
		t.Fatal(err)
	}
	service := content.Entries[0].Service
	if !service.Open || service.Pub || service.Api || service.Name.Pos.Line != 3 {
		t.Fatalf("wrong open declaration: %+v", service)
	}
	for _, declaration := range []string{
		"pub open service StorageService", "open pub service StorageService", "api open service StorageApiService", "open api service StorageApiService", "open open service StorageService", "open data Item", "open actor UserActor", "open resource Item",
	} {
		if _, err := ParseSource("invalid.skel", []byte("domain demo\n"+declaration+" {}")); err == nil {
			t.Fatalf("accepted %s", declaration)
		}
	}
}
