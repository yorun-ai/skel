package descriptor

// TypeKind identifies a type representation.
type TypeKind string

const (
	// TypeKindScalar identifies the scalar type representation.
	TypeKindScalar TypeKind = "scalar"
	// TypeKindList identifies the list type representation.
	TypeKindList TypeKind = "list"
	// TypeKindMap identifies the map type representation.
	TypeKindMap TypeKind = "map"
	// TypeKindEnum identifies the enum type representation.
	TypeKindEnum TypeKind = "enum"
	// TypeKindData identifies the data type representation.
	TypeKindData TypeKind = "data"
	// TypeKindConfig identifies the config type representation.
	TypeKindConfig TypeKind = "config"
	// TypeKindEvent identifies the event type representation.
	TypeKindEvent TypeKind = "event"
	// TypeKindTypeParameter identifies the typeParameter type representation.
	TypeKindTypeParameter TypeKind = "typeParameter"
)

// Scalar identifies a Skel scalar type.
type Scalar string

const (
	// ScalarString identifies the string scalar.
	ScalarString Scalar = "string"
	// ScalarBool identifies the bool scalar.
	ScalarBool Scalar = "bool"
	// ScalarInt identifies the int scalar.
	ScalarInt Scalar = "int"
	// ScalarLong identifies the long scalar.
	ScalarLong Scalar = "long"
	// ScalarFloat identifies the float scalar.
	ScalarFloat Scalar = "float"
	// ScalarDouble identifies the double scalar.
	ScalarDouble Scalar = "double"
	// ScalarDecimal identifies the decimal scalar.
	ScalarDecimal Scalar = "decimal"
	// ScalarJson identifies the json scalar.
	ScalarJson Scalar = "json"
	// ScalarUuid identifies the uuid scalar.
	ScalarUuid Scalar = "uuid"
	// ScalarTimestamp identifies the timestamp scalar.
	ScalarTimestamp Scalar = "timestamp"
	// ScalarDuration identifies the duration scalar.
	ScalarDuration Scalar = "duration"
	// ScalarLocalDate identifies the localdate scalar.
	ScalarLocalDate Scalar = "localdate"
	// ScalarLocalTime identifies the localtime scalar.
	ScalarLocalTime Scalar = "localtime"
	// ScalarLocalDateTime identifies the localdatetime scalar.
	ScalarLocalDateTime Scalar = "localdatetime"
	// ScalarBinary identifies the binary scalar.
	ScalarBinary Scalar = "binary"
)

// Type describes a Skel type without references to compiler objects.
// Named declarations use SkelName; type parameters use Name. Collection element
// types and generic arguments are represented recursively.
type Type struct {
	Kind          TypeKind `json:"kind"`
	Nullable      bool     `json:"nullable,omitzero"`
	Scalar        Scalar   `json:"scalar,omitempty"`
	Name          string   `json:"name,omitempty"`
	SkelName      string   `json:"skelName,omitempty"`
	TypeArguments []*Type  `json:"typeArguments,omitempty"`
	Element       *Type    `json:"element,omitempty"`
	Key           *Type    `json:"key,omitempty"`
	Value         *Type    `json:"value,omitempty"`
}
