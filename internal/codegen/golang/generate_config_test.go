package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/golang"
	"go.yorun.ai/skelc/internal/compiler"
)

func TestGeneratedConfigTags(t *testing.T) {
	for _, lifecycle := range []string{"eternal", "instant"} {
		t.Run(lifecycle, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "config.skel")
			source := "domain demo\nconfig TextConfig " + lifecycle + ` {
    plain: string
    raw: string
    optional: string?
    items: list<string?>?
    @sensitive
    values: map<string, string?>?
    @sensitive
    secret: string
}`
			if err := os.WriteFile(input, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			parsed, err := compiler.Compile(compiler.Option{SkelIn: input})
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, "generated")
			if err := generateFixture(parsed.Domain, golang.Option{Out: out}); err != nil {
				t.Fatal(err)
			}
			content := readFileForTest(t, filepath.Join(out, "config.go"))
			if strings.Contains(content, `yaml:"`) {
				t.Fatalf("unexpected YAML tag in:\n%s", content)
			}
			for _, tag := range []string{
				"`json:\"plain\"`", "`json:\"raw\"`", "`json:\"optional\"`",
				"`json:\"items\"`", `json:"values" skel:"sensitive"`, `json:"secret" skel:"sensitive"`,
			} {
				if !strings.Contains(content, tag) {
					t.Fatalf("missing %s in:\n%s", tag, content)
				}
			}
		})
	}
}

func TestGeneratedConfigMapEnumValues(t *testing.T) {
	for _, lifecycle := range []string{"eternal", "instant"} {
		t.Run(lifecycle, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "config.skel")
			source := "domain demo\nenum Mode { ACTIVE DISABLED }\nconfig AppConfig " + lifecycle + ` {
    modes: map<string, Mode>
    optionalModes: map<string, Mode?>?
    transitions: map<Mode, Mode>
}`
			if err := os.WriteFile(input, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			parsed, err := compiler.Compile(compiler.Option{SkelIn: input})
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, "generated")
			if err := generateFixture(parsed.Domain, golang.Option{Out: out}); err != nil {
				t.Fatal(err)
			}
			content := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "config.go"))), " ")
			for _, field := range []string{
				"Modes map[string]Mode",
				"OptionalModes *map[string]*Mode",
				"Transitions map[Mode]Mode",
			} {
				if !strings.Contains(content, field) {
					t.Fatalf("missing %s in:\n%s", field, content)
				}
			}
		})
	}
}
