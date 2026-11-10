package schema

// Actor describes a caller identity and its authentication and authorization contracts.
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
	// Result is the derived authentication result envelope.
	Result *Data
	// ImpersonatorInfo is the optional mutually exclusive operator-info data.
	ImpersonatorInfo *Data
	// Impersonators lists the explicitly supported operator actors.
	Impersonators []*ActorImpersonator
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

// ActorImpersonator associates an operator actor with its generated info field.
type ActorImpersonator struct {
	// Pos is the source position of the actor reference.
	Pos Position
	// Actor is the source actor reference, including an import qualifier when needed.
	Actor string
	// ActorSkelName is the resolved fully qualified actor name.
	ActorSkelName string
	// InfoFieldName is the member name in the derived impersonator-info data.
	InfoFieldName string
}
