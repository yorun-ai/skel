package hasher

import "go.yorun.ai/skelc/internal/model"

type _ActorRefHashValue struct {
	Name     string `json:"name"`
	SkelName string `json:"skelName"`
	Hash     string `json:"hash,omitempty"`
	Via      string `json:"via,omitempty"`
}

type _EnumItemHashValue struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitempty"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
}

type _MemberHashValue struct {
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	Deprecated       bool            `json:"deprecated,omitempty"`
	DeprecatedReason string          `json:"deprecatedReason,omitempty"`
	Example          string          `json:"example,omitempty"`
	Sensitive        bool            `json:"sensitive,omitempty"`
	Type             *_TypeHashValue `json:"type,omitempty"`
}

type _TypeHashValue struct {
	Kind          string            `json:"kind"`
	Nullable      bool              `json:"nullable,omitempty"`
	Scalar        string            `json:"scalar,omitempty"`
	Name          string            `json:"name,omitempty"`
	SkelName      string            `json:"skelName,omitempty"`
	Hash          string            `json:"hash,omitempty"`
	TypeArguments []*_TypeHashValue `json:"typeArguments,omitempty"`
	Element       *_TypeHashValue   `json:"element,omitempty"`
	Key           *_TypeHashValue   `json:"key,omitempty"`
	Value         *_TypeHashValue   `json:"value,omitempty"`
}

type _EnumHashValue struct {
	Name             string                `json:"name"`
	SkelName         string                `json:"skelName"`
	Description      string                `json:"description,omitempty"`
	Deprecated       bool                  `json:"deprecated,omitempty"`
	DeprecatedReason string                `json:"deprecatedReason,omitempty"`
	Items            []*_EnumItemHashValue `json:"items,omitempty"`
}

type _DataHashValue struct {
	Name             string              `json:"name"`
	SkelName         string              `json:"skelName"`
	Description      string              `json:"description,omitempty"`
	Deprecated       bool                `json:"deprecated,omitempty"`
	DeprecatedReason string              `json:"deprecatedReason,omitempty"`
	Kind             model.DataKind      `json:"kind,omitempty"`
	Pub              bool                `json:"pub,omitempty"`
	Sensitive        bool                `json:"sensitive,omitempty"`
	Lifecycle        string              `json:"lifecycle,omitempty"`
	TypeParameters   []string            `json:"typeParameters,omitempty"`
	Members          []*_MemberHashValue `json:"members,omitempty"`
}

type _WebHashValue struct {
	MountPath        string                `json:"mountPath,omitempty"`
	Name             string                `json:"name"`
	SkelName         string                `json:"skelName"`
	Description      string                `json:"description,omitempty"`
	Deprecated       bool                  `json:"deprecated,omitempty"`
	DeprecatedReason string                `json:"deprecatedReason,omitempty"`
	Actors           []*_ActorRefHashValue `json:"actors,omitempty"`
}

type _ActorHashValue struct {
	IdentifierField    string   `json:"identifierField,omitempty"`
	Name               string   `json:"name"`
	SkelName           string   `json:"skelName"`
	Description        string   `json:"description,omitempty"`
	Deprecated         bool     `json:"deprecated,omitempty"`
	DeprecatedReason   string   `json:"deprecatedReason,omitempty"`
	Vias               []string `json:"vias,omitempty"`
	AuthEnabled        bool     `json:"authEnabled,omitempty"`
	AuthCredential     string   `json:"authCredential,omitempty"`
	AuthCredentialHash string   `json:"authCredentialHash,omitempty"`
	AuthInfo           string   `json:"authInfo,omitempty"`
	AuthInfoHash       string   `json:"authInfoHash,omitempty"`
	AuthMethod         string   `json:"authMethod,omitempty"`
	AuthMethodHash     string   `json:"authMethodHash,omitempty"`
	PermEnabled        bool     `json:"permEnabled,omitempty"`
	PermMethod         string   `json:"permMethod,omitempty"`
	PermMethodHash     string   `json:"permMethodHash,omitempty"`
}

type _ResourceHashValue struct {
	Name             string             `json:"name"`
	SkelName         string             `json:"skelName"`
	Description      string             `json:"description,omitempty"`
	Deprecated       bool               `json:"deprecated,omitempty"`
	DeprecatedReason string             `json:"deprecatedReason,omitempty"`
	Pub              bool               `json:"pub,omitempty"`
	Checks           []*_ResourceCheck  `json:"checks,omitempty"`
	Actions          []*_ResourceAction `json:"actions,omitempty"`
	CheckService     string             `json:"checkService,omitempty"`
	CheckServiceHash string             `json:"checkServiceHash,omitempty"`
}

type _ResourceAction struct {
	Name             string            `json:"name"`
	PermissionCode   string            `json:"permissionCode"`
	Description      string            `json:"description,omitempty"`
	Deprecated       bool              `json:"deprecated,omitempty"`
	DeprecatedReason string            `json:"deprecatedReason,omitempty"`
	Checks           []*_ResourceCheck `json:"checks,omitempty"`
}

type _ResourceCheck struct {
	Name             string              `json:"name"`
	Deprecated       bool                `json:"deprecated,omitempty"`
	DeprecatedReason string              `json:"deprecatedReason,omitempty"`
	Method           string              `json:"method"`
	MethodHash       string              `json:"methodHash"`
	Arguments        []*_MemberHashValue `json:"arguments,omitempty"`
}

type _ServiceHashValue struct {
	Ext              bool                  `json:"ext,omitempty"`
	Name             string                `json:"name"`
	SkelName         string                `json:"skelName"`
	Description      string                `json:"description,omitempty"`
	Deprecated       bool                  `json:"deprecated,omitempty"`
	DeprecatedReason string                `json:"deprecatedReason,omitempty"`
	Pub              bool                  `json:"pub,omitempty"`
	Api              bool                  `json:"api,omitempty"`
	Actors           []*_ActorRefHashValue `json:"actors,omitempty"`
	Auth             string                `json:"auth,omitempty"`
	Require          *_RequireHashValue    `json:"require,omitempty"`
	Methods          []*_NamedValue        `json:"methods,omitempty"`
}

type _RequireHashValue struct {
	Expr *_RequireExprHashValue `json:"expr"`
}

type _RequireExprHashValue struct {
	Mode     string                   `json:"mode"`
	Code     string                   `json:"code,omitempty"`
	Check    *_RequireCheckHashValue  `json:"check,omitempty"`
	Children []*_RequireExprHashValue `json:"children,omitempty"`
}

type _RequireCheckHashValue struct {
	ResourceSkelName string                   `json:"resourceSkelName"`
	ActionName       string                   `json:"actionName"`
	CheckName        string                   `json:"checkName"`
	Arguments        []*_RequireCheckArgument `json:"arguments,omitempty"`
}

type _RequireCheckArgument struct {
	Name     string          `json:"name"`
	JsonPath string          `json:"jsonPath"`
	Type     *_TypeHashValue `json:"type"`
}

type _MethodHashValue struct {
	Name               string              `json:"name"`
	SkelName           string              `json:"skelName"`
	Description        string              `json:"description,omitempty"`
	Deprecated         bool                `json:"deprecated,omitempty"`
	DeprecatedReason   string              `json:"deprecatedReason,omitempty"`
	Example            string              `json:"example,omitempty"`
	Auth               string              `json:"auth,omitempty"`
	Require            *_RequireHashValue  `json:"require,omitempty"`
	InputDescription   string              `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool                `json:"argumentsSensitive,omitempty"`
	OutputDescription  string              `json:"outputDescription,omitempty"`
	OutputExample      string              `json:"outputExample,omitempty"`
	ResultSensitive    bool                `json:"resultSensitive,omitempty"`
	Arguments          []*_MemberHashValue `json:"arguments,omitempty"`
	ResultType         *_TypeHashValue     `json:"resultType,omitempty"`
}

type _TaskHashValue struct {
	Name             string         `json:"name"`
	SkelName         string         `json:"skelName"`
	Description      string         `json:"description,omitempty"`
	Deprecated       bool           `json:"deprecated,omitempty"`
	DeprecatedReason string         `json:"deprecatedReason,omitempty"`
	Triggers         []*_NamedValue `json:"triggers,omitempty"`
}

type _TriggerHashValue struct {
	Name               string              `json:"name"`
	SkelName           string              `json:"skelName"`
	Description        string              `json:"description,omitempty"`
	Deprecated         bool                `json:"deprecated,omitempty"`
	DeprecatedReason   string              `json:"deprecatedReason,omitempty"`
	InputDescription   string              `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool                `json:"argumentsSensitive,omitempty"`
	Arguments          []*_MemberHashValue `json:"arguments,omitempty"`
}

type _DomainHashValue struct {
	Domain      string         `json:"domain"`
	Description string         `json:"description,omitempty"`
	Enums       []*_NamedValue `json:"enums,omitempty"`
	Data        []*_NamedValue `json:"data,omitempty"`
	Configs     []*_NamedValue `json:"configs,omitempty"`
	Webs        []*_NamedValue `json:"webs,omitempty"`
	Events      []*_NamedValue `json:"events,omitempty"`
	Actors      []*_NamedValue `json:"actors,omitempty"`
	Resources   []*_NamedValue `json:"resources,omitempty"`
	Services    []*_NamedValue `json:"services,omitempty"`
	Tasks       []*_NamedValue `json:"tasks,omitempty"`
}
