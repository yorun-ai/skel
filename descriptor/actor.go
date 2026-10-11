package descriptor

// Actor describes an identity and its generated authentication and permission contracts.
type Actor struct {
	Name             string           `json:"name"`
	SkelName         string           `json:"skelName"`
	Description      string           `json:"description,omitempty"`
	Deprecated       bool             `json:"deprecated,omitzero"`
	DeprecatedReason string           `json:"deprecatedReason,omitempty"`
	Hash             string           `json:"hash"`
	Auth             *ActorAuth       `json:"auth,omitzero"`
	Permission       *ActorPermission `json:"permission,omitzero"`
}

// ActorAuth describes an actor's declared authentication and its derived service.
// A nil Actor.Auth means the actor does not declare authentication.
type ActorAuth struct {
	Credential      *Data    `json:"credential"`
	Info            *Data    `json:"info"`
	IdentifierField string   `json:"identifierField,omitempty"`
	Service         *Service `json:"service"`
	// MethodName references the authentication method by its local name in Service.
	MethodName string `json:"methodName"`
}

// Method returns the authentication method from Service, or nil if unavailable.
func (auth *ActorAuth) Method() *Method {
	if auth == nil {
		return nil
	}
	return auth.Service.Method(auth.MethodName)
}

// ActorPermission describes an actor's declared permission support and its derived service.
// A nil Actor.Permission means the actor does not declare permission support.
type ActorPermission struct {
	Service *Service `json:"service"`
	// MethodName references the permission-checking method by its local name in Service.
	MethodName string `json:"methodName"`
}

// Method returns the permission-checking method from Service, or nil if unavailable.
func (permission *ActorPermission) Method() *Method {
	if permission == nil {
		return nil
	}
	return permission.Service.Method(permission.MethodName)
}

// ActorAudience identifies an actor allowed to access an entry point.
type ActorAudience struct {
	Name     string `json:"name"`
	SkelName string `json:"skelName"`
}

func hasAudience(audiences []*ActorAudience, actor string) bool {
	for _, audience := range audiences {
		if audience.SkelName == actor {
			return true
		}
	}
	return false
}
