package api_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/diagnostic"
)

func TestImportNameRulesAcrossAnalysisModes(t *testing.T) {
	for _, test := range []struct {
		name, domain string
		imports      []string
		conflict     string
	}{
		{"current domain", "app", []string{"first as app"}, "current domain app"},
		{"other original", "app", []string{"first as a", "second as first"}, "imported domain first"},
		{"other original reversed", "app", []string{"second as first", "first as a"}, "imported domain first"},
		{"unaliased original", "app", []string{"first", "second as first"}, "imported domain first"},
		{"unaliased original reversed", "app", []string{"second as first", "first"}, "imported domain first"},
		{"duplicate qualifier", "app", []string{"first as shared", "second as shared"}, "duplicated import alias shared"},
		{"duplicate qualifier reversed", "app", []string{"second as shared", "first as shared"}, "duplicated import alias shared"},
		{"current domain suffix", "shop.order", []string{"first as order"}, ""},
		{"imported domain suffix", "app", []string{"identity.user as identity", "second as user"}, ""},
		{"same suffix", "app", []string{"first.user", "second.user"}, ""},
		{"own original", "app", []string{"first as first"}, ""},
		{"repeated import", "app", []string{"first as shared", "first as shared"}, ""},
		{"multiple aliases for one domain", "app", []string{"first as a", "first as b"}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source.skel")
			content := "domain " + test.domain + "\n"
			sources := map[string][]byte{}
			imports := map[string]string{}
			for _, declaration := range test.imports {
				content += "import " + declaration + "\n"
				name := strings.Fields(declaration)[0]
				dependency := filepath.Join(root, name+".skel")
				imports[name] = dependency
				sources[dependency] = []byte("domain " + name + "\n")
			}
			sources[path] = []byte(content)
			checked, err := api.Check(api.CheckOption{SkelIn: path, Sources: sources})
			if err != nil {
				t.Fatal(err)
			}
			if checked.Valid != (test.conflict == "") {
				t.Fatalf("check: %+v", checked)
			}
			if test.conflict != "" {
				if len(checked.Diagnostics) != 1 {
					t.Fatalf("expected one name conflict: %+v", checked.Diagnostics)
				}
				d := checked.Diagnostics[0]
				if d.Code != diagnostic.CodeSemanticDuplicate || !strings.Contains(d.Message, test.conflict) || d.Position.File != path || len(d.Related) != 1 {
					t.Fatalf("missing structured name conflict: %+v", d)
				}
			}
			input := api.Input{SkelIn: path, Sources: sources}
			_, queryErr := api.QuerySchema(input, api.SchemaQueryOption{})
			input.SkelImports = imports
			_, parseErr := api.Parse(input)
			for mode, err := range map[string]error{"query": queryErr, "parse": parseErr} {
				if test.conflict == "" {
					if err != nil {
						t.Fatalf("%s rejected valid imports: %v", mode, err)
					}
				} else if err == nil || !strings.Contains(err.Error(), test.conflict) {
					t.Fatalf("%s did not reject name conflict: %v", mode, err)
				}
			}
		})
	}
}

func TestCheckImportNamesAcrossFiles(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		root := t.TempDir()
		originalPath, aliasPath := filepath.Join(root, "a.skel"), filepath.Join(root, "b.skel")
		if reverse {
			originalPath, aliasPath = aliasPath, originalPath
		}
		sources := map[string][]byte{
			filepath.Join(root, "domain.skel"): []byte("domain app\n"),
			originalPath:                       []byte("domain app\nimport first as a\n"),
			aliasPath:                          []byte("domain app\nimport second as first\n"),
		}
		result, err := api.Check(api.CheckOption{SkelIn: root, Sources: sources})
		if err != nil || result.Valid || len(result.Diagnostics) != 1 {
			t.Fatalf("cross-file conflict: %+v, %v", result, err)
		}
		d := result.Diagnostics[0]
		if d.Position.File != aliasPath || d.Position.Line != 2 || d.Position.Column != 18 || len(d.Related) != 1 || d.Related[0].Range.Start.File != originalPath {
			t.Fatalf("wrong conflict locations: %+v", d)
		}
		// Repeating one import across files is still permitted.
		sources[aliasPath] = sources[originalPath]
		result, err = api.Check(api.CheckOption{SkelIn: root, Sources: sources})
		if err != nil || !result.Valid {
			t.Fatalf("repeated import rejected: %+v, %v", result, err)
		}
	}
}
