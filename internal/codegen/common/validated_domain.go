package common

import "go.yorun.ai/skel/internal/model"

// ValidatedDomain marks the boundary between the mutable public model builder
// and internal renderers. The model and its reachable declarations are read-only
// for the lifetime of a generation operation; target metadata lives in bindings.
type ValidatedDomain struct{ domain *model.Domain }

func PrepareDomain(domain *model.Domain) (ValidatedDomain, error) {
	if err := ValidateDomain(domain); err != nil {
		return ValidatedDomain{}, err
	}
	return ValidatedDomain{domain: domain}, nil
}
func (d ValidatedDomain) Model() *model.Domain { return d.domain }
