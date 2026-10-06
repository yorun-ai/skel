package codegen

import (
	"go.yorun.ai/skel/internal/model"
)

// ApiImportDomains identifies foreign domains required by the selected API view.
func ApiImportDomains(domain *model.Domain, selection ApiFilter) (map[string]bool, error) {
	if err := ValidateDomain(domain); err != nil {
		return nil, err
	}
	view, err := BuildApiView(domain, selection)
	if err != nil {
		return nil, err
	}
	domains := map[string]bool{}
	VisitTypes(ApiTypeRoots(view.Data, view.Services), func(kind *model.Type) {
		if kind.ExternalDomain != "" {
			domains[kind.ExternalDomain] = true
		}
	})
	return domains, nil
}
