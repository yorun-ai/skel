package analyzer

import (
	"go.yorun.ai/skel/schema"
	"testing"
)

func TestTypeContainsBinaryType(t *testing.T) {
	asset := &schema.Data{
		Name: "Asset",
		Members: []*schema.DataMember{
			{
				Name: "Payload",
				Type: &schema.Type{
					Kind:   schema.TypeKindScalar,
					Scalar: schema.ScalarBinary,
				},
			},
		},
	}
	wrapper := &schema.Type{
		Kind: schema.TypeKindList,
		List: &schema.ListType{
			Value: &schema.Type{
				Kind: schema.TypeKindMap,
				Map: &schema.MapType{
					Key: &schema.Type{
						Kind:   schema.TypeKindScalar,
						Scalar: schema.ScalarString,
					},
					Value: &schema.Type{
						Kind: schema.TypeKindData,
						Data: asset,
					},
				},
			},
		},
	}

	if !wrapper.ContainsBinaryType() {
		t.Fatalf("expected nested type to contain binary type")
	}
}
