package descriptor

// ActorVia identifies an actor transport.
type ActorVia string

const (
	// ActorViaClient identifies the client transport.
	ActorViaClient ActorVia = "client"
	// ActorViaAgent identifies the agent transport.
	ActorViaAgent ActorVia = "agent"
	// ActorViaOpenAPI identifies the openapi transport.
	ActorViaOpenAPI ActorVia = "openapi"
)

// Actor describes an identity and its generated authentication and permission contracts.
type Actor struct {
	Name              string     `json:"name"`
	SkelName          string     `json:"skelName"`
	Description       string     `json:"description,omitempty"`
	Deprecated        bool       `json:"deprecated,omitzero"`
	DeprecatedReason  string     `json:"deprecatedReason,omitempty"`
	Hash              string     `json:"hash"`
	Vias              []ActorVia `json:"vias"`
	AuthEnabled       bool       `json:"authEnabled"`
	AuthCredential    *Data      `json:"authCredential,omitempty"`
	AuthInfo          *Data      `json:"authInfo,omitempty"`
	IdentifierField   string     `json:"identifierField,omitempty"`
	AuthService       *Service   `json:"authService,omitempty"`
	AuthMethod        *Method    `json:"authMethod,omitempty"`
	PermissionEnabled bool       `json:"permissionEnabled"`
	PermissionService *Service   `json:"permissionService,omitempty"`
	PermissionMethod  *Method    `json:"permissionMethod,omitempty"`
}

// ActorAudience identifies an actor and an optional transport allowed to access an entry point.
type ActorAudience struct {
	Name     string   `json:"name"`
	SkelName string   `json:"skelName"`
	Via      ActorVia `json:"via,omitempty"`
}
