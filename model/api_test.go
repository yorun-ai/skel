package model_test

import (
	"fmt"
	"testing"

	"go.yorun.ai/skel/model"
)

func ExampleNewDomainFromSpec() {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{{Name: "User"}},
	})

	fmt.Println(domain.Name())
	fmt.Println(domain.Data()[0].Name)

	// Output:
	// demo.user
	// User
}

func TestFacadePreservesSemanticModelMethods(t *testing.T) {
	attachment := new(model.Data{
		Name: "Attachment",
		Members: []*model.DataMember{
			new(model.DataMember{
				Name: "content",
				Type: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarBinary}),
			}),
		},
	})
	attachmentType := new(model.Type{Kind: model.TypeKindData, Data: attachment})

	if attachmentType.Name() != "Attachment" {
		t.Fatalf("Name() = %q, want Attachment", attachmentType.Name())
	}
	if !attachmentType.ContainsBinaryType() {
		t.Fatal("ContainsBinaryType() = false, want true")
	}
}

func TestFacadeEnumValues(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "actor via", got: model.ActorViaOpenAPI, want: model.ActorViaKind("openapi")},
		{name: "data kind", got: model.DataKindConfig, want: model.DataKind("config")},
		{name: "config lifecycle", got: model.ConfigLifecycleInstant, want: model.ConfigLifecycle("instant")},
		{name: "auth mode", got: model.AuthModeNoAuth, want: model.AuthMode("noauth")},
		{name: "permission mode", got: model.PermissionRequireModeAny, want: model.PermissionRequireMode("any")},
		{name: "type kind", got: model.TypeKindScalar, want: model.TypeKind(2)},
		{name: "scalar", got: model.ScalarJSON, want: model.Scalar(13)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("value = %v, want %v", test.got, test.want)
			}
		})
	}
}
