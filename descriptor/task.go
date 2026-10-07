package descriptor

// Task describes a background task.
type Task struct {
	Name             string         `json:"name"`
	SkelName         string         `json:"skelName"`
	Description      string         `json:"description,omitempty"`
	Deprecated       bool           `json:"deprecated,omitzero"`
	DeprecatedReason string         `json:"deprecatedReason,omitempty"`
	Hash             string         `json:"hash"`
	Triggers         []*TaskTrigger `json:"triggers"`
}

// TaskTrigger describes one task trigger.
type TaskTrigger struct {
	Name               string    `json:"name"`
	SkelName           string    `json:"skelName"`
	Description        string    `json:"description,omitempty"`
	Deprecated         bool      `json:"deprecated,omitzero"`
	DeprecatedReason   string    `json:"deprecatedReason,omitempty"`
	Hash               string    `json:"hash"`
	InputDescription   string    `json:"inputDescription,omitempty"`
	ArgumentsSensitive bool      `json:"argumentsSensitive,omitzero"`
	Arguments          []*Member `json:"arguments,omitempty"`
}
