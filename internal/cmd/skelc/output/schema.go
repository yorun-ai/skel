package output

import "go.yorun.ai/skel/schema"

// SchemaMetadata contains common declaration and member documentation metadata.
type SchemaMetadata struct {
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitempty"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
}

// SchemaDeclaration is one complete serialized declaration emitted by schema get.
type SchemaDeclaration struct {
	SchemaMetadata
	// Pub reports whether the source declaration uses the pub modifier.
	Pub      bool                   `json:"pub"`
	Name     string                 `json:"name"`
	Kind     schema.DeclarationType `json:"type"`
	SkelName string                 `json:"skelName"`
	Enum     *SchemaEnum            `json:"enum,omitempty"`
	Data     *SchemaData            `json:"data,omitempty"`
	Actor    *SchemaActor           `json:"actor,omitempty"`
	Resource *SchemaResource        `json:"resource,omitempty"`
	Service  *SchemaService         `json:"service,omitempty"`
	Web      *SchemaWeb             `json:"web,omitempty"`
	Task     *SchemaTask            `json:"task,omitempty"`
}

// SchemaEntry is one declaration summary emitted by schema list.
type SchemaEntry struct {
	// Pub reports whether the source declaration uses the pub modifier.
	Pub      bool                   `json:"pub"`
	Name     string                 `json:"name"`
	Kind     schema.DeclarationType `json:"type"`
	SkelName string                 `json:"skelName"`
}

// SchemaEnum is the serialized body of an enum declaration.
type SchemaEnum struct {
	Items []*SchemaEnumItem `json:"items"`
}

// SchemaEnumItem is one serialized enum item.
type SchemaEnumItem struct {
	SchemaMetadata
	Name string `json:"name"`
}

// SchemaData is the serialized body shared by data, config, and event declarations.
type SchemaData struct {
	Ext            bool            `json:"ext,omitempty"`
	Lifecycle      string          `json:"lifecycle,omitempty"`
	Sensitive      bool            `json:"sensitive,omitempty"`
	TypeParameters []string        `json:"typeParameters,omitempty"`
	Members        []*SchemaMember `json:"members"`
}

// SchemaMember is one serialized structured-data member.
type SchemaMember struct {
	SchemaMetadata
	Name      string      `json:"name"`
	Example   string      `json:"example,omitempty"`
	Sensitive bool        `json:"sensitive,omitempty"`
	Type      *SchemaType `json:"type"`
}

// SchemaType is a serialized type expression.
type SchemaType struct {
	Kind      string        `json:"kind"`
	Nullable  bool          `json:"nullable,omitempty"`
	Name      string        `json:"name,omitempty"`
	Arguments []*SchemaType `json:"arguments,omitempty"`
	Element   *SchemaType   `json:"element,omitempty"`
	Key       *SchemaType   `json:"key,omitempty"`
	Value     *SchemaType   `json:"value,omitempty"`
}

// SchemaActor is the serialized body of an actor declaration.
type SchemaActor struct {
	Vias       []*SchemaActorVia      `json:"vias"`
	Auth       *SchemaActorAuth       `json:"auth,omitzero"`
	Permission *SchemaActorPermission `json:"permission,omitzero"`
}

// SchemaActorAuth is the authentication declaration without derived services.
type SchemaActorAuth struct {
	Credential      *SchemaData `json:"credential"`
	Info            *SchemaData `json:"info"`
	IdentifierField string      `json:"identifierField,omitempty"`
}

// SchemaActorPermission records the presence of an actor permission declaration.
type SchemaActorPermission struct{}

// SchemaActorVia is one actor transport capability.
type SchemaActorVia struct {
	Name string `json:"name"`
}

// SchemaResource is the serialized body of a resource declaration.
type SchemaResource struct {
	Checks  []*SchemaResourceCheck  `json:"checks,omitempty"`
	Actions []*SchemaResourceAction `json:"actions"`
}

// SchemaResourceAction is one serialized resource action.
type SchemaResourceAction struct {
	SchemaMetadata
	Name           string                 `json:"name"`
	PermissionCode string                 `json:"permissionCode"`
	Checks         []*SchemaResourceCheck `json:"checks,omitempty"`
}

// SchemaResourceCheck is one serialized resource permission check.
type SchemaResourceCheck struct {
	SchemaMetadata
	Name      string            `json:"name"`
	Arguments []*SchemaArgument `json:"arguments"`
}

// SchemaService is the serialized body of a service declaration.
type SchemaService struct {
	Audiences []*SchemaAudience  `json:"audiences"`
	Api       bool               `json:"api,omitempty"`
	Ext       bool               `json:"ext,omitempty"`
	AuthMode  string             `json:"authMode"`
	Require   *SchemaRequirement `json:"require,omitempty"`
	Methods   []*SchemaMethod    `json:"methods"`
}

// SchemaAudience is one serialized actor audience.
type SchemaAudience struct {
	Actor string `json:"actor"`
	Via   string `json:"via,omitempty"`
}

// SchemaMethod is one serialized service method.
type SchemaMethod struct {
	SchemaMetadata
	Name               string             `json:"name"`
	SkelName           string             `json:"skelName"`
	Example            string             `json:"example,omitempty"`
	AuthMode           string             `json:"authMode"`
	Require            *SchemaRequirement `json:"require,omitempty"`
	EffectiveAuthMode  string             `json:"effectiveAuthMode"`
	EffectiveRequire   *SchemaRequirement `json:"effectiveRequire,omitempty"`
	InputDescription   string             `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool               `json:"argumentsSensitive,omitempty"`
	OutputDescription  string             `json:"outputDescription,omitempty"`
	OutputExample      string             `json:"outputExample,omitempty"`
	ResultSensitive    bool               `json:"resultSensitive,omitempty"`
	Arguments          []*SchemaArgument  `json:"arguments"`
	Result             *SchemaType        `json:"result,omitempty"`
}

// SchemaArgument is one serialized method, check, or trigger argument.
type SchemaArgument struct {
	SchemaMetadata
	Name      string      `json:"name"`
	Example   string      `json:"example,omitempty"`
	Sensitive bool        `json:"sensitive,omitempty"`
	Type      *SchemaType `json:"type"`
}

// SchemaRequirement is one node in a serialized permission expression.
type SchemaRequirement struct {
	Mode     string                  `json:"mode"`
	Code     string                  `json:"code,omitempty"`
	Check    *SchemaRequirementCheck `json:"check,omitempty"`
	Children []*SchemaRequirement    `json:"children,omitempty"`
}

// SchemaRequirementCheck is one serialized permission check invocation.
type SchemaRequirementCheck struct {
	Resource  string                            `json:"resource"`
	Action    string                            `json:"action,omitempty"`
	Check     string                            `json:"check"`
	Arguments []*SchemaRequirementCheckArgument `json:"arguments,omitempty"`
}

// SchemaRequirementCheckArgument binds one permission-check argument.
type SchemaRequirementCheckArgument struct {
	Name     string      `json:"name"`
	JSONPath string      `json:"jsonPath"`
	Type     *SchemaType `json:"type"`
}

// SchemaWeb is the serialized body of a web declaration.
type SchemaWeb struct {
	Audiences []*SchemaAudience `json:"audiences"`
	AuthMode  string            `json:"authMode,omitempty"`
	MountPath string            `json:"mountPath,omitempty"`
}

// SchemaTask is the serialized body of a task declaration.
type SchemaTask struct {
	Triggers []*SchemaTrigger `json:"triggers"`
}

// SchemaTrigger is one serialized task trigger.
type SchemaTrigger struct {
	SchemaMetadata
	Name               string            `json:"name"`
	SkelName           string            `json:"skelName"`
	InputDescription   string            `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool              `json:"argumentsSensitive,omitempty"`
	Arguments          []*SchemaArgument `json:"arguments"`
}
