package codegen

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestBuildPublicViewRejectsNonPublicReferences(t *testing.T) {
	tests := []struct {
		name     string
		spec     schema.DomainSpec
		expected string
	}{
		{
			name: "non-pub enum",
			spec: schema.DomainSpec{
				Data: []*schema.Data{{Pub: true, Name: "Payload", Members: []*schema.DataMember{{
					Name: "status",
					Type: &schema.Type{
						Kind:           schema.TypeKindEnum,
						Enum:           &schema.Enum{Name: "Status"},
						ExternalDomain: "demo.shared",
					},
				}}}},
			},
			expected: "references non-pub enum Status",
		},
		{
			name: "non-pub imported data",
			spec: schema.DomainSpec{
				Data: []*schema.Data{{Pub: true, Name: "Payload", Members: []*schema.DataMember{{
					Name: "remote",
					Type: &schema.Type{
						Kind:           schema.TypeKindData,
						Data:           &schema.Data{Kind: schema.DataKindData, Name: "Remote"},
						ExternalDomain: "demo.shared",
					},
				}}}},
			},
			expected: "references non-pub data Remote",
		},
		{
			name: "invalid list type",
			spec: schema.DomainSpec{
				Data: []*schema.Data{{Pub: true, Name: "Payload", Members: []*schema.DataMember{{
					Name: "values",
					Type: &schema.Type{Kind: schema.TypeKindList},
				}}}},
			},
			expected: "contains an invalid list type",
		},
		{
			name: "invalid map type",
			spec: schema.DomainSpec{
				Data: []*schema.Data{{Pub: true, Name: "Payload", Members: []*schema.DataMember{{
					Name: "values",
					Type: &schema.Type{Kind: schema.TypeKindMap},
				}}}},
			},
			expected: "contains an invalid map type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			domain := schema.NewDomainFromSpec(test.spec)
			_, err := BuildPublicView(domain)
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, err)
			}
		})
	}
}

func TestBuildPublicViewValidatesNestedPublicDataOnce(t *testing.T) {
	shared := &schema.Data{Pub: true, Name: "Shared", Members: []*schema.DataMember{{
		Name: "label",
		Type: &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString},
	}}}
	payload := &schema.Data{Pub: true, Name: "Payload", Members: []*schema.DataMember{
		{Name: "shared", Type: &schema.Type{Kind: schema.TypeKindData, Data: shared}},
		{Name: "sharedAgain", Type: &schema.Type{Kind: schema.TypeKindData, Data: shared}},
	}}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo.user", Data: []*schema.Data{payload, shared}})

	view, err := BuildPublicView(domain)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Data) != 2 {
		t.Fatalf("unexpected public data: %+v", view.Data)
	}
}
