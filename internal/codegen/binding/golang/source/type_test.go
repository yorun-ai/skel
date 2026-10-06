package source

import (
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
)

func TestCastTypeUsesDefaultExternalPubPackageNameWithoutImportAlias(t *testing.T) {
	kindRef := &model.Type{
		Kind:          model.TypeKindData,
		Data:          &model.Data{Name: "UserSummary"},
		ExternalAlias: "userpub",
	}
	got := (_Types{bindings: binding.TypeBindings{kindRef: {Domain: kindRef.ExternalDomain, Alias: "userpub", Path: "go.yorun.ai/app/vine/demo/user/userpub", Explicit: false}}}).castType(kindRef)
	if got.Plain != "userpub.UserSummary" {
		t.Fatalf("unexpected external type: %s", got.Plain)
	}
	if len(got.Imports) != 1 {
		t.Fatalf("unexpected imports: %+v", got.Imports)
	}
	if got.Imports[0].Alias != "" {
		t.Fatalf("unexpected import alias: %s", got.Imports[0].Alias)
	}
}

func TestCastTypePreservesExplicitExternalImportAlias(t *testing.T) {
	kindRef := &model.Type{
		Kind:                  model.TypeKindData,
		Data:                  &model.Data{Name: "UserSummary"},
		ExternalAlias:         "account",
		ExternalAliasExplicit: true,
	}
	got := (_Types{bindings: binding.TypeBindings{kindRef: {Domain: kindRef.ExternalDomain, Alias: "account", Path: "go.yorun.ai/app/vine/demo/user/userpub", Explicit: true}}}).castType(kindRef)
	if got.Plain != "account.UserSummary" {
		t.Fatalf("unexpected external type: %s", got.Plain)
	}
	if len(got.Imports) != 1 {
		t.Fatalf("unexpected imports: %+v", got.Imports)
	}
	if got.Imports[0].Alias != "account" {
		t.Fatalf("unexpected import alias: %s", got.Imports[0].Alias)
	}
}

func TestCastEnumTypeUsesQualifiedUnspecifiedDefaultValue(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind: model.TypeKindEnum,
		Enum: &model.Enum{Name: "UserStatus", UnspecifiedItem: &model.EnumItem{Name: "UNSPECIFIED"}},
	})
	if got.DefaultValue != "UserStatusUnspecified" {
		t.Fatalf("unexpected default value: %s", got.DefaultValue)
	}
}

func TestCastExternalEnumTypeUsesQualifiedUnspecifiedDefaultValue(t *testing.T) {
	kindRef := &model.Type{
		Kind:          model.TypeKindEnum,
		Enum:          &model.Enum{Name: "UserStatus", UnspecifiedItem: &model.EnumItem{Name: "UNSPECIFIED"}},
		ExternalAlias: "userpub",
	}
	got := (_Types{bindings: binding.TypeBindings{kindRef: {Domain: kindRef.ExternalDomain, Alias: "userpub", Path: "go.yorun.ai/app/vine/demo/user/userpub", Explicit: false}}}).castType(kindRef)
	if got.DefaultValue != "userpub.UserStatusUnspecified" {
		t.Fatalf("unexpected default value: %s", got.DefaultValue)
	}
}

func TestCastMapTypeMapsUUIDKeyToSkelUUID(t *testing.T) {
	got := (_Types{}).castType(&model.Type{
		Kind: model.TypeKindMap,
		Map: &model.MapType{
			Key:   &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarUUID},
			Value: &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString},
		},
	})

	if got.Plain != "map[types.UUID]string" {
		t.Fatalf("unexpected map type: %s", got.Plain)
	}
	if len(got.Imports) != 1 || got.Imports[0].Path != typesImport {
		t.Fatalf("unexpected map imports: %+v", got.Imports)
	}
}

func TestCastCollectionTypesUsePointersOnlyWhenNullable(t *testing.T) {
	tests := []struct {
		name        string
		type_       *model.Type
		wantPlain   string
		wantDefault string
	}{
		{name: "list", type_: codegentest.ListType(codegentest.StringType()), wantPlain: "[]string", wantDefault: "[]string{}"},
		{name: "nullable list", type_: codegentest.NullableType(codegentest.ListType(codegentest.StringType())), wantPlain: "*[]string", wantDefault: "nil"},
		{name: "map", type_: codegentest.MapType(codegentest.StringType(), codegentest.StringType()), wantPlain: "map[string]string", wantDefault: "map[string]string{}"},
		{name: "nullable map", type_: codegentest.NullableType(codegentest.MapType(codegentest.StringType(), codegentest.StringType())), wantPlain: "*map[string]string", wantDefault: "nil"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := (_Types{}).castType(test.type_)
			if got.Plain != test.wantPlain || got.DefaultValue != test.wantDefault {
				t.Fatalf("unexpected collection type: plain=%q default=%q", got.Plain, got.DefaultValue)
			}
		})
	}
}

func TestCastTypeEmitsGeneratedCollisionAlias(t *testing.T) {
	for _, kind := range []model.TypeKind{model.TypeKindData, model.TypeKindEnum} {
		kindRef := &model.Type{Kind: kind, Data: &model.Data{Name: "Value"}, Enum: &model.Enum{Name: "Value", UnspecifiedItem: &model.EnumItem{Name: "UNSPECIFIED"}}, ExternalDomain: "first.user", ExternalAlias: "firstuser"}
		got := (_Types{bindings: binding.TypeBindings{kindRef: {Domain: kindRef.ExternalDomain, Alias: "firstuser", Path: "example.com/first/userpub"}}}).castType(kindRef)
		if len(got.Imports) != 1 || got.Imports[0].Alias != "firstuser" || got.Plain != "firstuser.Value" {
			t.Fatalf("collision alias lost: %+v", got)
		}
	}
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
		{name: "nullable parameter", kind: nullable, want: "*TValue"},
		{name: "list elements", kind: codegentest.ListType(nullable), want: "[]*TValue"},
		{name: "map values", kind: codegentest.MapType(codegentest.StringType(), nullable), want: "map[string]*TValue"},
		{name: "nested generic", kind: codegentest.DataType(box, nullable), want: "Box[*TValue]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := (_Types{}).castType(test.kind)
			if got.Plain != test.want {
				t.Fatalf("type: got %q want %q", got.Plain, test.want)
			}
		})
	}
	if got := (_Types{}).castType(nullable); got.DefaultValue != "nil" {
		t.Fatalf("nullable default: %q", got.DefaultValue)
	}
}
