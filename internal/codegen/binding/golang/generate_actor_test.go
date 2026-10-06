package golang_test

import (
	"go/ast"
	stdparser "go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/api"
)

func TestOptionalActorCredentialGeneration(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeFileForTest(t, filepath.Join(source, "domain.skel"), "domain demo.auth")
	writeFileForTest(t, filepath.Join(source, "actor.skel"), `domain demo.auth
pub actor UserActor {
    via client {}
    auth {
        credential {
            token: string
            session: string?
        }
        info { userId: string }
    }
}
`)
	input := api.Input{SkelIn: source}
	goOut, tsOut, skelOut := filepath.Join(root, "golang"), filepath.Join(root, "ts"), filepath.Join(root, "skel")
	if _, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: goOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, Out: tsOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := api.CompileSkeleton(input, api.SkeletonOption{Out: skelOut, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	goFile, err := stdparser.ParseFile(token.NewFileSet(), filepath.Join(goOut, "actor.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	ast.Inspect(goFile, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "UserActorCredential" {
			return true
		}
		fields := spec.Type.(*ast.StructType).Fields.List
		for _, field := range fields {
			if len(field.Names) == 0 {
				continue
			}
			if field.Names[0].Name == "Session" {
				pointer, ok := field.Type.(*ast.StarExpr)
				if !ok {
					t.Fatal("optional credential must generate a Go pointer")
				}
				if name, ok := pointer.X.(*ast.Ident); !ok || name.Name != "string" {
					t.Fatal("optional credential must point to string")
				}
				found = true
			}
		}
		return false
	})
	if !found {
		t.Fatal("missing generated optional credential")
	}
	roundTrip, err := api.Parse(api.Input{SkelIn: skelOut})
	if err != nil {
		t.Fatal(err)
	}
	credential := roundTrip.Domain.Actors()[0].AuthCredential
	if credential.Members[0].Type.Nullable || !credential.Members[1].Type.Nullable {
		t.Fatal("public Skel did not preserve required and optional credential fields")
	}
}
