package skelc_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skelc"
)

func TestApiPruneQueryAndGeneration(t *testing.T) {
	root := t.TempDir()
	entry, foreign, unused, deep := filepath.Join(root, "entry.skel"), filepath.Join(root, "foreign.skel"), filepath.Join(root, "unused.skel"), filepath.Join(root, "deep.skel")
	writeTestFile(t, deep, "domain deep\npub data Detail { text: string }\n")
	writeTestFile(t, foreign, `domain foreign
import deep
pub data Box<TItem> { value: TItem }
pub data Money { amount: int detail: deep.Detail }
pub enum Currency { USD EUR }
`)
	writeTestFile(t, unused, "domain unused\npub data Secret { value: string }\n")
	writeTestFile(t, entry, `domain shop.order
import foreign as f
import unused
actor UserActor { via client {} }
actor AdminActor { via client {} }
actor IdleActor { via client {} }
pub data Unused { secret: unused.Secret }
pub enum UnusedEnum { ONE TWO }
data Node { next: Peer? value: f.Money currency: f.Currency }
data Peer { next: Node? values: map<string,list<f.Money?>> }
pub data Extra { value: string }
api service UserApiService {
 for UserActor via client
 auth anonymous
 method read { output f.Box<list<Node>> }
}
api service AdminApiService {
 for AdminActor via client
 auth anonymous
 method read { output Unused }
}
`)
	input := skelc.Input{SkelIn: entry, SkelImports: map[string]string{"foreign": foreign, "unused": unused, "deep": deep}}
	actor := []string{"shop.order.UserActor"}
	selection := skelc.ApiFilter{Prune: true, Actors: actor}
	result, err := skelc.QueryApiDependencies(input, selection)
	if err != nil {
		t.Fatal(err)
	}
	want := &skelc.ApiDependencyReport{Domain: "shop.order", Services: []string{"shop.order.UserApiService"}, Data: []string{"shop.order.Node", "shop.order.Peer"}, Enums: []string{}, Dependencies: []skelc.ApiTypeDependency{{Domain: "foreign", Name: "Box", Kind: "data"}, {Domain: "foreign", Name: "Currency", Kind: "enum"}, {Domain: "foreign", Name: "Money", Kind: "data"}}}
	if !reflect.DeepEqual(result.Report, want) {
		t.Fatalf("report=%+v, want %+v", result.Report, want)
	}
	// Foreign members belong to a separate query; aliases never appear in reports.
	next, err := skelc.QueryApiDependencies(skelc.Input{SkelIn: foreign, SkelImports: map[string]string{"deep": deep}}, skelc.ApiFilter{Prune: true, Types: []string{"foreign.Money"}})
	if err != nil || len(next.Report.Dependencies) != 1 || next.Report.Dependencies[0].Domain != "deep" {
		t.Fatalf("foreign closure: %+v, %v", next, err)
	}
	union := skelc.ApiFilter{Prune: true, Actors: actor, Types: []string{"shop.order.Extra", "shop.order.UnusedEnum", "shop.order.Extra"}}
	combined, err := skelc.QueryApiDependencies(input, union)
	if err != nil || len(combined.Report.Data) != 3 || len(combined.Report.Enums) != 1 {
		t.Fatalf("union: %+v, %v", combined, err)
	}
	empty, err := skelc.QueryApiDependencies(input, skelc.ApiFilter{Prune: true, Actors: []string{"shop.order.IdleActor"}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(empty.Report)
	if strings.Contains(string(encoded), "null") || len(empty.Report.Data) != 0 {
		t.Fatalf("empty selection: %s", encoded)
	}
	// Root order and repeated roots must not affect either the report or output.
	ordered := skelc.ApiFilter{Prune: true, Actors: []string{"shop.order.UserActor", "shop.order.AdminActor"}, Types: []string{"shop.order.Extra", "shop.order.UnusedEnum"}}
	reordered := skelc.ApiFilter{Prune: true, Actors: []string{"shop.order.AdminActor", "shop.order.UserActor", "shop.order.AdminActor"}, Types: []string{"shop.order.UnusedEnum", "shop.order.Extra", "shop.order.UnusedEnum"}}
	first, err := skelc.QueryApiDependencies(input, ordered)
	if err != nil {
		t.Fatal(err)
	}
	second, err := skelc.QueryApiDependencies(input, reordered)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Report, second.Report) {
		t.Fatalf("root order changed dependency report: %+v vs %+v", first.Report, second.Report)
	}
	for _, target := range []string{"go", "go-module", "ts", "ts-module"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			generate := func(filter skelc.ApiFilter) error {
				if strings.HasPrefix(target, "go") {
					_, err := skelc.CompileGolang(input, skelc.GolangOption{ApiOnly: true, ApiFilter: filter, Out: out, AsModule: target == "go-module", Module: moduleForTarget(target, "example.com/orderapi"), Imports: map[string]string{"foreign": "example.com/foreignapi", "unused": "example.com/unusedapi"}})
					return err
				}
				_, err := skelc.CompileTypeScript(input, skelc.TypeScriptOption{ApiOnly: true, ApiFilter: filter, Out: out, AsModule: target == "ts-module", Module: moduleForTarget(target, "@demo/orderapi"), Imports: map[string]string{"foreign": "@demo/foreignapi", "unused": "@demo/unusedapi"}})
				return err
			}
			read := func() string {
				var contents []string
				err := filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if !d.IsDir() {
						data, e := os.ReadFile(path)
						if e != nil {
							return e
						}
						contents = append(contents, string(data))
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				return strings.Join(contents, "\n")
			}
			if err := generate(skelc.ApiFilter{Actors: actor}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(read(), "Unused") {
				t.Fatal("default public types were pruned")
			}
			if err := generate(selection); err != nil {
				t.Fatal(err)
			}
			text := read()
			for _, name := range []string{"Unused", "UnusedEnum", "AdminApiService", "unusedapi", "Extra"} {
				if strings.Contains(text, name) {
					t.Fatalf("retained %s", name)
				}
			}
			if !strings.Contains(text, "Node") || !strings.Contains(text, "UserApiService") {
				t.Fatal("missing selected declarations")
			}
			if err := generate(skelc.ApiFilter{Prune: true, Types: []string{"shop.order.Extra"}}); err != nil {
				t.Fatal(err)
			}
			text = read()
			if !strings.Contains(text, "Extra") || strings.Contains(text, "UserApiService") || strings.Contains(text, "Node") || strings.Contains(text, "foreignapi") {
				t.Fatal("types-only output retained unrelated declarations or dependencies")
			}
			if err := generate(skelc.ApiFilter{Prune: true, Types: []string{"shop.order.Missing"}}); err == nil {
				t.Fatal("unknown type accepted")
			}
			if read() != text {
				t.Fatal("invalid selection changed existing output")
			}
			if err := generate(ordered); err != nil {
				t.Fatal(err)
			}
			before := apiOutputSnapshot(t, out)
			if err := generate(reordered); err != nil {
				t.Fatal(err)
			}
			if after := apiOutputSnapshot(t, out); !reflect.DeepEqual(before, after) {
				t.Fatal("root order changed generated file paths or contents")
			}

			// Enum-only output must remove previous services, data and package dependencies.
			if err := generate(skelc.ApiFilter{Prune: true, Types: []string{"shop.order.UnusedEnum"}}); err != nil {
				t.Fatal(err)
			}
			text = read()
			if !strings.Contains(text, "UnusedEnum") {
				t.Fatal("enum-only output omitted the selected enum")
			}
			for _, stale := range []string{"UserApiService", "AdminApiService", "Node", "Peer", "Extra", "foreignapi", "unusedapi"} {
				if strings.Contains(text, stale) {
					t.Fatalf("enum-only output retained %s", stale)
				}
			}

			// Export a populated view before selecting an actor with no services.
			if err := generate(ordered); err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, filepath.Join(out, "user-note.txt"), "keep handwritten file")
			idle := skelc.ApiFilter{Prune: true, Actors: []string{"shop.order.IdleActor"}}
			if err := generate(idle); err != nil {
				t.Fatal(err)
			}
			cleared := apiOutputSnapshot(t, out)
			if cleared["user-note.txt"] != "keep handwritten file" {
				t.Fatal("empty selection removed user content")
			}
			delete(cleared, "user-note.txt")
			// Equality with a fresh empty export detects stale files, barrel entries,
			// Go requirements and package.json dependencies without fixing a file layout.
			out = filepath.Join(t.TempDir(), "api")
			if err := generate(idle); err != nil {
				t.Fatal(err)
			}
			if fresh := apiOutputSnapshot(t, out); !reflect.DeepEqual(cleared, fresh) {
				t.Fatalf("empty selection retained stale output: cleared=%v fresh=%v", cleared, fresh)
			}

		})
	}
}

func moduleForTarget(target, module string) string {
	if strings.HasSuffix(target, "-module") {
		return module
	}
	return ""
}

func apiOutputSnapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}
