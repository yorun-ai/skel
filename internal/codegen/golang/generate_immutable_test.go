package golang_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/codegen/golang"
	"go.yorun.ai/skel/internal/codegen/typescript"
	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/model"
)

func TestGenerationReusesSemanticModelAcrossTargetsAndGoroutines(t *testing.T) {
	root := t.TempDir()
	shared, consumer := filepath.Join(root, "shared.skel"), filepath.Join(root, "consumer.skel")
	writeFileForTest(t, shared, "domain shared.user\npub enum State { READY }\npub data Value { state: State }\n")
	writeFileForTest(t, consumer, "domain consumer\nimport shared.user\npub data Box<TItem> { value: TItem }\nactor TestActor { via client {} }\napi service ReadApiService { for TestActor via client method get { output Box<shared.user.Value> } }\n")
	compiled, err := compiler.Compile(compiler.Option{SkelIn: consumer, SkelImports: map[string]string{"shared.user": shared}})
	if err != nil {
		t.Fatal(err)
	}
	domain := compiled.Domain
	before := map[*model.Type]model.Type{}
	common.VisitTypes(common.ApiTypeRoots(domain.Data(), domain.Services()), func(kind *model.Type) { before[kind] = *kind })
	renderGo := func(out string) error {
		return generateFixture(domain, golang.Option{Out: out, ApiOnly: true, AsModule: true, Module: "example.com/consumerapi", ModulePrefix: "example.com"})
	}
	renderTS := func(out string) error {
		return typescript.Generate(domain, typescript.Option{Out: out, AsModule: true, ModuleScope: "@example"})
	}
	first, ts, second := filepath.Join(root, "first"), filepath.Join(root, "ts"), filepath.Join(root, "second")
	if err := renderGo(first); err != nil {
		t.Fatal(err)
	}
	if err := renderTS(ts); err != nil {
		t.Fatal(err)
	}
	if err := renderGo(second); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(generatedFiles(t, first), generatedFiles(t, second)) {
		t.Fatal("Go output depends on previous target generation")
	}
	var group sync.WaitGroup
	failures := make(chan error, 4)
	for i := range 4 {
		out := filepath.Join(root, string(rune('a'+i)))
		group.Go(func() {
			if i%2 == 0 {
				failures <- renderGo(out)
			} else {
				failures <- renderTS(out)
			}
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	for kind, original := range before {
		if !reflect.DeepEqual(original, *kind) {
			t.Fatalf("generation mutated semantic type %s", kind.SkelName)
		}
	}
}

func generatedFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		result[relative] = string(content)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
