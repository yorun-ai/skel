package golang

import "testing"

func TestResolveVrpcVersion(t *testing.T) {
	for _, test := range []struct{ version, want string }{
		{"", "v0.14.0"},
		{"  ", "v0.14.0"},
		{"v0.13.0", "v0.13.0"},
		{"v0.14.0", "v0.14.0"},
	} {
		got, err := resolveVrpcVersion(test.version)
		if err != nil || got != test.want {
			t.Fatalf("resolve %q: got %q, want %q, error %v", test.version, got, test.want, err)
		}
	}
	for _, version := range []string{"v0.11.0", "v0.12.0", "v0.13.0-rc.1", "v01.12.0", "v1.2", "v2.0.0"} {
		if _, err := resolveVrpcVersion(version); err == nil {
			t.Fatalf("accepted unsupported vRPC version %q", version)
		}
	}
}
