package command

import (
	"encoding/json"
	"testing"

	"go.yorun.ai/skel/diagnostic"
)

func TestErrorReportsMessageAndStableJSON(t *testing.T) {
	commandError := Error{Code: ErrorCodeInvalidArgument, Message: "invalid --skel-in"}
	if commandError.Error() != "invalid --skel-in" {
		t.Fatalf("unexpected error message: %q", commandError.Error())
	}

	encoded, err := json.Marshal(commandError)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"code":"INVALID_ARGUMENT","message":"invalid --skel-in"}`; got != want {
		t.Fatalf("Error JSON = %s; want %s", got, want)
	}
}

func TestResultsEncodeStableJSON(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "check",
			value: CheckResult{Valid: false, Diagnostics: []diagnostic.Diagnostic{}},
			want:  `{"valid":false,"diagnostics":[]}`,
		},
		{
			name:  "format",
			value: FormatResult{Changed: true, Files: []string{"contract.skel"}},
			want:  `{"changed":true,"files":["contract.skel"]}`,
		},
		{
			name:  "generation",
			value: GenerationResult{Generated: true},
			want:  `{"generated":true}`,
		},
		{
			name: "version",
			value: VersionResult{
				Name:      "skelc",
				Version:   "v0.21.0",
				Platform:  "darwin/arm64",
				GoVersion: "go1.27.1",
				GolangCodeGen: VersionGolangCodeGenResult{
					MinimumVineVersion: "v0.27.0",
					DefaultVineVersion: "v0.27.0",
				},
			},
			want: `{"name":"skelc","version":"v0.21.0","platform":"darwin/arm64","goVersion":"go1.27.1",` +
				`"golangCodeGen":{"minimumVineVersion":"v0.27.0","defaultVineVersion":"v0.27.0"}}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(encoded); got != test.want {
				t.Fatalf("JSON = %s; want %s", got, test.want)
			}
		})
	}
}
