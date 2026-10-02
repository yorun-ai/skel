package parser

import "testing"

func TestApiServiceModifier(t *testing.T) {
	content, err := ParseSource("api.skel", []byte("domain demo.order\n// entry\napi service OrderApiService { method ping {} }\n"))
	if err != nil {
		t.Fatal(err)
	}
	service := content.Entries[0].Service
	if !service.Api || service.Pub || service.Name.Pos.Line != 3 {
		t.Fatalf("unexpected API declaration: %+v", service)
	}
	for _, declaration := range []string{"pub api service OrderApiService { method ping {} }", "api pub service OrderService { method ping {} }", "api data Order {}", "api enum Status { READY }"} {
		if _, err := ParseSource("invalid.skel", []byte("domain demo.order\n"+declaration)); err == nil {
			t.Fatalf("accepted %s", declaration)
		}
	}
}
