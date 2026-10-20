package source

import (
	"os"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/testutil"
	"go.yorun.ai/skel/schema"
)

func TestGeneratedNullableTypeParametersRoundTrip(t *testing.T) {
	parameter := codegentest.TypeParam("TValue")
	nullable := func() *schema.Type { return codegentest.NullableType(codegentest.TypeParamType(parameter)) }
	box := &schema.Data{Name: "Box", TypeParameters: []*schema.TypeParameter{parameter}, Members: []*schema.DataMember{{Name: "value", Type: codegentest.TypeParamType(parameter)}}}
	wrapper := &schema.Data{Name: "Wrapper", TypeParameters: []*schema.TypeParameter{parameter}, Members: []*schema.DataMember{
		{Name: "optional", Type: nullable()},
		{Name: "items", Type: codegentest.ListType(nullable())},
		{Name: "values", Type: codegentest.MapType(codegentest.StringType(), nullable())},
		{Name: "nested", Type: codegentest.DataType(box, nullable())},
	}}
	domain := buildSchemaDomainForTest(t, schema.DomainSpec{Name: "demo.generic", Data: []*schema.Data{box, wrapper}})
	output := t.TempDir()
	generator := newGen(Option{Domain: domain, View: mustView(t, view.ModeRegular, domain), Mode: view.ModeRegular, PackageName: "generic", Out: output})
	generator.genDataGo()
	if err := generator.Renderer.Err(); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"go.mod": "module example.com/generic\n\ngo 1.27.0\n",
		"data_test.go": `package generic
import (
 "encoding/json/v2"
 "reflect"
 "testing"
)
func checkRoundTrip[TValue any](t *testing.T, input string, zero TValue) {
 t.Helper()
 var value Wrapper[TValue]
 if err := json.Unmarshal([]byte(input), &value); err != nil { t.Fatal(err) }
 if value.Optional != nil || len(value.Items) != 2 || value.Items[0] != nil || value.Items[1] == nil || !reflect.DeepEqual(*value.Items[1], zero) { t.Fatalf("null or zero lost: %+v", value) }
 if value.Values["null"] != nil || value.Values["zero"] == nil || value.Nested.Value != nil { t.Fatal("nullable map or nested argument lost") }
 encoded, err := json.Marshal(value)
 if err != nil { t.Fatal(err) }
 var roundTripped Wrapper[TValue]
 if err := json.Unmarshal(encoded, &roundTripped); err != nil { t.Fatal(err) }
 if !reflect.DeepEqual(value, roundTripped) { t.Fatalf("roundtrip mismatch: %s", encoded) }
}
func TestGeneratedGenerics(t *testing.T) {
 checkRoundTrip(t, ` + "`" + `{"optional":null,"items":[null,""],"values":{"null":null,"zero":""},"nested":{"value":null}}` + "`" + `, "")
 checkRoundTrip(t, ` + "`" + `{"optional":null,"items":[null,0],"values":{"null":null,"zero":0},"nested":{"value":null}}` + "`" + `, 0)
 checkRoundTrip(t, ` + "`" + `{"optional":null,"items":[null,false],"values":{"null":null,"zero":false},"nested":{"value":null}}` + "`" + `, false)
 checkRoundTrip(t, ` + "`" + `{"optional":null,"items":[null,[]],"values":{"null":null,"zero":[]},"nested":{"value":null}}` + "`" + `, []string{})
 checkRoundTrip(t, ` + "`" + `{"optional":null,"items":[null,""],"values":{"null":null,"zero":""},"nested":{"value":null}}` + "`" + `, []byte{})
 checkRoundTrip(t, ` + "`" + `{"optional":null,"items":[null,""],"values":{"null":null,"zero":""},"nested":{"value":null}}` + "`" + `, new(""))
}
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(output, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	testutil.UseLocalSkel(t, output)
	testutil.Go(t, output, "test", ".")
}
