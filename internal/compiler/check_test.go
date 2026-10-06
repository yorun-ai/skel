package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/internal/loader"
)

func TestCheckReusesSingleSyntaxSnapshotPerSource(t *testing.T) {
	directory := t.TempDir()
	files := map[string]string{
		"domain.skel": "domain test.check\n",
		"data.skel":   "domain test.check\npub data User { id: string }\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	workspaceAnalyzer := NewWorkspaceAnalyzer()
	result, err := checkWithAnalyzer(Option{SkelIn: directory}, workspaceAnalyzer)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
	}
	stats := workspaceAnalyzer.Stats()
	if stats.ParsedSources != 0 || stats.ReusedSources != len(files) {
		t.Fatalf("check did not reuse prepared syntax snapshots: %+v", stats)
	}
}

func TestCheckAndCompileShareDirectoryStructureRules(t *testing.T) {
	tests := []struct {
		name             string
		domain           string
		extra            string
		extraName        string
		expectedCode     string
		expectedFragment string
	}{
		{
			name: "mismatched domain", domain: "domain demo\n", extraName: "data.skel",
			extra: "domain other\ndata User {}\n", expectedCode: DiagnosticCodeDomainMismatch, expectedFragment: "found=other, expected=demo",
		},
		{
			name: "decorator outside domain file", domain: "domain demo\n", extraName: "data.skel",
			extra: "@desc(\"wrong file\")\ndomain demo\ndata User {}\n", expectedCode: DiagnosticCodeDomainDecorator, expectedFragment: "domain decorator is only allowed",
		},
		{
			name: "entries in domain file", domain: "domain demo\ndata User {}\n",
			expectedCode: DiagnosticCodeDomainFileContent, expectedFragment: "can only contain domain declaration",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			writeFile(t, filepath.Join(directory, loader.DomainFileName), test.domain)
			if test.extraName != "" {
				writeFile(t, filepath.Join(directory, test.extraName), test.extra)
			}

			_, compileErr := Compile(Option{SkelIn: directory})
			require.Error(t, compileErr)
			checked, checkErr := Check(Option{SkelIn: directory})
			require.NoError(t, checkErr)
			require.Len(t, checked.Diagnostics, 1)
			assert.Equal(t, test.expectedCode, checked.Diagnostics[0].Code)
			assert.Contains(t, compileErr.Error(), test.expectedFragment)
			assert.Contains(t, checked.Diagnostics[0].Message, test.expectedFragment)
		})
	}
}

func TestCheckDoesNotCascadeDomainMismatchWhenDomainFileIsInvalid(t *testing.T) {
	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, loader.DomainFileName), "domain {\n")
	writeFile(t, filepath.Join(directory, "data.skel"), "domain demo\ndata User {}\n")

	result, err := Check(Option{SkelIn: directory})
	require.NoError(t, err)
	require.NotEmpty(t, result.Diagnostics)
	for _, item := range result.Diagnostics {
		assert.NotEqual(t, DiagnosticCodeDomainMismatch, item.Code)
	}
}

func TestCheckWebMountDiagnostics(t *testing.T) {
	for _, test := range []struct {
		body, message string
		line, column  int
	}{
		{"mount /a\n    mount /b", "at most once", 5, 5},
		{"mount /a/../b", "path segments", 4, 11},
		{"mount /a//b", "empty path segments", 4, 11},
		{"mount /a?q=x", "literal path", 4, 11},
		{"mount /a#anchor", "literal path", 4, 11},
		{"mount /:name", "literal path", 4, 11},
		{"mount /a/*", "literal path", 4, 11},
		{"mount /a%2Fb", "literal path", 4, 11},
		{"mount relative", "unexpected", 4, 11},
		{`mount "/a"`, "unexpected", 4, 11},
	} {
		t.Run(test.body, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mount.skel")
			writeFile(t, path, "domain demo\nactor ClientActor { via client {} }\nweb PortalWeb {\n    "+test.body+"\n    for ClientActor\n    auth required\n}\ndata Later { id: string }\n")
			result, err := Check(Option{SkelIn: path, Strict: true})
			require.NoError(t, err)
			require.Len(t, result.Diagnostics, 1)
			diagnostic := result.Diagnostics[0]
			assert.Contains(t, diagnostic.Message, test.message)
			assert.Equal(t, path, diagnostic.Position.File)
			assert.Equal(t, test.line, diagnostic.Position.Line)
			assert.Equal(t, test.column, diagnostic.Position.Column)
		})
	}
}

func TestCompileConfigValueTypes(t *testing.T) {
	for _, test := range []struct {
		name       string
		source     string
		diagnostic string
	}{
		{"nested values", `data Entry<TValue> { value: TValue }
 data Node { content: binary children: list<Node> }
 config AppConfig instant {
  root: Node?
  entries: list<map<string, Entry<binary?>?>>
  content: binary
 }`, ""},
		{"nested config", `config ChildConfig eternal {}
 config AppConfig instant { child: list<ChildConfig> }`, "config ChildConfig cannot be used as a value type"},
		{"generic config argument", `data Entry<TValue> { value: TValue }
 config ChildConfig eternal {}
 config AppConfig instant { child: Entry<ChildConfig> }`, "config ChildConfig cannot be used as a value type"},
		{"nested event", `event ChangedEvent { payload { content: binary } }
 data Entry { change: ChangedEvent }
 config AppConfig instant { entry: Entry }`, "event ChangedEvent cannot be used as a value type"},
		{"binary map key", `config AppConfig instant { values: map<binary, string> }`, "incorrect key type"},
		{"hard cycle", `data Node { child: Node }
 config AppConfig instant { root: Node }`, "hard reference chain detected"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.skel")
			writeFile(t, path, "domain demo\n"+test.source)
			_, err := Compile(Option{SkelIn: path})
			if test.diagnostic == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.diagnostic)
				require.ErrorContains(t, err, path)
			}
		})
	}
}

func TestCompileConfigImportedData(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared.skel")
	writeFile(t, shared, `domain shared
 pub data Entry<TValue> { value: TValue }
 pub data Record { content: binary }
`)
	input := filepath.Join(dir, "config.skel")
	writeFile(t, input, `domain demo
 import shared
 pub config AppConfig eternal {
  record: shared.Record
  entries: list<shared.Entry<binary?>>
 }
`)
	result, err := Compile(Option{SkelIn: input, SkelImports: map[string]string{"shared": shared}})
	require.NoError(t, err)
	require.Len(t, result.Domain.Configs(), 1)
	require.True(t, result.Domain.Configs()[0].Members[0].Type.ContainsBinaryType())
}

func TestCompileRejectsFinalDeclarationsAsValueTypes(t *testing.T) {
	for _, declaration := range []struct {
		name       string
		source     string
		diagnostic string
	}{
		{"config", "pub config SettingsConfig eternal {}", "config SettingsConfig cannot be used as a value type"},
		{"event", "pub event ChangedEvent { payload {} }", "event ChangedEvent cannot be used as a value type"},
	} {
		name := "SettingsConfig"
		if declaration.name == "event" {
			name = "ChangedEvent"
		}
		for _, expression := range []string{name, name + "?", "list<" + name + ">", "map<string, " + name + ">", "Entry<" + name + ">"} {
			for _, owner := range []string{
				"data Wrapper { value: " + expression + " }",
				"config AppConfig instant { value: " + expression + " }",
				"event WrapperEvent { payload { value: " + expression + " } }",
				"service SampleService { method get { output " + expression + " } }",
				"service SampleService { method put { input { value: " + expression + " } } }",
			} {
				t.Run(declaration.name+"/"+owner, func(t *testing.T) {
					path := filepath.Join(t.TempDir(), "contract.skel")
					writeFile(t, path, "domain demo\ndata Entry<TValue> { value: TValue }\n"+declaration.source+"\n"+owner)
					_, err := Compile(Option{SkelIn: path})
					require.ErrorContains(t, err, declaration.diagnostic)
					require.ErrorContains(t, err, path)
				})
			}
		}
		t.Run("imported/"+declaration.name, func(t *testing.T) {
			dir := t.TempDir()
			shared := filepath.Join(dir, "shared.skel")
			writeFile(t, shared, "domain shared\n"+declaration.source)
			path := filepath.Join(dir, "contract.skel")
			writeFile(t, path, "domain demo\nimport shared\ndata Wrapper { value: shared."+name+" }")
			_, err := Compile(Option{SkelIn: path, SkelImports: map[string]string{"shared": shared}})
			require.ErrorContains(t, err, declaration.diagnostic)
			require.ErrorContains(t, err, path)
		})
	}
}

func TestCheckLoadedUsesOriginalRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	require.NoError(t, os.WriteFile(path, []byte("domain demo\n"), 0o600))
	loaded, err := loader.LoadFrom(t.Context(), loader.FileSystem{}, path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("invalid replacement"), 0o600))
	result, err := CheckLoaded(t.Context(), loaded, Option{SkelIn: path, Strict: true})
	require.NoError(t, err)
	require.Empty(t, result.Diagnostics)
}
