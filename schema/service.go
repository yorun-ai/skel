package schema

// AuthMode controls whether authentication is required for a service or method.
type AuthMode string

const (
	// AuthModeUnset inherits authentication behavior from the enclosing context.
	AuthModeUnset AuthMode = "unset"
	// AuthModeRequired requires valid credentials.
	AuthModeRequired AuthMode = "required"
	// AuthModeOptional allows anonymous callers and authenticates supplied credentials.
	AuthModeOptional AuthMode = "optional"
	// AuthModeAnonymous allows only anonymous callers and rejects supplied invalid credentials.
	AuthModeAnonymous AuthMode = "anonymous"
	// AuthModeOff bypasses portal authentication for web declarations.
	AuthModeOff AuthMode = "off"

	// AuthModeAuth requires an authenticated actor.
	AuthModeAuth AuthMode = "auth"
	// AuthModeNoAuth explicitly allows unauthenticated access.
	AuthModeNoAuth AuthMode = "noauth"
)

// Service describes a callable service declaration.
type Service struct {
	// Pos is the service declaration's source position.
	Pos Position
	// Name is the service's local name.
	Name string
	// SkelName is the service's fully qualified Skel name.
	SkelName string
	// Hash is the service's compatibility hash.
	Hash string
	// Description is the service's documentation text.
	Description string
	// Deprecated reports whether the service should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the service is deprecated and what to use instead.
	DeprecatedReason string
	// Pub reports whether the service uses the pub modifier.
	Pub bool
	// Api restricts invocation to the portal client entry path.
	Api bool
	// Ext exports the server contract for implementation by other domains.
	Ext bool
	// Audiences lists actors allowed to call the service.
	Audiences []*ActorAudience
	// Auth is the service-level authentication mode.
	Auth AuthMode
	// AuthPos is the auth marker source position, or zero when omitted.
	AuthPos Position
	// Require is the service-level permission requirement.
	Require *PermissionRequire
	// Methods lists methods in source order.
	Methods []*Method
}

// ActorAudience identifies an actor and transport allowed to access a service
// or web entry point.
type ActorAudience struct {
	// Actor is the referenced actor name as resolved from source.
	Actor string
	// Via is the required actor transport, or empty when no transport is selected.
	Via string
	// Pos is the audience declaration's source position.
	Pos Position
}

// Method describes one callable service method.
type Method struct {
	// Pos is the method declaration's source position.
	Pos Position
	// Name is the method's normalized local name.
	Name string
	// SkelName is the method name as represented in Skel metadata.
	SkelName string
	// Hash is the method's compatibility hash.
	Hash string
	// Description is the method's documentation text.
	Description string
	// Deprecated reports whether the method should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the method is deprecated and what to use instead.
	DeprecatedReason string
	// Example is the method's example text.
	Example string
	// Auth is the method-level authentication mode.
	Auth AuthMode
	// AuthPos is the auth marker source position, or zero when omitted.
	AuthPos Position
	// Require is the method-level permission requirement.
	Require *PermissionRequire
	// Arguments lists input arguments in source order.
	Arguments []*Argument
	// ArgumentsData is the language-defined data schema representing method arguments.
	ArgumentsData *Data
	// InputDescription documents the method input as a whole.
	InputDescription string
	// ArgumentsSensitive reports whether the method input is sensitive as a whole.
	ArgumentsSensitive bool
	// OutputDescription documents the method result.
	OutputDescription string
	// OutputExample is the method result's example text.
	OutputExample string
	// ResultSensitive reports whether the method output is sensitive as a whole.
	ResultSensitive bool
	// ResultType is the resolved result type, or nil for a method with no result.
	ResultType *Type
}

// ArgumentSource identifies who supplies an argument's value.
type ArgumentSource int

const (
	// ArgumentSourceDeclared identifies an argument supplied by the caller.
	ArgumentSourceDeclared ArgumentSource = iota
	// ArgumentSourcePermissionCode identifies the permission code injected by the runtime.
	ArgumentSourcePermissionCode
)

// Argument describes one service-method or task-trigger argument.
type Argument struct {
	// Source identifies whether the caller or runtime supplies this argument.
	Source ArgumentSource
	// Pos is the argument's source position.
	Pos Position
	// Name is the argument's local name.
	Name string
	// Description is the argument's documentation text.
	Description string
	// Deprecated reports whether the argument should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the argument is deprecated and what to use instead.
	DeprecatedReason string
	// Example is the argument's example value as source text.
	Example string
	// Sensitive reports whether the argument carries redaction semantics.
	Sensitive bool
	// Type is the argument's resolved semantic type.
	Type *Type
}

// HasClientRules reports whether a service declares portal admission rules.
func (s *Service) HasClientRules() bool {
	if len(s.Audiences) > 0 || (s.Auth != "" && s.Auth != AuthModeUnset) || s.Require != nil {
		return true
	}
	for _, method := range s.Methods {
		if (method.Auth != "" && method.Auth != AuthModeUnset) || method.Require != nil {
			return true
		}
	}
	return false
}

// ClientApi includes API services and legacy client rules, excluding extension contracts.
func (s *Service) ClientApi() bool { return !s.Ext && (s.Api || s.HasClientRules()) }

// Public reports whether the service is exported by pub or ext.
func (s *Service) Public() bool {
	return s.Pub || s.Ext
}
