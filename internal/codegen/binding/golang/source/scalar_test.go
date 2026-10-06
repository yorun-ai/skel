package source

import (
	"testing"

	"go.yorun.ai/skel/internal/model"
)

func TestCastTypeMapsBinaryToBytes(t *testing.T) {
	got := (_Types{}).castType(&model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarBinary})
	if got.Plain != "types.Binary" {
		t.Fatalf("unexpected binary type mapping: %s", got.Plain)
	}
	if got.Imports[0].Path != typesImport {
		t.Fatalf("unexpected binary import: %s", got.Imports[0].Path)
	}
}

func TestCastTypeMapsUUIDToSkelUUID(t *testing.T) {
	got := (_Types{}).castType(&model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarUUID})
	if got.Plain != "types.UUID" {
		t.Fatalf("unexpected uuid type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsJSONToSkelJSON(t *testing.T) {
	got := (_Types{}).castType(&model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarJSON})
	if got.Plain != "types.JSON" {
		t.Fatalf("unexpected json type mapping: %s", got.Plain)
	}
}
