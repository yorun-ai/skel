package common

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/model"
)

func TestBuildPublicViewRejectsNonPublicReferences(t *testing.T) {
	tests := []struct {
		name     string
		spec     model.DomainSpec
		expected string
	}{
		{
			name: "non-pub enum",
			spec: model.DomainSpec{
				Data: []*model.Data{{Pub: true, Name: "Payload", Members: []*model.DataMember{{
					Name: "status",
					Type: &model.Type{
						Kind:           model.TypeKindEnum,
						Enum:           &model.Enum{Name: "Status"},
						ExternalDomain: "demo.shared",
					},
				}}}},
			},
			expected: "references non-pub enum Status",
		},
		{
			name: "non-pub imported data",
			spec: model.DomainSpec{
				Data: []*model.Data{{Pub: true, Name: "Payload", Members: []*model.DataMember{{
					Name: "remote",
					Type: &model.Type{
						Kind:           model.TypeKindData,
						Data:           &model.Data{Kind: model.DataKindData, Name: "Remote"},
						ExternalDomain: "demo.shared",
					},
				}}}},
			},
			expected: "references non-pub data Remote",
		},
		{
			name: "invalid list type",
			spec: model.DomainSpec{
				Data: []*model.Data{{Pub: true, Name: "Payload", Members: []*model.DataMember{{
					Name: "values",
					Type: &model.Type{Kind: model.TypeKindList},
				}}}},
			},
			expected: "contains an invalid list type",
		},
		{
			name: "invalid map type",
			spec: model.DomainSpec{
				Data: []*model.Data{{Pub: true, Name: "Payload", Members: []*model.DataMember{{
					Name: "values",
					Type: &model.Type{Kind: model.TypeKindMap},
				}}}},
			},
			expected: "contains an invalid map type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			domain := model.NewDomainFromSpec(test.spec)
			_, err := BuildPublicView(domain)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, err)
			}
		})
	}
}

func TestBuildPublicViewValidatesNestedPublicDataOnce(t *testing.T) {
	shared := &model.Data{Pub: true, Name: "Shared", Members: []*model.DataMember{{
		Name: "label",
		Type: &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString},
	}}}
	payload := &model.Data{Pub: true, Name: "Payload", Members: []*model.DataMember{
		{Name: "shared", Type: &model.Type{Kind: model.TypeKindData, Data: shared}},
		{Name: "sharedAgain", Type: &model.Type{Kind: model.TypeKindData, Data: shared}},
	}}
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo.user", Data: []*model.Data{payload, shared}})

	view, err := BuildPublicView(domain)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Data) != 2 {
		t.Fatalf("unexpected public data: %+v", view.Data)
	}
}
