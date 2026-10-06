// Package codegentest holds the model constructors shared by generator tests.
// The type constructors are pure, so they need no testing.T; only test files
// import this package and it never ships in a build.
package codegentest

import "go.yorun.ai/skel/internal/model"

// ScalarType builds a type that refers to the given scalar.
func ScalarType(scalar model.Scalar) *model.Type {
	return new(model.Type{Kind: model.TypeKindScalar, Scalar: scalar})
}

// StringType builds a string type.
func StringType() *model.Type {
	return ScalarType(model.ScalarString)
}

// IntType builds an int type.
func IntType() *model.Type {
	return ScalarType(model.ScalarInt)
}

// LocalDateTimeType builds a local date-time type.
func LocalDateTimeType() *model.Type {
	return ScalarType(model.ScalarLocalDateTime)
}

// BinaryType builds a binary type.
func BinaryType() *model.Type {
	return ScalarType(model.ScalarBinary)
}

// UUIDType builds a uuid type.
func UUIDType() *model.Type {
	return ScalarType(model.ScalarUUID)
}

// NullableType marks value nullable and returns it.
func NullableType(value *model.Type) *model.Type {
	value.Nullable = true
	return value
}

// ListType builds a list of value.
func ListType(value *model.Type) *model.Type {
	return new(model.Type{Kind: model.TypeKindList, List: new(model.ListType{Value: value})})
}

// MapType builds a map from key to value.
func MapType(key *model.Type, value *model.Type) *model.Type {
	return new(model.Type{Kind: model.TypeKindMap, Map: new(model.MapType{Key: key, Value: value})})
}

// DataType builds a reference to data with optional type arguments.
func DataType(data *model.Data, typeArgs ...*model.Type) *model.Type {
	return new(model.Type{
		Kind:          model.TypeKindData,
		Data:          data,
		SkelName:      data.SkelName,
		TypeArguments: typeArgs,
	})
}

// EnumType builds a reference to enum.
func EnumType(enum *model.Enum) *model.Type {
	return new(model.Type{
		Kind:     model.TypeKindEnum,
		Enum:     enum,
		SkelName: enum.SkelName,
	})
}

// TypeParam declares a type parameter named name.
func TypeParam(name string) *model.TypeParameter {
	return new(model.TypeParameter{Name: name})
}

// TypeParamType builds a reference to typeParam.
func TypeParamType(typeParam *model.TypeParameter) *model.Type {
	return new(model.Type{Kind: model.TypeKindTypeParameter, TypeParameter: typeParam})
}

// ActorVia builds the actor via identified by kind.
func ActorVia(kind model.ActorViaKind) *model.ActorVia {
	return new(model.ActorVia{Name: string(kind)})
}

// DomainModel builds a named domain spec.
func DomainModel(name string) model.DomainSpec {
	return DomainModelWithDescription(name, "")
}

// DomainModelWithDescription builds a named domain spec carrying description.
func DomainModelWithDescription(name string, description string) model.DomainSpec {
	return model.DomainSpec{Name: name, Description: description}
}
