package schema

import (
	"fmt"
	"strings"
)

// TypeParameter describes a generic data type parameter.
type TypeParameter struct {
	// Pos is the parameter declaration's source position.
	Pos Position
	// Name is the parameter's local name.
	Name string
}

// Type describes a semantic type. Named references may remain unresolved while
// a domain is undergoing semantic analysis.
//
// Kind selects the corresponding Scalar, List, Map, Enum, Data, or
// TypeParameter field. An unresolved named reference uses SkelName while it is
// being analyzed. TypeArguments applies to generic data references.
type Type struct {
	// Pos is the type expression's source position.
	Pos Position
	// ReferencePos is the source position of a named type reference.
	ReferencePos Position
	// Kind identifies the type representation.
	Kind TypeKind
	// Scalar carries the value for TypeKindScalar.
	Scalar Scalar
	// List carries the value for TypeKindList.
	List *ListType
	// Map carries the value for TypeKindMap.
	Map *MapType
	// Enum is the resolved declaration for TypeKindEnum.
	Enum *Enum
	// Data is the resolved declaration for TypeKindData.
	Data *Data
	// TypeParameter is the resolved parameter for TypeKindTypeParameter.
	TypeParameter *TypeParameter
	// SkelName is the referenced declaration's fully qualified Skel name.
	SkelName string
	// Nullable reports whether the type accepts null.
	Nullable bool
	// TypeArguments contains resolved generic arguments in declaration order.
	TypeArguments []*Type
	// ExternalDomain is the fully qualified domain name for an imported type.
	ExternalDomain string
	// ExternalAlias is the source qualifier for an imported type.
	ExternalAlias string
	// ExternalAliasExplicit reports whether ExternalAlias was declared in source.
	ExternalAliasExplicit bool
}

// Name returns a stable identifier-style name for t.
// Generic, list, and map names recursively include their component types.
func (t *Type) Name() string {
	switch t.Kind {
	case TypeKindScalar:
		return t.Scalar.Name()
	case TypeKindList:
		return fmt.Sprintf("ListOf%s", t.List.Value.Name())
	case TypeKindMap:
		return fmt.Sprintf("MapOf%sAnd%s", t.Map.Key.Name(), t.Map.Value.Name())
	case TypeKindEnum:
		return t.Enum.Name
	case TypeKindData:
		if len(t.TypeArguments) == 0 {
			return t.Data.Name
		}
		names := make([]string, 0, len(t.TypeArguments))
		for _, typeArg := range t.TypeArguments {
			names = append(names, typeArg.Name())
		}
		return fmt.Sprintf("%sOf%s", t.Data.Name, strings.Join(names, "And"))
	default:
		return fmt.Sprintf("UnknownTypeKind%d", t.Kind)
	}
}

// ContainsBinaryType reports whether t is binary or recursively contains a
// binary member. It handles recursive data declarations and a nil receiver.
func (t *Type) ContainsBinaryType() bool {
	return t.containsBinaryType(nil, map[_BinaryVisit]bool{})
}

type _BinaryVisit struct {
	data      *Data
	arguments string
}

func (t *Type) containsBinaryType(parameters map[*TypeParameter]bool, visited map[_BinaryVisit]bool) bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case TypeKindScalar:
		return t.Scalar == ScalarBinary
	case TypeKindTypeParameter:
		return parameters[t.TypeParameter]
	case TypeKindList:
		return t.List.Value.containsBinaryType(parameters, visited)
	case TypeKindMap:
		return t.Map.Key.containsBinaryType(parameters, visited) || t.Map.Value.containsBinaryType(parameters, visited)
	case TypeKindData:
		bound := map[*TypeParameter]bool{}
		arguments := make([]byte, len(t.TypeArguments))
		for i, arg := range t.TypeArguments {
			if arg.containsBinaryType(parameters, visited) {
				bound[t.Data.TypeParameters[i]] = true
				arguments[i] = 1
			}
		}
		key := _BinaryVisit{data: t.Data, arguments: string(arguments)}
		if visited[key] {
			return false
		}
		visited[key] = true
		defer delete(visited, key)
		for _, member := range t.Data.Members {
			if member.Type.containsBinaryType(bound, visited) {
				return true
			}
		}
	}
	return false
}

// TypeKind identifies the representation carried by a [Type].
type TypeKind int

const (
	// TypeKindUnresolvedReference identifies a named reference that has not yet
	// been resolved during semantic analysis.
	TypeKindUnresolvedReference TypeKind = -1
	// TypeKindScalar identifies a built-in scalar type.
	TypeKindScalar TypeKind = 2
	// TypeKindList identifies a list type.
	TypeKindList TypeKind = 3
	// TypeKindMap identifies a map type.
	TypeKindMap TypeKind = 4
	// TypeKindEnum identifies a resolved enum reference.
	TypeKindEnum TypeKind = 5
	// TypeKindData identifies a resolved data reference.
	TypeKindData TypeKind = 6
	// TypeKindTypeParameter identifies a generic type parameter reference.
	TypeKindTypeParameter TypeKind = 7
)

// Scalar identifies a built-in Skel scalar type.
type Scalar int

const (
	// ScalarInt identifies an integer.
	ScalarInt Scalar = iota + 1
	// ScalarFloat identifies a floating-point number.
	ScalarFloat
	// ScalarBoolean identifies a boolean.
	ScalarBoolean
	// ScalarString identifies a UTF-8 string.
	ScalarString
	// ScalarDecimal identifies an exact decimal value.
	ScalarDecimal
	// ScalarBinary identifies binary data.
	ScalarBinary
	// ScalarTimestamp identifies an absolute timestamp.
	ScalarTimestamp
	// ScalarDuration identifies a duration.
	ScalarDuration
	// ScalarLocalDate identifies a calendar date without a timezone.
	ScalarLocalDate
	// ScalarLocalTime identifies a time of day without a timezone.
	ScalarLocalTime
	// ScalarLocalDateTime identifies a date and time without a timezone.
	ScalarLocalDateTime
	// ScalarUUID identifies a UUID.
	ScalarUUID
	// ScalarJSON identifies an arbitrary JSON value.
	ScalarJSON
)

var scalarNames = map[Scalar]string{
	ScalarInt: "Int", ScalarFloat: "Float", ScalarBoolean: "Bool", ScalarString: "String",
	ScalarDecimal: "Decimal", ScalarBinary: "Binary", ScalarTimestamp: "Timestamp",
	ScalarDuration: "Duration", ScalarLocalDate: "LocalDate", ScalarLocalTime: "LocalTime",
	ScalarLocalDateTime: "LocalDateTime", ScalarUUID: "UUID", ScalarJSON: "JSON",
}

// Name returns the canonical identifier-style name of s.
func (s Scalar) Name() string {
	if name, ok := scalarNames[s]; ok {
		return name
	}
	return fmt.Sprintf("UnknownScalar%d", s)
}

// ListType describes the element type of a list.
type ListType struct {
	// Value is the list element type.
	Value *Type
}

// MapType describes the key and value types of a map.
type MapType struct {
	// Key is the map key type.
	Key *Type
	// Value is the map value type.
	Value *Type
}
