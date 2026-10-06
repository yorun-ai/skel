package descriptor

// Web describes a web entry point and its admission rules.
type Web struct {
	Name             string           `json:"name"`
	SkelName         string           `json:"skelName"`
	Description      string           `json:"description,omitempty"`
	Deprecated       bool             `json:"deprecated,omitzero"`
	DeprecatedReason string           `json:"deprecatedReason,omitempty"`
	Hash             string           `json:"hash"`
	AuthMode         AuthMode         `json:"authMode,omitzero"`
	Audiences        []*ActorAudience `json:"audiences"`
	MountPath        string           `json:"mountPath"`
}
