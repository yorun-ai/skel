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

// Dependencies consumes a normalized projection without loading or compiling inputs.
func Dependencies(document *Document) *DependencyReport {
	result := newDependencyReport(document.Domain)
	seen := map[Dependency]bool{}
	add := func(name, kind string) {
		i := strings.LastIndex(name, ".")
		if i < 0 || name[:i] == document.Domain {
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
		case TypeKindData, TypeKindEnum, TypeKindConfig, TypeKindEvent:
			add(value.Name, string(value.Kind))
		}
		for _, arg := range value.Arguments {
			visitType(arg)
		}
		visitType(value.Element)
		visitType(value.Key)
		visitType(value.Value)
	}
	data := func(value *DataSchema) {
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
	var requirement func(*Requirement)
	requirement = func(value *Requirement) {
		if value == nil {
			return
		}
		if value.Check != nil {
			add(value.Check.Resource, "resource")
			for _, arg := range value.Check.Arguments {
				visitType(arg.Type)
			}
		}
		if value.Mode == RequirementModeCode {
			if resource, _, ok := strings.Cut(value.Code, ":"); ok {
				add(resource, "resource")
			}
		}
		for _, child := range value.Children {
			requirement(child)
		}
	}
	audiences := func(values []*Audience) {
		for _, audience := range values {
			add(audience.Actor, "actor")
		}
	}
	checks := func(values []*ResourceCheck) {
		for _, check := range values {
			arguments(check.Arguments)
		}
	}
	lists := map[DeclarationType]*[]string{
		DeclarationTypeService: &result.Services, DeclarationTypeData: &result.Data,
		DeclarationTypeEnum: &result.Enums, DeclarationTypeActor: &result.Actors,
		DeclarationTypeConfig: &result.Configs, DeclarationTypeEvent: &result.Events,
		DeclarationTypeResource: &result.Resources, DeclarationTypeWeb: &result.Webs,
		DeclarationTypeTask: &result.Tasks,
	}
	for _, declaration := range document.Declarations {
		list := lists[declaration.Kind]
		if list != nil {
			*list = append(*list, declaration.SkelName)
		}
		data(declaration.Data)
		if declaration.Actor != nil {
			data(declaration.Actor.AuthCredential)
			data(declaration.Actor.AuthInfo)
		}
		if declaration.Resource != nil {
			checks(declaration.Resource.Checks)
			for _, action := range declaration.Resource.Actions {
				checks(action.Checks)
			}
		}
		if declaration.Service != nil {
			audiences(declaration.Service.Audiences)
			requirement(declaration.Service.Require)
			for _, method := range declaration.Service.Methods {
				arguments(method.Arguments)
				visitType(method.Result)
				requirement(method.Require)
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
