package schema_test

import (
	"reflect"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestWireEnums(t *testing.T) {
	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{
			name: "declaration type",
			got: stringsOf(
				schema.DeclarationTypeActor,
				schema.DeclarationTypeConfig,
				schema.DeclarationTypeData,
				schema.DeclarationTypeEnum,
				schema.DeclarationTypeEvent,
				schema.DeclarationTypeResource,
				schema.DeclarationTypeService,
				schema.DeclarationTypeTask,
				schema.DeclarationTypeWeb,
			),
			want: []string{"actor", "config", "data", "enum", "event", "resource", "service", "task", "web"},
		},
		{
			name: "config lifecycle",
			got:  stringsOf(schema.ConfigLifecycleEternal, schema.ConfigLifecycleInstant),
			want: []string{"eternal", "instant"},
		},
		{
			name: "type kind",
			got: stringsOf(
				schema.TypeKindScalar,
				schema.TypeKindEnum,
				schema.TypeKindData,
				schema.TypeKindConfig,
				schema.TypeKindEvent,
				schema.TypeKindTypeParameter,
				schema.TypeKindImportedReference,
				schema.TypeKindList,
				schema.TypeKindMap,
			),
			want: []string{"scalar", "enum", "data", "config", "event", "typeParameter", "importedReference", "list", "map"},
		},
		{
			name: "authentication mode",
			got:  stringsOf(schema.AuthModeUnset, schema.AuthModeAuth, schema.AuthModeNoAuth),
			want: []string{"unset", "auth", "noauth"},
		},
		{
			name: "requirement mode",
			got: stringsOf(
				schema.RequirementModeCode,
				schema.RequirementModeReference,
				schema.RequirementModeCheck,
				schema.RequirementModeAll,
				schema.RequirementModeAny,
			),
			want: []string{"code", "reference", "check", "all", "any"},
		},
		{
			name: "impact level",
			got:  stringsOf(schema.ImpactBreaking, schema.ImpactDangerous, schema.ImpactCompatible),
			want: []string{"BREAKING", "DANGEROUS", "COMPATIBLE"},
		},
		{
			name: "change type",
			got:  stringsOf(schema.ChangeAdded, schema.ChangeRemoved, schema.ChangeModified),
			want: []string{"ADDED", "REMOVED", "MODIFIED"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("wire values = %v, want %v", test.got, test.want)
			}
		})
	}
}

func stringsOf[T ~string](values ...T) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}
