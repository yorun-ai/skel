package codegen

import (
	"cmp"
	"slices"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/schema"
)

// ApiTypeDependency identifies a foreign declaration, not a generic instantiation.
// Generic arguments are reported independently in their owning domains.
type ApiTypeDependency = schema.Dependency

// ApiDependencyReport describes exactly the selected API declaration view.
type ApiDependencyReport struct {
	Domain       string              `json:"domain"`
	Services     []string            `json:"services"`
	Data         []string            `json:"data"`
	Enums        []string            `json:"enums"`
	Dependencies []ApiTypeDependency `json:"dependencies"`
}

// ApiDependencies expands local types and reports foreign type boundaries.
func ApiDependencies(domain *model.Domain, selection ApiFilter) (*ApiDependencyReport, error) {
	view, err := BuildApiView(domain, selection)
	if err != nil {
		return nil, err
	}
	result := &ApiDependencyReport{Domain: domain.Name(), Services: []string{}, Data: []string{}, Enums: []string{}, Dependencies: []ApiTypeDependency{}}
	for _, service := range view.Services {
		result.Services = append(result.Services, domain.Name()+"."+service.Name)
	}
	for _, data := range view.Data {
		result.Data = append(result.Data, domain.Name()+"."+data.Name)
	}
	for _, enum := range view.Enums {
		result.Enums = append(result.Enums, domain.Name()+"."+enum.Name)
	}
	seen := map[ApiTypeDependency]bool{}
	VisitTypes(ApiTypeRoots(view.Data, view.Services), func(kind *model.Type) {
		if kind.ExternalDomain == "" {
			return
		}
		dependency := ApiTypeDependency{Domain: kind.ExternalDomain}
		switch kind.Kind {
		case model.TypeKindData:
			dependency.Name = kind.Data.Name
			dependency.Kind = "data"
		case model.TypeKindEnum:
			dependency.Name = kind.Enum.Name
			dependency.Kind = "enum"
		default:
			return
		}
		if !seen[dependency] {
			seen[dependency] = true
			result.Dependencies = append(result.Dependencies, dependency)
		}
	})
	slices.Sort(result.Services)
	slices.Sort(result.Data)
	slices.Sort(result.Enums)
	slices.SortFunc(result.Dependencies, func(a, b ApiTypeDependency) int {
		return cmp.Or(cmp.Compare(a.Domain, b.Domain), cmp.Compare(a.Name, b.Name), cmp.Compare(a.Kind, b.Kind))
	})
	return result, nil
}
