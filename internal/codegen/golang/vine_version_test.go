package golang

import "testing"

func TestResolveVineVersion(t *testing.T) {
	for _, test := range []struct {
		name     string
		version  string
		expected string
	}{
		{name: "default", expected: DefaultVineVersion},
		{name: "trimmed default", version: "  ", expected: DefaultVineVersion},
		{name: "explicit minimum", version: "v0.25.1", expected: "v0.25.1"},
		{name: "higher", version: "v1.2.3", expected: "v1.2.3"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveVineVersion(test.version)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.expected {
				t.Fatalf("unexpected Vine version: got %q want %q", got, test.expected)
			}
		})
	}
}

func TestResolveVineVersionRejectsInvalidVersion(t *testing.T) {
	for _, version := range []string{"v0.15.6", "v0.20.2", "v0.25.0", "v0.25.1-rc.1", "0.25.1", "v-invalid", "v01.15.7", "v1.2", "v2.0.0"} {
		t.Run(version, func(t *testing.T) {
			if _, err := resolveVineVersion(version); err == nil {
				t.Fatalf("expected %q to return an error", version)
			}
		})
	}
}
