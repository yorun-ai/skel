package golang_test

import (
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/testutil"
)

func TestGeneratedGoArgumentNames(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "domain.skel")
	var fields strings.Builder
	for keyword := token.BREAK; keyword <= token.VAR; keyword++ {
		if keyword.IsKeyword() {
			fields.WriteString(keyword.String() + ": string\n")
		}
	}
	for _, name := range []string{"recover", "ex", "skeltype", "service", "runner", "launcher", "ret", "err", "client", "ctx", "string", "typeName"} {
		fields.WriteString(name + ": string\n")
	}
	writeFileForTest(t, input, `domain demo.names
pub data Item { type: string typeName: string }
pub actor ClientActor { via client {} }
api service NamesApiService {
 for ClientActor via client
 auth required
 method exchange {
  input { `+fields.String()+` }
  output Item
 }
}
pub service NamesService {
 method exchange {
  input { `+fields.String()+` }
  output Item
 }
}
task NamesTask {
 trigger run { input { `+fields.String()+` } }
}
`)
	for _, mode := range []string{"backend", "api", "public", "split"} {
		t.Run(mode, func(t *testing.T) {
			out := filepath.Join(root, mode)
			option := api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, AsModule: true, Module: "example.com/names", ApiOnly: mode == "api", PubOnly: mode == "public"}
			dataOut := out
			if mode == "split" {
				option.PubOut = filepath.Join(root, "splitpub")
				option.PubModule = "example.com/namespub"
				dataOut = option.PubOut
			}
			if _, err := api.CompileGolang(api.Input{SkelIn: input}, option); err != nil {
				t.Fatal(err)
			}
			data := readFileForTest(t, filepath.Join(dataOut, "data.go"))
			codegentest.AssertGoSourceContains(t, data, "Type string `json:\"type\"`")
			codegentest.AssertGoSourceContains(t, data, "TypeName string `json:\"typeName\"`")
			filename := "service.go"
			serviceOut := out
			if mode == "public" {
				serviceOut = dataOut
			}
			service := readFileForTest(t, filepath.Join(serviceOut, filename))
			for keyword := token.BREAK; keyword <= token.VAR; keyword++ {
				if keyword.IsKeyword() {
					codegentest.AssertGoSourceContains(t, service, keyword.String()+"_ string")
				}
			}
			codegentest.AssertGoSourceContains(t, service, "recover_ string")
			codegentest.AssertGoSourceContains(t, service, "typeName string")
			if mode == "backend" || mode == "split" {
				task := readFileForTest(t, filepath.Join(out, "task.go"))
				codegentest.AssertGoSourceContains(t, task, "launcher_ *")
				codegentest.AssertGoSourceContains(t, task, "runner_ *")
				codegentest.AssertGoSourceContains(t, task, "err_ ex.Error")
			}
			// Backend compilation is a consumer check; CI stays independent of Vine.
			if mode == "api" {
				t.Run("compile", func(t *testing.T) {
					testutil.RequireToolchain(t)
					testutil.UseLocalSkel(t, out)
					testutil.Go(t, out, "test", "-mod=mod", "./...")
				})
			}
		})
	}
}
