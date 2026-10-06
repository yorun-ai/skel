package vineschema

type _MemberSchema struct {
	Name             string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Example          string
	Sensitive        bool
	Type             *_TypeSchema
}

type _TypeKind string

const (
	typeKindScalar        _TypeKind = "scalar"
	typeKindList          _TypeKind = "list"
	typeKindMap           _TypeKind = "map"
	typeKindEnum          _TypeKind = "enum"
	typeKindData          _TypeKind = "data"
	typeKindConfig        _TypeKind = "config"
	typeKindEvent         _TypeKind = "event"
	typeKindTypeParameter _TypeKind = "typeParameter"
)

type _Scalar string

const (
	scalarString        _Scalar = "string"
	scalarBool          _Scalar = "bool"
	scalarInt           _Scalar = "int"
	scalarLong          _Scalar = "long"
	scalarFloat         _Scalar = "float"
	scalarDouble        _Scalar = "double"
	scalarDecimal       _Scalar = "decimal"
	scalarJson          _Scalar = "json"
	scalarUuid          _Scalar = "uuid"
	scalarTimestamp     _Scalar = "timestamp"
	scalarDuration      _Scalar = "duration"
	scalarLocalDate     _Scalar = "localdate"
	scalarLocalTime     _Scalar = "localtime"
	scalarLocalDateTime _Scalar = "localdatetime"
	scalarBinary        _Scalar = "binary"
)

type _TypeSchema struct {
	Kind          _TypeKind
	Nullable      bool
	Scalar        _Scalar
	Name          string
	SkelName      string
	TypeArguments []*_TypeSchema
	Element       *_TypeSchema
	Key           *_TypeSchema
	Value         *_TypeSchema
}
