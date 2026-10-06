package descriptor

// Event describes an event payload and its visibility.
type Event struct {
	Name             string `json:"name"`
	SkelName         string `json:"skelName"`
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitzero"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
	Hash             string `json:"hash"`
	Pub              bool   `json:"pub"`
	// Ext exports the emitter contract for use by other domains.
	Ext       bool      `json:"ext,omitzero"`
	Sensitive bool      `json:"sensitive,omitzero"`
	Members   []*Member `json:"members,omitempty"`
}
