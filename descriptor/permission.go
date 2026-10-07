package descriptor

// PermissionRequireMode identifies the operation carried by a permission expression.
type PermissionRequireMode string

const (
	// PermissionRequireModeCode requires a resource action code.
	PermissionRequireModeCode PermissionRequireMode = "code"
	// PermissionRequireModeCheck invokes a permission-check method.
	PermissionRequireModeCheck PermissionRequireMode = "check"
	// PermissionRequireModeAll requires every child expression.
	PermissionRequireModeAll PermissionRequireMode = "all"
	// PermissionRequireModeAny requires at least one child expression.
	PermissionRequireModeAny PermissionRequireMode = "any"
)

// PermissionRequire describes a permission requirement.
type PermissionRequire struct {
	Expression *PermissionExpression `json:"expression"`
}

// PermissionExpression is one node of a permission expression.
type PermissionExpression struct {
	Mode     PermissionRequireMode      `json:"mode"`
	Code     string                     `json:"code,omitempty"`
	Check    *PermissionCheckInvocation `json:"check,omitempty"`
	Children []*PermissionExpression    `json:"children,omitempty"`
}

// PermissionCheckInvocation describes a resolved permission-check call and its argument bindings.
type PermissionCheckInvocation struct {
	ResourceSkelName string `json:"resourceSkelName"`
	ActionName       string `json:"actionName"`
	CheckName        string `json:"checkName"`
	ServiceSkelName  string `json:"serviceSkelName"`
	MethodSkelName   string `json:"methodSkelName"`
	// CodeArgumentName names the parameter receiving the injected permission code.
	CodeArgumentName string                     `json:"codeArgumentName,omitempty"`
	Arguments        []*PermissionCheckArgument `json:"arguments,omitempty"`
}

// PermissionCheckArgument binds a request value to a permission-check argument.
type PermissionCheckArgument struct {
	Name     string `json:"name"`
	JsonPath string `json:"jsonPath"`
	Type     *Type  `json:"type"`
}
