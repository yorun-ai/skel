package schema

// Resource describes a permission resource and its actions and checks.
type Resource struct {
	// Pos is the resource declaration's source position.
	Pos Position
	// Name is the resource's local name.
	Name string
	// SkelName is the resource's fully qualified Skel name.
	SkelName string
	// Hash is the resource's compatibility hash.
	Hash string
	// Description is the resource's documentation text.
	Description string
	// Deprecated reports whether the resource should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the resource is deprecated and what to use instead.
	DeprecatedReason string
	// Pub reports whether the resource belongs to the public contract.
	Pub bool
	// Checks lists checks declared directly on the resource.
	Checks []*ResourceCheck
	// Actions lists the resource's permission-bearing actions.
	Actions []*ResourceAction
	// CheckService is the language-defined service that implements resource checks.
	CheckService *Service
}

// ResourceAction describes one permission-bearing action on a resource.
type ResourceAction struct {
	// Pos is the action declaration's source position.
	Pos Position
	// Name is the action's local name.
	Name string
	// PermissionCode is the action's fully qualified permission code.
	PermissionCode string
	// Description is the action's documentation text.
	Description string
	// Deprecated reports whether the action should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the action is deprecated and what to use instead.
	DeprecatedReason string
	// Checks lists checks available for this action.
	Checks []*ResourceCheck
}

// ResourceCheck describes a named permission check and its generated method.
type ResourceCheck struct {
	// Name is the check's local name.
	Name string
	// Deprecated reports whether the check should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the check is deprecated and what to use instead.
	DeprecatedReason string
	// Method is the canonical method node in the owning Resource.CheckService.Methods.
	Method *Method
}
