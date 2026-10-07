package descriptor

// Service describes a service contract and its methods.
type Service struct {
	Name             string `json:"name"`
	SkelName         string `json:"skelName"`
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitzero"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
	Hash             string `json:"hash"`
	// Pub reports whether the service uses the pub modifier.
	Pub bool `json:"pub"`
	// Api restricts calls to the portal client path.
	Api bool `json:"api,omitzero"`
	// Ext exports the server contract for implementation by other domains.
	Ext       bool             `json:"ext,omitzero"`
	AuthMode  AuthMode         `json:"authMode"`
	Audiences []*ActorAudience `json:"audiences,omitempty"`

	Require *PermissionRequire `json:"require,omitempty"`
	Methods []*Method          `json:"methods"`
}

// HasAudience reports whether the service accepts the actor's Skel name and transport.
// An audience with no Via restriction accepts any transport for that actor.
func (service *Service) HasAudience(actor string, via ActorViaKind) bool {
	return hasAudience(service.Audiences, actor, via)
}

// Method returns the method with the given local name, or nil if unavailable.
func (service *Service) Method(name string) *Method {
	if service == nil || name == "" {
		return nil
	}
	for _, method := range service.Methods {
		if method != nil && method.Name == name {
			return method
		}
	}
	return nil
}

// MethodBySkelName returns the method with the given wire name, or nil if unavailable.
func (service *Service) MethodBySkelName(name string) *Method {
	if service == nil || name == "" {
		return nil
	}
	for _, method := range service.Methods {
		if method != nil && method.SkelName == name {
			return method
		}
	}
	return nil
}

// Method describes a callable method.
type Method struct {
	Name               string             `json:"name"`
	SkelName           string             `json:"skelName"`
	Description        string             `json:"description,omitempty"`
	Deprecated         bool               `json:"deprecated,omitzero"`
	DeprecatedReason   string             `json:"deprecatedReason,omitempty"`
	Hash               string             `json:"hash"`
	Example            string             `json:"example,omitempty"`
	AuthMode           AuthMode           `json:"authMode"`
	Require            *PermissionRequire `json:"require,omitempty"`
	EffectiveAuthMode  AuthMode           `json:"effectiveAuthMode"`
	EffectiveRequire   *PermissionRequire `json:"effectiveRequire,omitempty"`
	InputDescription   string             `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool               `json:"argumentsSensitive,omitzero"`
	OutputDescription  string             `json:"outputDescription,omitempty"`
	OutputExample      string             `json:"outputExample,omitempty"`
	ResultSensitive    bool               `json:"resultSensitive,omitzero"`
	Arguments          []*Member          `json:"arguments,omitempty"`
	ResultType         *Type              `json:"resultType,omitempty"`
}
