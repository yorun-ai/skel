package output_test

import (
	"encoding/json"
	"testing"

	"go.yorun.ai/skel/cmd/skelc/output"
	"go.yorun.ai/skel/diagnostic"
)

func TestFacadeWireContract(t *testing.T) {
	if output.ExitCodeSuccess != 0 || output.ExitCodeUnsatisfied != 1 || output.ExitCodeError != 2 {
		t.Fatalf("unexpected exit codes: %d %d %d", output.ExitCodeSuccess, output.ExitCodeUnsatisfied, output.ExitCodeError)
	}
	for _, test := range []struct {
		name  string
		value any
		want  string
	}{
		{name: "error", value: output.Error{Code: output.ErrorCodeCompilationFailed, Message: "invalid input"}, want: `{"code":"COMPILATION_FAILED","message":"invalid input"}`},
		{name: "check", value: output.CheckResult{Valid: true, Diagnostics: []diagnostic.Diagnostic{}}, want: `{"valid":true,"diagnostics":[]}`},
		{name: "format", value: output.FormatResult{Changed: false, Files: []string{}}, want: `{"changed":false,"files":[]}`},
		{name: "generation", value: output.GenerationResult{Generated: true}, want: `{"generated":true}`},
		{
			name: "version",
			value: output.VersionResult{
				Name: "Skelc CLI", Version: "v0.14.0", Platform: "darwin/arm64", GoVersion: "go1.26.6",
				GolangCodeGen: output.VersionGolangCodeGenResult{MinimumVineVersion: "v0.13.1", DefaultVineVersion: "v0.13.1"},
			},
			want: `{"name":"Skelc CLI","version":"v0.14.0","platform":"darwin/arm64","goVersion":"go1.26.6","golangCodeGen":{"minimumVineVersion":"v0.13.1","defaultVineVersion":"v0.13.1"}}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != test.want {
				t.Fatalf("unexpected JSON: %s", encoded)
			}
		})
	}
}
