package schema_test

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestExternalTypeTraversal(t *testing.T) {
	attachment := new(schema.Data{
		Name: "Attachment",
		Members: []*schema.DataMember{
			new(schema.DataMember{
				Name: "content",
				Type: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarBinary}),
			}),
		},
	})
	attachmentType := new(schema.Type{Kind: schema.TypeKindData, Data: attachment})

	if attachmentType.Name() != "Attachment" {
		t.Fatalf("Name() = %q, want Attachment", attachmentType.Name())
	}
	if !attachmentType.ContainsBinaryType() {
		t.Fatal("ContainsBinaryType() = false, want true")
	}
}

func TestExternalEnumValues(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "actor via", got: schema.ActorViaOpenAPI, want: schema.ActorViaKind("openapi")},
		{name: "data kind", got: schema.DataKindConfig, want: schema.DataKind("config")},
		{name: "config lifecycle", got: schema.ConfigLifecycleInstant, want: schema.ConfigLifecycle("instant")},
		{name: "auth mode", got: schema.AuthModeNoAuth, want: schema.AuthMode("noauth")},
		{name: "permission mode", got: schema.PermissionRequireModeAny, want: schema.PermissionRequireMode("any")},
		{name: "type kind", got: schema.TypeKindScalar, want: schema.TypeKind(2)},
		{name: "scalar", got: schema.ScalarJSON, want: schema.Scalar(13)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("value = %v, want %v", test.got, test.want)
			}
		})
	}
}
