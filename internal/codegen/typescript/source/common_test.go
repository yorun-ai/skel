package source

import (
	"sort"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/codegentest"
	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

func TestCastTypeMapsBinaryToUint8Array(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind:   model.TypeKindScalar,
		Scalar: model.ScalarBinary,
	})
	if got.Plain != "Uint8Array" {
		t.Fatalf("unexpected binary type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsUUIDToString(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind:   model.TypeKindScalar,
		Scalar: model.ScalarUUID,
	})
	if got.Plain != "string" {
		t.Fatalf("unexpected uuid type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsJSONToString(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind:   model.TypeKindScalar,
		Scalar: model.ScalarJSON,
	})
	if got.Plain != "string" {
		t.Fatalf("unexpected json type mapping: %s", got.Plain)
	}
}

func TestCastMapTypeMapsUUIDKeyToString(t *testing.T) {
	got := (_Types{}).castType(codegentest.MapType(codegentest.UUIDType(), codegentest.StringType()))
	if got.Plain != "Record<string, string>" {
		t.Fatalf("unexpected uuid-keyed map type: %s", got.Plain)
	}
}

func TestCastTypeQualifiesExternalDataWithAlias(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind: model.TypeKindData,
		Data: &model.Data{
			Name: "UserSummary",
		},
		ExternalAlias: "userpub",
	})
	if got.Plain != "userpub.UserSummary" {
		t.Fatalf("unexpected external type: %s", got.Plain)
	}
}

func TestCastTypeQualifiesExternalEnumWithAlias(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind: model.TypeKindEnum,
		Enum: &model.Enum{
			Name: "UserStatus",
		},
		ExternalAlias: "userpub",
	})
	if got.Plain != "userpub.UserStatus" {
		t.Fatalf("unexpected external type: %s", got.Plain)
	}
}

func buildModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()
	for _, enum := range spec.Enums {
		if enum.UnspecifiedItem == nil {
			enum.UnspecifiedItem = &model.EnumItem{Name: "UNSPECIFIED"}
		}
	}
	sort.Slice(spec.Enums, func(i, j int) bool { return spec.Enums[i].Name < spec.Enums[j].Name })
	sort.Slice(spec.Data, func(i, j int) bool { return spec.Data[i].Name < spec.Data[j].Name })
	sort.Slice(spec.Services, func(i, j int) bool { return spec.Services[i].Name < spec.Services[j].Name })

	return model.NewDomainFromSpec(spec)
}

func externalDataTypeForTest(data *model.Data, domainName string, alias string, explicitAlias bool) *model.Type {
	type_ := codegentest.DataType(data)
	type_.ExternalDomain = domainName
	type_.ExternalAlias = alias
	type_.ExternalAliasExplicit = explicitAlias
	return type_
}

func renderTemplate(t *testing.T, template string, payload any) string {
	t.Helper()
	content, err := common.RenderTemplate(template, payload)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func TestCastNullableTypeParameterInCollectionsAndArguments(t *testing.T) {
	parameter := codegentest.TypeParam("TValue")
	plain := codegentest.TypeParamType(parameter)
	nullable := codegentest.NullableType(codegentest.TypeParamType(parameter))
	box := &model.Data{Name: "Box", TypeParameters: []*model.TypeParameter{parameter}}
	for _, test := range []struct {
		name string
		kind *model.Type
		want string
	}{
		{name: "parameter", kind: plain, want: "TValue"},
		{name: "nullable parameter", kind: nullable, want: "TValue | null"},
		{name: "list elements", kind: codegentest.ListType(nullable), want: "Array<TValue | null>"},
		{name: "map values", kind: codegentest.MapType(codegentest.StringType(), nullable), want: "Record<string, TValue | null>"},
		{name: "nested generic", kind: codegentest.DataType(box, nullable), want: "Box<TValue | null>"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := (_Types{}).castType(test.kind)
			if got.Plain != test.want {
				t.Fatalf("type: got %q want %q", got.Plain, test.want)
			}
		})
	}
}
