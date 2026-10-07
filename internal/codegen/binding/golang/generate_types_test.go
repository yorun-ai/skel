package golang_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/testutil"
)

func TestGeneratedScalarTypes(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "domain.skel")
	writeFileForTest(t, input, `domain demo.scalar
pub actor ClientActor { via client {} }
@sensitive
pub data Values {
    amount: decimal
    bytes: binary
    at: timestamp
    elapsed: duration
    date: localdate
    clock: localtime
    local: localdatetime
    id: uuid
    document: json
    optional: timestamp?
    lookup: map<uuid, list<decimal>>
}
api service ScalarApiService {
    for ClientActor via client
    auth required
    method exchange {
        input { types: Values skeltype: Values }
        output uuid
    }
}
`)
	for _, mode := range []string{"backend", "api", "public", "split"} {
		t.Run(mode, func(t *testing.T) {
			out := filepath.Join(root, mode)
			option := api.GolangOption{
				CompilerVersion: "v0.0.0-dev", Out: out, AsModule: true,
				Module: "example.com/scalar", ApiOnly: mode == "api", PubOnly: mode == "public",
			}
			dataOut := out
			if mode == "split" {
				option.PubOut = filepath.Join(root, "splitpub")
				option.PubModule = "example.com/scalarpub"
				dataOut = option.PubOut
			}
			if _, err := api.CompileGolang(api.Input{SkelIn: input, Strict: true}, option); err != nil {
				t.Fatal(err)
			}
			data := readFileForTest(t, filepath.Join(dataOut, "data.go"))
			if !strings.Contains(data, `skeltype "go.yorun.ai/skel/types"`) {
				t.Fatalf("data does not import shared types:\n%s", data)
			}
			if strings.Contains(data, `"go.yorun.ai/vrpc/skel"`) || strings.Contains(data, `"go.yorun.ai/vine/core/skel"`) {
				t.Fatalf("data still imports runtime scalar types:\n%s", data)
			}
			for _, field := range []string{
				"Amount skeltype.Decimal", "Bytes skeltype.Binary", "At skeltype.Timestamp",
				"Elapsed skeltype.Duration", "Date skeltype.LocalDate", "Clock skeltype.LocalTime",
				"Local skeltype.LocalDateTime", "Id skeltype.UUID", "Document skeltype.JSON",
				"Optional *skeltype.Timestamp", "Lookup map[skeltype.UUID][]skeltype.Decimal",
			} {
				codegentest.AssertGoSourceContains(t, data, field)
			}
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
