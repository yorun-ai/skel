package source

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestCastEventVersionedMethodNames(t *testing.T) {
	for _, test := range []struct{ name, method string }{
		{"OrderPlacedEvent", "OrderPlaced"},
		{"OrderPlacedEventV1", "OrderPlacedV1"},
		{"OrderPlacedEventV2", "OrderPlacedV2"},
		{"OrderPlacedEventV10", "OrderPlacedV10"},
		{"OrderV2PlacedEventV10", "OrderV2PlacedV10"},
		{"OrderPlacedV2Event", "OrderPlacedV2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			event := new(_Gen).castEvent(&schema.Data{Name: test.name, SkelName: "demo." + test.name}, false, false)
			if event.Name != test.name || event.SkelName != "demo."+test.name || event.EmitterName != test.name+"Emitter" || event.ListenerName != test.name+"Listener" {
				t.Fatalf("event identity changed: %+v", event)
			}
			if event.EmitterMethodName != "Emit"+test.method || event.ListenerMethodName != "On"+test.method {
				t.Fatalf("unexpected event methods: %s / %s", event.EmitterMethodName, event.ListenerMethodName)
			}
		})
	}
}
