package features

import (
	"slices"
	"testing"

	lsource "go.yorun.ai/skel/internal/lsp/source"
)

func TestAuthModeCompletions(t *testing.T) {
	for _, test := range []struct {
		source string
		want   []string
	}{
		{"api service UserApiService {\n auth ", []string{"required", "optional", "anonymous"}},
		{"pub service UserService { method ping {\n auth op", []string{"required", "optional", "anonymous"}},
		{"web ConsoleWeb {\n // service IgnoredService {\n auth ", []string{"required", "optional", "anonymous", "off"}},
		{"actor ClientActor {\n auth ", nil},
	} {
		buffer := lsource.New(test.source)
		got := completionValuesBeforePositionBuffer(buffer, buffer.Range(len(test.source), len(test.source)).Start)
		if !slices.Equal(got, test.want) {
			t.Fatalf("%q: %v, want %v", test.source, got, test.want)
		}
	}
}
