package features

import (
	"strings"
	"testing"
)

func FuzzLanguageFeaturesHandleInvalidSource(f *testing.F) {
	f.Add("domain demo\n" + strings.Join(editingDeclarations[2:], "\n"))
	for _, declaration := range editingDeclarations {
		f.Add("domain demo\n" + declaration)
	}
	for _, seed := range []string{"", "domain demo\ndata User {\n @desc(\"unfinished\")\n}", "domain demo\nservice UserService { method get { require any(all("} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 4096 {
			t.Skip()
		}
		exerciseEditingSource(t, text)
	})
}
