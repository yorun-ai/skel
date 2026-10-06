package schema

import (
	"cmp"
	"slices"
	"strings"
)

// Dependency identifies an external declaration referenced by the selected schema.
type Dependency struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
}

// DependencyReport lists local declarations and their direct external references.
// Foreign declaration members belong to a query in their owning domain.
type DependencyReport struct {
	Domain       string       `json:"domain"`
	Services     []string     `json:"services"`
	Data         []string     `json:"data"`
	Enums        []string     `json:"enums"`
	Actors       []string     `json:"actors,omitzero"`
	Configs      []string     `json:"configs,omitzero"`
	Events       []string     `json:"events,omitzero"`
	Resources    []string     `json:"resources,omitzero"`
	Webs         []string     `json:"webs,omitzero"`
	Tasks        []string     `json:"tasks,omitzero"`
	Dependencies []Dependency `json:"dependencies"`
}

func newDependencyReport(domain string) *DependencyReport {
	return new(DependencyReport{
		Domain: domain, Services: []string{}, Data: []string{}, Enums: []string{},
		Actors: []string{}, Configs: []string{}, Events: []string{},
		Resources: []string{}, Webs: []string{}, Tasks: []string{},
		Dependencies: []Dependency{},
	})
}

// Dependencies inspects semantic declarations without loading or compiling inputs.
func Dependencies(domain *Domain) *DependencyReport {
	result := newDependencyReport(domain.Name())
	seen := map[Dependency]bool{}
	add := func(name, kind string) {
		i := strings.LastIndex(name, ".")
		if i < 0 || name[:i] == domain.Name() {
			return
		}
		dependency := Dependency{Domain: name[:i], Name: name[i+1:], Kind: kind}
		seen[dependency] = true
	}
	var visitType func(*Type)
	visitType = func(value *Type) {
		if value == nil {
			return
		}
		switch value.Kind {
		case TypeKindData:
			if value.Data != nil {
				add(domain.TypeReferenceName(value), string(value.Data.Kind))
			}
		case TypeKindEnum:
			add(domain.TypeReferenceName(value), "enum")
		case TypeKindUnresolvedReference:
			add(domain.TypeReferenceName(value), "importedReference")
		}
		for _, arg := range value.TypeArguments {
			visitType(arg)
		}
		if value.List != nil {
			visitType(value.List.Value)
		}
		if value.Map != nil {
			visitType(value.Map.Key)
			visitType(value.Map.Value)
		}
	}
	data := func(value *Data) {
		if value != nil {
			for _, member := range value.Members {
				visitType(member.Type)
			}
		}
	}
	arguments := func(values []*Argument) {
		for _, arg := range values {
			visitType(arg.Type)
		}
	}
	var requirement func(*PermissionExpression)
	requirement = func(value *PermissionExpression) {
		if value == nil {
			return
		}
		if value.Check != nil {
			add(domain.ReferenceName(value.Check.ResourceSkelName), "resource")
			for _, arg := range value.Check.Arguments {
				visitType(arg.Type)
			}
		}
		if value.Mode == PermissionRequireModeCode {
			if resource, _, ok := strings.Cut(value.Code, ":"); ok {
				add(domain.ReferenceName(resource), "resource")
			}
		}
		for _, child := range value.Children {
			requirement(child)
		}
	}
	audiences := func(values []*ActorAudience) {
		for _, audience := range values {
			add(domain.ReferenceName(audience.Actor), "actor")
		}
	}
	checks := func(values []*ResourceCheck) {
		for _, check := range values {
			arguments(check.Method.Arguments)
		}
	}
	lists := map[DeclarationType]*[]string{
		DeclarationTypeService: &result.Services, DeclarationTypeData: &result.Data,
		DeclarationTypeEnum: &result.Enums, DeclarationTypeActor: &result.Actors,
		DeclarationTypeConfig: &result.Configs, DeclarationTypeEvent: &result.Events,
		DeclarationTypeResource: &result.Resources, DeclarationTypeWeb: &result.Webs,
		DeclarationTypeTask: &result.Tasks,
	}
	for _, declaration := range domain.Declarations() {
		list := lists[declaration.Kind]
		if list != nil {
			*list = append(*list, declaration.SkelName)
		}
		data(declaration.Data)
		if declaration.Actor != nil && declaration.Actor.Auth != nil {
			data(declaration.Actor.Auth.Credential)
			data(declaration.Actor.Auth.Info)
		}
		if declaration.Resource != nil {
			checks(declaration.Resource.Checks)
			for _, action := range declaration.Resource.Actions {
				checks(action.Checks)
			}
		}
		if declaration.Service != nil {
			audiences(declaration.Service.Audiences)
			if declaration.Service.Require != nil {
				requirement(declaration.Service.Require.Expression)
			}
			for _, method := range declaration.Service.Methods {
				arguments(method.Arguments)
				visitType(method.ResultType)
				if method.Require != nil {
					requirement(method.Require.Expression)
				}
			}
		}
		if declaration.Web != nil {
			audiences(declaration.Web.Audiences)
		}
		if declaration.Task != nil {
			for _, trigger := range declaration.Task.Triggers {
				arguments(trigger.Arguments)
			}
		}
	}
	for _, list := range lists {
		slices.Sort(*list)
	}
	for dependency := range seen {
		result.Dependencies = append(result.Dependencies, dependency)
	}
	slices.SortFunc(result.Dependencies, func(a, b Dependency) int {
		return cmp.Or(cmp.Compare(a.Domain, b.Domain), cmp.Compare(a.Name, b.Name), cmp.Compare(a.Kind, b.Kind))
	})
	return result
}
