package vineschema

type _ResourceSchema struct {
	Name             string
	SkelName         string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Hash             string
	Checks           []*_ResourceCheckSchema
	Actions          []*_ResourceActionSchema
	CheckService     *_ServiceSchema
}

type _ResourceActionSchema struct {
	Name             string
	PermissionCode   string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Checks           []*_ResourceCheckSchema
}

type _ResourceCheckSchema struct {
	Name             string
	Deprecated       bool
	DeprecatedReason string
	Method           *_MethodSchema
	Arguments        []*_MemberSchema
}
