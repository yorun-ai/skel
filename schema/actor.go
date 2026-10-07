package schema

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
	// Auth describes authentication; nil means no auth section was declared.
	Auth *ActorAuth
	// Permission describes permission support; nil means no permission section was declared.
	Permission *ActorPermission
}

// ActorAuth describes an actor's authentication declaration and derived service.
type ActorAuth struct {
	// Pos is the auth section's source position.
	Pos Position
	// Credential is the language-defined authentication credential data schema.
	Credential *Data
	// Info is the language-defined authenticated-actor information data schema.
	Info *Data
	// IdentifierField names the optional identity field in Info.
	IdentifierField string
	// Service is the language-defined authentication service.
	Service *Service
	// Method is the canonical authentication method node in Service.Methods.
	Method *Method
}

// ActorPermission describes an actor's permission declaration and derived service.
type ActorPermission struct {
	// Pos is the permission section's source position.
	Pos Position
	// Service is the language-defined permission service.
	Service *Service
	// Method is the canonical permission-checking method node in Service.Methods.
	Method *Method
}

// ActorVia describes one transport declared by an actor.
type ActorVia struct {
	// Pos is the transport declaration's source position.
	Pos Position
	// Name is the transport name and corresponds to an [ActorViaKind].
	Name string
}
