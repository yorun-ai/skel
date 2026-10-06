package schema

// EnumSchema is the normalized body of an enum declaration.
type EnumSchema struct {
	Items []*EnumItem `json:"items"`
}

// EnumItem is one normalized enum item.
type EnumItem struct {
	Metadata
	Name string   `json:"name"`
	Pos  Position `json:"-"`
}

// DataSchema is the normalized body shared by data, config, and event declarations.
type DataSchema struct {
	Ext            bool            `json:"ext,omitempty"`
	Lifecycle      ConfigLifecycle `json:"lifecycle,omitempty"`
	Sensitive      bool            `json:"sensitive,omitempty"`
	TypeParameters []string        `json:"typeParameters,omitempty"`
	Members        []*Member       `json:"members"`
}

// Member is one normalized structured-data member.
type Member struct {
	Metadata
	Name      string   `json:"name"`
	Example   string   `json:"example,omitempty"`
	Sensitive bool     `json:"sensitive,omitempty"`
	Type      *Type    `json:"type"`
	Pos       Position `json:"-"`
}

// Type is a normalized type expression.
type Type struct {
	Kind      TypeKind `json:"kind"`
	Nullable  bool     `json:"nullable,omitempty"`
	Name      string   `json:"name,omitempty"`
	Arguments []*Type  `json:"arguments,omitempty"`
	Element   *Type    `json:"element,omitempty"`
	Key       *Type    `json:"key,omitempty"`
	Value     *Type    `json:"value,omitempty"`
}

// ActorSchema is the normalized body of an actor declaration.
type ActorSchema struct {
	IdentifierField string      `json:"identifierField,omitempty"`
	Vias            []*ActorVia `json:"vias"`
	AuthEnabled     bool        `json:"authEnabled,omitempty"`
	AuthCredential  *DataSchema `json:"authCredential,omitempty"`
	AuthInfo        *DataSchema `json:"authInfo,omitempty"`
	PermEnabled     bool        `json:"permEnabled,omitempty"`
}

// ActorVia is one actor transport capability.
type ActorVia struct {
	Name string   `json:"name"`
	Pos  Position `json:"-"`
}

// ResourceSchema is the normalized body of a resource declaration.
type ResourceSchema struct {
	Checks  []*ResourceCheck  `json:"checks,omitempty"`
	Actions []*ResourceAction `json:"actions"`
}

// ResourceAction is one normalized resource action.
type ResourceAction struct {
	Metadata
	Name           string           `json:"name"`
	PermissionCode string           `json:"permissionCode"`
	Checks         []*ResourceCheck `json:"checks,omitempty"`
	Pos            Position         `json:"-"`
}

// ResourceCheck is one normalized resource permission check.
type ResourceCheck struct {
	Metadata
	Name      string      `json:"name"`
	Arguments []*Argument `json:"arguments"`
	Pos       Position    `json:"-"`
}

// ServiceSchema is the normalized body of a service declaration.
type ServiceSchema struct {
	Audiences []*Audience  `json:"audiences"`
	Api       bool         `json:"api,omitempty"`
	Ext       bool         `json:"ext,omitempty"`
	Auth      AuthMode     `json:"auth"`
	Require   *Requirement `json:"require,omitempty"`
	Methods   []*Method    `json:"methods"`
}

// Audience is one normalized actor audience.
type Audience struct {
	Actor string   `json:"actor"`
	Via   string   `json:"via,omitempty"`
	Pos   Position `json:"-"`
}

// Method is one normalized service method.
type Method struct {
	Metadata
	Name               string       `json:"name"`
	SkelName           string       `json:"skelName"`
	Example            string       `json:"example,omitempty"`
	Auth               AuthMode     `json:"auth"`
	Require            *Requirement `json:"require,omitempty"`
	InputDescription   string       `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool         `json:"argumentsSensitive,omitempty"`
	OutputDescription  string       `json:"outputDescription,omitempty"`
	OutputExample      string       `json:"outputExample,omitempty"`
	ResultSensitive    bool         `json:"resultSensitive,omitempty"`
	Arguments          []*Argument  `json:"arguments"`
	Result             *Type        `json:"result,omitempty"`
	Pos                Position     `json:"-"`
}

// Argument is one normalized method, check, or trigger argument.
type Argument struct {
	Metadata
	Name      string   `json:"name"`
	Example   string   `json:"example,omitempty"`
	Sensitive bool     `json:"sensitive,omitempty"`
	Type      *Type    `json:"type"`
	Pos       Position `json:"-"`
}

// Requirement is one node in a normalized permission expression.
type Requirement struct {
	Mode     RequirementMode   `json:"mode"`
	Code     string            `json:"code,omitempty"`
	Check    *RequirementCheck `json:"check,omitempty"`
	Children []*Requirement    `json:"children,omitempty"`
}

// RequirementCheck is one normalized permission check invocation.
type RequirementCheck struct {
	Resource  string                      `json:"resource"`
	Action    string                      `json:"action,omitempty"`
	Check     string                      `json:"check"`
	Arguments []*RequirementCheckArgument `json:"arguments,omitempty"`
}

// RequirementCheckArgument binds one permission-check argument.
type RequirementCheckArgument struct {
	Name     string `json:"name"`
	JSONPath string `json:"jsonPath"`
	Type     *Type  `json:"type"`
}

// WebSchema is the normalized body of a web declaration.
type WebSchema struct {
	Audiences []*Audience `json:"audiences"`
	Auth      AuthMode    `json:"auth,omitempty"`
	MountPath string      `json:"mountPath,omitempty"`
}

// TaskSchema is the normalized body of a task declaration.
type TaskSchema struct {
	Triggers []*Trigger `json:"triggers"`
}

// Trigger is one normalized task trigger.
type Trigger struct {
	Metadata
	Name               string      `json:"name"`
	SkelName           string      `json:"skelName"`
	InputDescription   string      `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool        `json:"argumentsSensitive,omitempty"`
	Arguments          []*Argument `json:"arguments"`
	Pos                Position    `json:"-"`
}
