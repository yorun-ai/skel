package golang_test

import (
	"fmt"
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

func TestGeneratedStructuredConfig(t *testing.T) {
	for _, lifecycle := range []string{"eternal", "instant"} {
		for _, pubOnly := range []bool{false, true} {
			t.Run(lifecycle+"/pub="+fmt.Sprint(pubOnly), func(t *testing.T) {
				dir := t.TempDir()
				input := filepath.Join(dir, "config.skel")
				writeFileForTest(t, input, `domain demo
 data Entry<TValue> { value: TValue }
 data Record { name: string content: binary children: list<Record> }
 pub config AppConfig `+lifecycle+` {
  record: Record?
  records: list<Record>
  groups: map<string, list<Entry<binary?>>>
  content: binary
 }`)
				parsed, err := compiler.Compile(compiler.Option{SkelIn: input})
				if err != nil {
					t.Fatal(err)
				}
				out := filepath.Join(dir, "generated")
				if err := generateFixture(parsed.Domain, golang.Option{Out: out, PubOnly: pubOnly}); err != nil {
					t.Fatal(err)
				}
				config := strings.Join(strings.Fields(readFileForTest(t, filepath.Join(out, "config.go"))), " ")
				for _, expected := range []string{"Record *Record", "Records []Record", "Groups map[string][]Entry[*skel.Binary]", "Content skel.Binary"} {
					if !strings.Contains(config, expected) {
						t.Fatalf("missing %s in %s", expected, config)
					}
				}
				data := readFileForTest(t, filepath.Join(out, "data.go"))
				for _, expected := range []string{"type Record struct", "type Entry[TValue any] struct"} {
					if !strings.Contains(data, expected) {
						t.Fatalf("missing dependency %s in %s", expected, data)
					}
				}
			})
		}
	}
}
