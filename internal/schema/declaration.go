package schema

import "go.yorun.ai/skelc/internal/model"

type EnumSchema struct {
	Items []*EnumItem `json:"items"`
}

type EnumItem struct {
	Metadata
	Name string         `json:"name"`
	Pos  model.Position `json:"-"`
}

type DataSchema struct {
	Lifecycle      ConfigLifecycle `json:"lifecycle,omitempty"`
	Sensitive      bool            `json:"sensitive,omitempty"`
	TypeParameters []string        `json:"typeParameters,omitempty"`
	Members        []*Member       `json:"members"`
}

type Member struct {
	Metadata
	Name      string         `json:"name"`
	Example   string         `json:"example,omitempty"`
	Sensitive bool           `json:"sensitive,omitempty"`
	Type      *Type          `json:"type"`
	Pos       model.Position `json:"-"`
}

type Type struct {
	Kind      TypeKind `json:"kind"`
	Nullable  bool     `json:"nullable,omitempty"`
	Name      string   `json:"name,omitempty"`
	Arguments []*Type  `json:"arguments,omitempty"`
	Element   *Type    `json:"element,omitempty"`
	Key       *Type    `json:"key,omitempty"`
	Value     *Type    `json:"value,omitempty"`
}

type ActorSchema struct {
	IdentifierField string      `json:"identifierField,omitempty"`
	Vias            []*ActorVia `json:"vias"`
	AuthEnabled     bool        `json:"authEnabled,omitempty"`
	AuthCredential  *DataSchema `json:"authCredential,omitempty"`
	AuthInfo        *DataSchema `json:"authInfo,omitempty"`
	PermEnabled     bool        `json:"permEnabled,omitempty"`
}

type ActorVia struct {
	Name string         `json:"name"`
	Pos  model.Position `json:"-"`
}

type ResourceSchema struct {
	Checks  []*ResourceCheck  `json:"checks,omitempty"`
	Actions []*ResourceAction `json:"actions"`
}

type ResourceAction struct {
	Metadata
	Name           string           `json:"name"`
	PermissionCode string           `json:"permissionCode"`
	Checks         []*ResourceCheck `json:"checks,omitempty"`
	Pos            model.Position   `json:"-"`
}

type ResourceCheck struct {
	Metadata
	Name      string         `json:"name"`
	Arguments []*Argument    `json:"arguments"`
	Pos       model.Position `json:"-"`
}

type ServiceSchema struct {
	Audiences []*Audience  `json:"audiences"`
	Api       bool         `json:"api,omitempty"`
	Open      bool         `json:"open,omitempty"`
	Auth      AuthMode     `json:"auth"`
	Require   *Requirement `json:"require,omitempty"`
	Methods   []*Method    `json:"methods"`
}

type Audience struct {
	Actor string         `json:"actor"`
	Via   string         `json:"via,omitempty"`
	Pos   model.Position `json:"-"`
}

type Method struct {
	Metadata
	Name               string         `json:"name"`
	SkelName           string         `json:"skelName"`
	Example            string         `json:"example,omitempty"`
	Auth               AuthMode       `json:"auth"`
	Require            *Requirement   `json:"require,omitempty"`
	InputDescription   string         `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool           `json:"argumentsSensitive,omitempty"`
	OutputDescription  string         `json:"outputDescription,omitempty"`
	OutputExample      string         `json:"outputExample,omitempty"`
	ResultSensitive    bool           `json:"resultSensitive,omitempty"`
	Arguments          []*Argument    `json:"arguments"`
	Result             *Type          `json:"result,omitempty"`
	Pos                model.Position `json:"-"`
}

type Argument struct {
	Metadata
	Name      string         `json:"name"`
	Example   string         `json:"example,omitempty"`
	Sensitive bool           `json:"sensitive,omitempty"`
	Type      *Type          `json:"type"`
	Pos       model.Position `json:"-"`
}

type Requirement struct {
	Mode     RequirementMode   `json:"mode"`
	Code     string            `json:"code,omitempty"`
	Check    *RequirementCheck `json:"check,omitempty"`
	Children []*Requirement    `json:"children,omitempty"`
}

type RequirementCheck struct {
	Resource  string                      `json:"resource"`
	Action    string                      `json:"action,omitempty"`
	Check     string                      `json:"check"`
	Arguments []*RequirementCheckArgument `json:"arguments,omitempty"`
}

type RequirementCheckArgument struct {
	Name     string `json:"name"`
	JSONPath string `json:"jsonPath"`
	Type     *Type  `json:"type"`
}

type WebSchema struct {
	Audiences []*Audience `json:"audiences"`
	MountPath string      `json:"mountPath,omitempty"`
}

type TaskSchema struct {
	Triggers []*Trigger `json:"triggers"`
}

type Trigger struct {
	Metadata
	Name               string         `json:"name"`
	SkelName           string         `json:"skelName"`
	InputDescription   string         `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool           `json:"argumentsSensitive,omitempty"`
	Arguments          []*Argument    `json:"arguments"`
	Pos                model.Position `json:"-"`
}
