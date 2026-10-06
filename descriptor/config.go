package descriptor

// Config describes configuration data and its lifecycle.
type Config struct {
	Name             string    `json:"name"`
	SkelName         string    `json:"skelName"`
	Description      string    `json:"description,omitempty"`
	Deprecated       bool      `json:"deprecated,omitzero"`
	DeprecatedReason string    `json:"deprecatedReason,omitempty"`
	Hash             string    `json:"hash"`
	Pub              bool      `json:"pub"`
	Sensitive        bool      `json:"sensitive,omitzero"`
	Lifecycle        string    `json:"lifecycle"`
	Members          []*Member `json:"members,omitempty"`
}
