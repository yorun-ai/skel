package vineschema

type _PermRequireMode string

const (
	permRequireModeCode  _PermRequireMode = "code"
	permRequireModeCheck _PermRequireMode = "check"
	permRequireModeAll   _PermRequireMode = "all"
	permRequireModeAny   _PermRequireMode = "any"
)

type _PermRequire struct {
	Expr *_PermExpr
}

type _PermExpr struct {
	Mode     _PermRequireMode
	Code     string
	Check    *_PermCheckInvocation
	Children []*_PermExpr
}

type _PermCheckInvocation struct {
	ResourceSkelName string
	ActionName       string
	CheckName        string
	ServiceSkelName  string
	MethodSkelName   string
	CodeArgumentName string
	Arguments        []*_PermCheckArgument
}

type _PermCheckArgument struct {
	Name     string
	JsonPath string
	Type     *_TypeSchema
}
