package source

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestCastTypeMapsBinaryToBytes(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarBinary})
	if got.Plain != "skeltype.Binary" {
		t.Fatalf("unexpected binary type mapping: %s", got.Plain)
	}
	if got.Imports[0].Path != typesImport {
		t.Fatalf("unexpected binary import: %s", got.Imports[0].Path)
	}
}

func TestCastTypeMapsUUIDToSkelUUID(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarUUID})
	if got.Plain != "skeltype.UUID" {
		t.Fatalf("unexpected uuid type mapping: %s", got.Plain)
	}
}

func TestCastTypeMapsJSONToSkelJSON(t *testing.T) {
	got := (_Types{}).castType(&schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarJSON})
	if got.Plain != "skeltype.JSON" {
		t.Fatalf("unexpected json type mapping: %s", got.Plain)
	}
}
