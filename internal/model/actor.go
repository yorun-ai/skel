package model

// ActorViaKind identifies a transport through which an actor can access a
// domain.
type ActorViaKind string

const (
	// ActorViaClient represents a client application transport.
	ActorViaClient ActorViaKind = "client"
	// ActorViaAgent represents an agent transport.
	ActorViaAgent ActorViaKind = "agent"
	// ActorViaOpenAPI represents an OpenAPI transport.
	ActorViaOpenAPI ActorViaKind = "openapi"
)

// Actor describes a caller identity and the transports and authorization
// facilities available to it.
type Actor struct {
	// Pos is the actor declaration's source position.
	Pos Position
	// Name is the actor's local name.
	Name string
	// SkelName is the actor's fully qualified Skel name.
	SkelName string
	// Hash is the actor's compatibility hash.
	Hash string
	// Description is the actor's documentation text.
	Description string
	// Deprecated reports whether the actor should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the actor is deprecated and what to use instead.
	DeprecatedReason string
	// Pub reports whether the actor belongs to the public contract.
	Pub bool
	// Vias lists the transports declared by the actor.
	Vias []*ActorVia
	// AuthEnabled reports whether the actor declares authentication.
	AuthEnabled bool
	// AuthCredential is the language-defined authentication credential data model.
	AuthCredential *Data
	// AuthInfo is the language-defined authenticated-actor information data model.
	AuthInfo *Data
	// IdentifierField names the optional identity field in AuthInfo.
	IdentifierField string
	// AuthService is the language-defined authentication service.
	AuthService *Service
	// AuthMethod is the authentication method in AuthService.
	AuthMethod *Method
	// PermEnabled reports whether the actor declares permission support.
	PermEnabled bool
	// PermService is the language-defined permission service.
	PermService *Service
	// PermMethod is the permission-checking method in PermService.
	PermMethod *Method
}

// ActorVia describes one transport declared by an actor.
type ActorVia struct {
	// Pos is the transport declaration's source position.
	Pos Position
	// Name is the transport name and corresponds to an [ActorViaKind].
	Name string
}
