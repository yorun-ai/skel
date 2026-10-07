// Package codegentest holds the schema constructors shared by generator tests.
// The type constructors are pure, so they need no testing.T; only test files
// import this package and it never ships in a build.
package codegentest

import "go.yorun.ai/skel/schema"

// ScalarType builds a type that refers to the given scalar.
func ScalarType(scalar schema.Scalar) *schema.Type {
	return new(schema.Type{Kind: schema.TypeKindScalar, Scalar: scalar})
}

// StringType builds a string type.
func StringType() *schema.Type {
	return ScalarType(schema.ScalarString)
}

// IntType builds an int type.
func IntType() *schema.Type {
	return ScalarType(schema.ScalarInt)
}

// LocalDateTimeType builds a local date-time type.
func LocalDateTimeType() *schema.Type {
	return ScalarType(schema.ScalarLocalDateTime)
}

// BinaryType builds a binary type.
func BinaryType() *schema.Type {
	return ScalarType(schema.ScalarBinary)
}

// UUIDType builds a uuid type.
func UUIDType() *schema.Type {
	return ScalarType(schema.ScalarUUID)
}

// NullableType marks value nullable and returns it.
func NullableType(value *schema.Type) *schema.Type {
	value.Nullable = true
	return value
}

// ListType builds a list of value.
func ListType(value *schema.Type) *schema.Type {
	return new(schema.Type{Kind: schema.TypeKindList, List: new(schema.ListType{Element: value})})
}

// MapType builds a map from key to value.
func MapType(key *schema.Type, value *schema.Type) *schema.Type {
	return new(schema.Type{Kind: schema.TypeKindMap, Map: new(schema.MapType{Key: key, Value: value})})
}

// DataType builds a reference to data with optional type arguments.
func DataType(data *schema.Data, typeArgs ...*schema.Type) *schema.Type {
	return new(schema.Type{
		Kind:          schema.TypeKindData,
		Data:          data,
		SkelName:      data.SkelName,
		TypeArguments: typeArgs,
	})
}

// EnumType builds a reference to enum.
func EnumType(enum *schema.Enum) *schema.Type {
	return new(schema.Type{
		Kind:     schema.TypeKindEnum,
		Enum:     enum,
		SkelName: enum.SkelName,
	})
}

// TypeParam declares a type parameter named name.
func TypeParam(name string) *schema.TypeParameter {
	return new(schema.TypeParameter{Name: name})
}

// TypeParamType builds a reference to typeParam.
func TypeParamType(typeParam *schema.TypeParameter) *schema.Type {
	return new(schema.Type{Kind: schema.TypeKindTypeParameter, TypeParameter: typeParam})
}

// ActorVia builds the actor via identified by kind.
func ActorVia(kind schema.ActorViaKind) *schema.ActorVia {
	return new(schema.ActorVia{Name: string(kind)})
}

// DomainSchema builds a named domain spec.
func DomainSchema(name string) schema.DomainSpec {
	return DomainSchemaWithDescription(name, "")
}

// DomainSchemaWithDescription builds a named domain spec carrying description.
func DomainSchemaWithDescription(name string, description string) schema.DomainSpec {
	return schema.DomainSpec{Name: name, Description: description}
}
