package descriptor

// Enum describes a named enumeration.
type Enum struct {
	Name             string      `json:"name"`
	SkelName         string      `json:"skelName"`
	Description      string      `json:"description,omitempty"`
	Deprecated       bool        `json:"deprecated,omitzero"`
	DeprecatedReason string      `json:"deprecatedReason,omitempty"`
	Hash             string      `json:"hash"`
	Items            []*EnumItem `json:"items"`
}

// EnumItem describes one enumeration value.
type EnumItem struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitzero"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
}
