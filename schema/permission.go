package schema

// PermissionRequireMode identifies the form of a permission expression.
type PermissionRequireMode string

const (
	// PermissionRequireModeCode requires one resource action code.
	PermissionRequireModeCode PermissionRequireMode = "code"
	// PermissionRequireModeCheck invokes a resource permission check.
	PermissionRequireModeCheck PermissionRequireMode = "check"
	// PermissionRequireModeAll requires every child expression to pass.
	PermissionRequireModeAll PermissionRequireMode = "all"
	// PermissionRequireModeAny requires at least one child expression to pass.
	PermissionRequireModeAny PermissionRequireMode = "any"
)

// PermissionRequire describes a source or normalized require clause.
type PermissionRequire struct {
	// Expression is the root of the permission expression tree.
	Expression *PermissionExpression
}

// PermissionExpression is one node in a source or normalized permission tree.
type PermissionExpression struct {
	// Mode identifies which of Code, Check, or Children carries this node's value.
	// An empty Mode with Check retains an unresolved resource action term; CheckName
	// may be empty. Resolution expands it into a code and, if present, a check.
	Mode PermissionRequireMode
	// Code is the fully qualified resource action code for code expressions.
	Code string
	// Check describes the invocation for check expressions.
	Check *PermissionCheckInvocation
	// Children contains operands for all and any expressions.
	Children []*PermissionExpression
}

// PermissionCheckInvocation describes a resource action term and its optional
// permission check. Resolution fills in the generated method bindings.
type PermissionCheckInvocation struct {
	// ResourceSkelName is the resource reference, fully qualified after resolution.
	ResourceSkelName string
	// ActionName is the resource action's local name.
	ActionName string
	// CheckName is the resource check's local name.
	CheckName string
	// ServiceSkelName is the generated check service's fully qualified Skel name.
	ServiceSkelName string
	// MethodSkelName is the generated check method's Skel name.
	MethodSkelName string
	// CodeArgumentName names the parameter receiving the injected permission code.
	CodeArgumentName string
	// Arguments lists business arguments in call order.
	Arguments []*PermissionCheckArgument
}

// PermissionCheckArgument describes one permission-check argument.
type PermissionCheckArgument struct {
	// Name is the target method argument name, populated by resolution.
	Name string
	// JsonPath selects the value supplied to the argument.
	JsonPath string
	// Type is the argument's semantic type, populated by resolution.
	Type *Type
}
