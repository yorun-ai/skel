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
		"pub ext service StorageService", "ext pub service StorageService", "api ext service StorageApiService", "ext api service StorageApiService", "ext ext service StorageService", "ext data Item", "ext actor UserActor", "ext resource Item", "ext task ItemTask", "ext config ItemConfig", "ext enum Item", "open service StorageService",
	} {
		if _, err := ParseSource("invalid.skel", []byte("domain demo\n"+declaration+" {}")); err == nil {
			t.Fatalf("accepted %s", declaration)
		}
	}
}

func TestExtEventModifier(t *testing.T) {
	content, err := ParseSource("ext.skel", []byte("domain demo\n// extension\n@desc(\"audit\")\next event AuditRecordedEvent { payload { message: string } }\n"))
	if err != nil {
		t.Fatal(err)
	}
	event := content.Entries[0].Event
	if !event.Ext || event.Pub || event.Name.Pos.Line != 4 {
		t.Fatalf("wrong extension event: %+v", event)
	}
	for _, declaration := range []string{"pub ext event", "ext pub event", "api ext event", "ext api event", "ext ext event"} {
		if _, err := ParseSource("invalid.skel", []byte("domain demo\n"+declaration+" AuditRecordedEvent { payload {} }")); err == nil {
			t.Fatalf("accepted %s", declaration)
		}
	}
}
