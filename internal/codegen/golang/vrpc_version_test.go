package golang

import "testing"

func TestResolveVrpcVersion(t *testing.T) {
	for _, version := range []string{"", "v0.13.0"} {
		got, err := resolveVrpcVersion(version)
		if err != nil || got != "v0.13.0" {
			t.Fatalf("resolve %q: %q %v", version, got, err)
		}
	}
	for _, version := range []string{"v0.11.0", "v0.12.0", "v0.13.0-rc.1", "v01.12.0", "v1.2", "v2.0.0"} {
		if _, err := resolveVrpcVersion(version); err == nil {
			t.Fatalf("accepted unsupported vRPC version %q", version)
		}
	}
}
