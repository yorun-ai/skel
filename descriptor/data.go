package descriptor

// Data describes a structured data declaration.
type Data struct {
	Name             string    `json:"name"`
	SkelName         string    `json:"skelName"`
	Description      string    `json:"description,omitempty"`
	Deprecated       bool      `json:"deprecated,omitzero"`
	DeprecatedReason string    `json:"deprecatedReason,omitempty"`
	Hash             string    `json:"hash"`
	Sensitive        bool      `json:"sensitive,omitzero"`
	TypeParameters   []string  `json:"typeParameters,omitempty"`
	Members          []*Member `json:"members,omitempty"`
}

// Member describes a field or callable argument.
type Member struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitzero"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
	Example          string `json:"example,omitempty"`
	Sensitive        bool   `json:"sensitive,omitzero"`
	Type             *Type  `json:"type"`
}
