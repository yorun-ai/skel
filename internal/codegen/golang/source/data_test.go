package source

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/golang/view"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/testutil"
)

func TestCastData(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name:        "Page",
		Description: "Paginated result",
		Sensitive:   true,
		TypeParameters: []*model.TypeParameter{
			{Name: "TItem"},
		},
		Members: []*model.DataMember{
			{
				Name:        "generatedAt",
				Description: "Generated at",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarTimestamp,
				},
			},
			{
				Name:        "avatarUrl",
				Description: "Avatar URL",
				Example:     `"https://xxx.com/a.png"`,
				Type: &model.Type{
					Kind:     model.TypeKindScalar,
					Scalar:   model.ScalarString,
					Nullable: true,
				},
			},
		},
	})

	if data.FullName != "Page[TItem any]" {
		t.Fatalf("unexpected full name: %s", data.FullName)
	}
	if !data.Sensitive {
		t.Fatal("expected sensitive data marker")
	}
	if len(data.CommentLines) == 0 || data.CommentLines[0] != "Page Paginated result" {
		t.Fatalf("unexpected data comment lines: %+v", data.CommentLines)
	}
	if len(data.Members) != 2 {
		t.Fatalf("unexpected member count: %d", len(data.Members))
	}
	if data.Members[0].Type.Plain != "skel.Timestamp" {
		t.Fatalf("unexpected first member type: %s", data.Members[0].Type.Plain)
	}
	if len(data.Members[1].CommentLines) == 0 || data.Members[1].CommentLines[0] != `AvatarUrl Avatar URL (e.g. "https://xxx.com/a.png")` {
		t.Fatalf("unexpected second member comment lines: %+v", data.Members[1].CommentLines)
	}
}

func TestBuildDataImports(t *testing.T) {
	imports := buildDataImports([]*Data{
		{
			Sensitive: true,
			Members: []*DataMember{
				{Type: &Type{Imports: []*Import{{Path: skelImport}}}},
				{Type: &Type{Imports: []*Import{{Path: skelImport}}}},
			},
		},
	})

	if got, want := importPaths(imports), []string{skelImport}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected import paths: got=%v want=%v", got, want)
	}
}

func TestSensitiveMarkerMethodNeedsNoImport(t *testing.T) {
	if imports := buildDataImports([]*Data{{Sensitive: true}}); len(imports) != 0 {
		t.Fatalf("unexpected imports for sensitive marker method: %v", importPaths(imports))
	}
}

func TestCastDataMapsDurationToSkelDuration(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name: "TimeoutConfig",
		Members: []*model.DataMember{
			{
				Name: "timeout",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarDuration,
				},
			},
		},
	})
	if data.Members[0].Type.Plain != "skel.Duration" {
		t.Fatalf("unexpected duration member type: %s", data.Members[0].Type.Plain)
	}
}

func TestCastDataMapsLocalDateToSkelLocalDate(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name: "Profile",
		Members: []*model.DataMember{
			{
				Name: "birthday",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarLocalDate,
				},
			},
		},
	})
	if data.Members[0].Type.Plain != "skel.LocalDate" {
		t.Fatalf("unexpected date member type: %s", data.Members[0].Type.Plain)
	}
}

func TestGeneratedNullableTypeParametersRoundTrip(t *testing.T) {
	parameter := codegentest.TypeParam("TValue")
	nullable := func() *model.Type { return codegentest.NullableType(codegentest.TypeParamType(parameter)) }
	box := &model.Data{Name: "Box", TypeParameters: []*model.TypeParameter{parameter}, Members: []*model.DataMember{{Name: "value", Type: codegentest.TypeParamType(parameter)}}}
	wrapper := &model.Data{Name: "Wrapper", TypeParameters: []*model.TypeParameter{parameter}, Members: []*model.DataMember{
		{Name: "optional", Type: nullable()},
		{Name: "items", Type: codegentest.ListType(nullable())},
		{Name: "values", Type: codegentest.MapType(codegentest.StringType(), nullable())},
		{Name: "nested", Type: codegentest.DataType(box, nullable())},
	}}
	domain := buildModelDomainForTest(t, model.DomainSpec{Name: "demo.generic", Data: []*model.Data{box, wrapper}})
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
	testutil.Go(t, output, "test", ".")
}
