package vineschema

type _EnumSchema struct {
	Name             string
	SkelName         string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Hash             string
	Items            []*_EnumItemSchema
}

type _EnumItemSchema struct {
	Name             string
	Description      string
	Deprecated       bool
	DeprecatedReason string
}

type _DataSchema struct {
	Name             string
	SkelName         string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Hash             string
	Sensitive        bool
	TypeParameters   []string
	Members          []*_MemberSchema
}

type _ConfigSchema struct {
	Name             string
	SkelName         string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Hash             string
	Pub              bool
	Sensitive        bool
	Lifecycle        string
	Members          []*_MemberSchema
}

type _EventSchema struct {
	Ext              bool
	Name             string
	SkelName         string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Hash             string
	Pub              bool
	Sensitive        bool
	Members          []*_MemberSchema
}
