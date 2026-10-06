package descriptor

// Resource describes permission actions, checks, and their generated service.
type Resource struct {
	Name             string            `json:"name"`
	SkelName         string            `json:"skelName"`
	Description      string            `json:"description,omitempty"`
	Deprecated       bool              `json:"deprecated,omitzero"`
	DeprecatedReason string            `json:"deprecatedReason,omitempty"`
	Hash             string            `json:"hash"`
	Checks           []*ResourceCheck  `json:"checks,omitempty"`
	Actions          []*ResourceAction `json:"actions"`
	CheckService     *Service          `json:"checkService,omitempty"`
}

// ResourceAction describes a permission-bearing resource action.
type ResourceAction struct {
	Name             string           `json:"name"`
	PermissionCode   string           `json:"permissionCode"`
	Description      string           `json:"description,omitempty"`
	Deprecated       bool             `json:"deprecated,omitzero"`
	DeprecatedReason string           `json:"deprecatedReason,omitempty"`
	Checks           []*ResourceCheck `json:"checks,omitempty"`
}

// ResourceCheck describes a resource check and its callable method.
type ResourceCheck struct {
	Name             string    `json:"name"`
	Deprecated       bool      `json:"deprecated,omitzero"`
	DeprecatedReason string    `json:"deprecatedReason,omitempty"`
	Method           *Method   `json:"method"`
	Arguments        []*Member `json:"arguments,omitempty"`
}
