package analyzer

import (
	"go.yorun.ai/skel/internal/model"
	"testing"
)

func TestTypeContainsBinaryType(t *testing.T) {
	asset := &model.Data{
		Name: "Asset",
		Members: []*model.DataMember{
			{
				Name: "Payload",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarBinary,
				},
			},
		},
	}
	wrapper := &model.Type{
		Kind: model.TypeKindList,
		List: &model.ListType{
			Value: &model.Type{
				Kind: model.TypeKindMap,
				Map: &model.MapType{
					Key: &model.Type{
						Kind:   model.TypeKindScalar,
						Scalar: model.ScalarString,
					},
					Value: &model.Type{
						Kind: model.TypeKindData,
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
