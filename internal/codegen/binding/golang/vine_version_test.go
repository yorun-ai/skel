package golang

import "testing"

func TestResolveVineVersionRejectsInvalidVersion(t *testing.T) {
	for _, version := range []string{"v0.31.99", "v0.32.0-rc.1", "0.32.0", "v-invalid", "v01.15.7", "v1.2", "v2.0.0"} {
		t.Run(version, func(t *testing.T) {
			if _, err := resolveVineVersion(version); err == nil {
				t.Fatalf("expected %q to return an error", version)
			}
		})
	}
}
