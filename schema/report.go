package schema

// ImpactLevel classifies a schema change's compatibility impact.
type ImpactLevel string

const (
	// ImpactBreaking identifies a structurally incompatible change.
	ImpactBreaking ImpactLevel = "BREAKING"
	// ImpactDangerous identifies a structurally compatible semantic change.
	ImpactDangerous ImpactLevel = "DANGEROUS"
	// ImpactCompatible identifies a compatible change.
	ImpactCompatible ImpactLevel = "COMPATIBLE"
)

// ChangeType identifies whether a schema element was added, removed, or modified.
type ChangeType string

const (
	// ChangeAdded identifies an added schema element.
	ChangeAdded ChangeType = "ADDED"
	// ChangeRemoved identifies a removed schema element.
	ChangeRemoved ChangeType = "REMOVED"
	// ChangeModified identifies a modified schema element.
	ChangeModified ChangeType = "MODIFIED"
)

// Change is one change in a Report.
type Change struct {
	Code      string      `json:"code"`
	Change    ChangeType  `json:"change"`
	Impact    ImpactLevel `json:"impact"`
	Symbol    string      `json:"symbol"`
	Message   string      `json:"message"`
	Baseline  *Position   `json:"baseline,omitempty"`
	Candidate *Position   `json:"candidate,omitempty"`
}

// Summary contains schema diff counts grouped by impact level.
type Summary struct {
	Breaking   int `json:"breaking"`
	Dangerous  int `json:"dangerous"`
	Compatible int `json:"compatible"`
}

// Report is the complete JSON report emitted by schema diff.
type Report struct {
	Compatible      bool      `json:"compatible"`
	BaselineDomain  string    `json:"baselineDomain"`
	CandidateDomain string    `json:"candidateDomain"`
	Summary         Summary   `json:"summary"`
	Changes         []*Change `json:"changes"`
}

// HasImpact reports whether the diff contains a change at the requested impact level.
func (r *Report) HasImpact(impact ImpactLevel) bool {
	for _, change := range r.Changes {
		if change.Impact == impact {
			return true
		}
	}
	return false
}
