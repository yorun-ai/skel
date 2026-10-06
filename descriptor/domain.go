package descriptor

// Domain describes the runtime contract of a Skel domain.
type Domain struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Hash        string         `json:"hash"`
	Full        bool           `json:"full"`
	Generated   *GeneratedInfo `json:"generated"`
	Enums       []*Enum        `json:"enums,omitempty"`
	Data        []*Data        `json:"data,omitempty"`
	Configs     []*Config      `json:"configs,omitempty"`
	Webs        []*Web         `json:"webs,omitempty"`
	Events      []*Event       `json:"events,omitempty"`
	Actors      []*Actor       `json:"actors,omitempty"`
	Resources   []*Resource    `json:"resources,omitempty"`
	Services    []*Service     `json:"services,omitempty"`
	Tasks       []*Task        `json:"tasks,omitempty"`
}

// GeneratedInfo identifies the compiler that produced a domain descriptor.
// Consumers own compiler-version acceptance and registration policies.
type GeneratedInfo struct {
	CompilerVersion string `json:"compilerVersion"`
}
