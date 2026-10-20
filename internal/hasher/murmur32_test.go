package hasher

import (
	"fmt"
	"testing"
)

func TestMurmur32(t *testing.T) {
	if got := fmt.Sprintf("%08x", murmur32([]byte("hello"))); got != "248bfa47" {
		t.Fatalf("unexpected murmur32 hash: %s", got)
	}
}
