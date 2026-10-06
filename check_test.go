package skelc_test

import (
	"path/filepath"
	"testing"

	"go.yorun.ai/skelc"
)

func TestCheckFrozenInputsReturnsDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	option := skelc.CheckOption{SkelIn: path, Sources: map[string][]byte{path: []byte("domain demo\nimport absent\npub data Value { value: absent.Value }\n")}}
	result, err := skelc.Check(option)
	if err != nil || !result.Valid {
		t.Fatalf("unresolved import rejected: %+v, %v", result, err)
	}
	option.Sources[path] = []byte("domain demo\npub data Value { value: Unknown }\n")
	result, err = skelc.CheckContext(t.Context(), option)
	if err != nil || result.Valid || len(result.Diagnostics) == 0 {
		t.Fatalf("invalid source: %+v, %v", result, err)
	}
	if result.Diagnostics[0].Position.File != path {
		t.Fatalf("lost source position: %+v", result.Diagnostics)
	}
	option.Sources[path] = []byte("domain demo\nservice LegacyService { method ping {} }\n")
	option.Strict = true
	result, err = skelc.Check(option)
	if err != nil || result.Valid {
		t.Fatalf("strict mode ignored: %+v, %v", result, err)
	}
}
