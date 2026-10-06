package codegen

import (
	"slices"

	"go.yorun.ai/skel/schema"
)

// ExternalDomains returns the sorted, unique domains directly referenced by
// the output's type roots. It includes generic arguments but does not follow
// members of imported declarations: those belong to the imported package.
func ExternalDomains(roots []*schema.Type) []string {
	seen := map[string]bool{}
	VisitTypes(roots, func(kind *schema.Type) {
		if kind.ExternalDomain != "" {
			seen[kind.ExternalDomain] = true
		}
	})
	domains := make([]string, 0, len(seen))
	for domain := range seen {
		domains = append(domains, domain)
	}
	slices.Sort(domains)
	return domains
}
