package codegen

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/optionvalidation"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

// ApiFilter selects API service/type roots and optionally prunes public types.
type ApiFilter struct {
	// Actors contains fully qualified actor names. Empty selects all API services
	// unless Prune is enabled, in which case it selects no services.
	// Multiple names select the union of matching services, regardless of transport.
	Actors []string
	// Prune retains only selected services/types and their local type closure.
	Prune bool
	// Types lists fully qualified local data/enum roots; requires Prune.
	Types []string
}

// BuildApiView selects client services and all locally owned API data dependencies.
// Without pruning, explicitly public types remain available for other domains.
func BuildApiView(domain *schema.Domain, selection ApiFilter) (*PublicView, error) {
	var err error
	selection, err = NormalizeApiFilter(selection)
	if err != nil {
		return nil, err
	}
	types := map[string]bool{}
	for _, name := range selection.Types {
		found := false
		for _, d := range domain.Data() {
			found = found || name == domain.Name()+"."+d.Name
		}
		for _, e := range domain.Enums() {
			found = found || name == domain.Name()+"."+e.Name
		}
		if !found {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldApiType, optionvalidation.RuleInvalid, fmt.Sprintf("unknown local API data or enum %q", name))
		}
		types[name] = true
	}
	selected := map[string]bool{}
	for _, name := range selection.Actors {
		found := false
		for _, actor := range domain.Actors() {
			found = found || name == domain.Name()+"."+actor.Name
		}
		for _, imported := range domain.Imports() {
			for _, actor := range imported.Domain.Actors() {
				found = found || name == imported.Domain.Name()+"."+actor.Name
			}
		}
		if !found {
			return nil, optionvalidation.NewValidationError(optionvalidation.FieldApiActor, optionvalidation.RuleInvalid, fmt.Sprintf("unknown fully qualified API actor %q", name))
		}
		selected[name] = true
	}
	result := &PublicView{
		Data:  filter(domain.Data(), func(d *schema.Data) bool { return (!selection.Prune && d.Pub) || types[domain.Name()+"."+d.Name] }),
		Enums: filter(domain.Enums(), func(e *schema.Enum) bool { return (!selection.Prune && e.Pub) || types[domain.Name()+"."+e.Name] }),
		Services: filter(domain.Services(), func(s *schema.Service) bool {
			return s.ClientApi() && ((!selection.Prune && len(selected) == 0) || matchesApiActors(domain, s, selected))
		}),
	}
	collectViewData(domain, result)
	return result, nil
}

func matchesApiActors(domain *schema.Domain, service *schema.Service, selected map[string]bool) bool {
	for _, audience := range service.Audiences {
		qualifier, name, qualified := nameutil.SplitQualified(audience.Actor)
		if !qualified {
			if selected[domain.Name()+"."+audience.Actor] {
				return true
			}
			continue
		}
		for _, imported := range domain.Imports() {
			if qualifier == imported.Alias || qualifier == imported.Name {
				if selected[imported.Domain.Name()+"."+name] {
					return true
				}
			}
		}
	}
	return false
}

// NormalizeApiFilter validates roots and copies their names without mutating caller options.
func NormalizeApiFilter(selection ApiFilter) (ApiFilter, error) {
	if len(selection.Types) > 0 && !selection.Prune {
		return ApiFilter{}, optionvalidation.NewValidationError(optionvalidation.FieldApiType, optionvalidation.RuleRequiresPrune, "type filter requires prune")
	}
	if selection.Prune && len(selection.Actors) == 0 && len(selection.Types) == 0 {
		return ApiFilter{}, optionvalidation.NewValidationError(optionvalidation.FieldApiPrune, optionvalidation.RuleRequired, "prune requires at least one actor or type")
	}
	result := ApiFilter{Prune: selection.Prune}
	for _, value := range selection.Actors {
		name := strings.TrimSpace(value)
		if qualifier, local, ok := nameutil.SplitQualified(name); !ok || qualifier == "" || local == "" {
			return ApiFilter{}, optionvalidation.NewValidationError(optionvalidation.FieldApiActor, optionvalidation.RuleInvalid, fmt.Sprintf("API actor %q must be a fully qualified name", value))
		}
		result.Actors = append(result.Actors, name)
	}
	for _, value := range selection.Types {
		name := strings.TrimSpace(value)
		if qualifier, local, ok := nameutil.SplitQualified(name); !ok || qualifier == "" || local == "" {
			return ApiFilter{}, optionvalidation.NewValidationError(optionvalidation.FieldApiType, optionvalidation.RuleInvalid, fmt.Sprintf("API type %q must be a fully qualified data or enum name", value))
		}
		result.Types = append(result.Types, name)
	}
	return result, nil
}

// ValidateApiFilterMode rejects API selection flags on non-API targets.
func ValidateApiFilterMode(selection ApiFilter, api bool) error {
	if api {
		return nil
	}
	if len(selection.Actors) > 0 {
		return optionvalidation.NewValidationError(optionvalidation.FieldApiActor, optionvalidation.RuleRequiresApi, "actor filter requires api")
	}
	if selection.Prune {
		return optionvalidation.NewValidationError(optionvalidation.FieldApiPrune, optionvalidation.RuleRequiresApi, "prune requires api")
	}
	if len(selection.Types) > 0 {
		return optionvalidation.NewValidationError(optionvalidation.FieldApiType, optionvalidation.RuleRequiresApi, "type filter requires api")
	}
	return nil
}

func collectViewData(domain *schema.Domain, view *PublicView) {
	data := map[*schema.Data]bool{}
	enums := map[*schema.Enum]bool{}
	var visitType func(*schema.Type)
	visitData := func(d *schema.Data) {
		if d == nil || data[d] {
			return
		}
		data[d] = true
		for _, member := range d.Members {
			visitType(member.Type)
		}
	}
	visitType = func(t *schema.Type) {
		if t == nil {
			return
		}
		// Generic arguments belong to the caller even when the generic definition is imported.
		for _, arg := range t.TypeArguments {
			visitType(arg)
		}
		if t.ExternalDomain != "" {
			return
		}
		switch t.Kind {
		case schema.TypeKindData:
			visitData(t.Data)
		case schema.TypeKindEnum:
			enums[t.Enum] = true
		case schema.TypeKindList:
			if t.List != nil {
				visitType(t.List.Element)
			}
		case schema.TypeKindMap:
			if t.Map != nil {
				visitType(t.Map.Key)
				visitType(t.Map.Value)
			}
		}
	}
	for _, d := range view.Data {
		visitData(d)
	}
	for _, e := range view.Enums {
		enums[e] = true
	}
	for _, d := range view.Configs {
		visitData(d)
	}
	for _, d := range view.Events {
		visitData(d)
	}
	for _, a := range view.Actors {
		if a.Auth != nil {
			visitData(a.Auth.Credential)
			visitData(a.Auth.Info)
		}
	}
	services := append([]*schema.Service{}, view.Services...)
	for _, a := range view.Actors {
		if a.Auth != nil {
			services = append(services, a.Auth.Service)
		}
		if a.Permission != nil {
			services = append(services, a.Permission.Service)
		}
	}
	for _, r := range view.Resources {
		services = append(services, r.CheckService)
	}
	for _, s := range services {
		if s == nil {
			continue
		}
		for _, method := range s.Methods {
			for _, arg := range method.Arguments {
				visitType(arg.Type)
			}
			visitType(method.ResultType)
		}
	}
	// Preserve declaration order and exclude generated argument/actor data.
	view.Data = filter(domain.Data(), func(d *schema.Data) bool { return data[d] })
	view.Enums = filter(domain.Enums(), func(e *schema.Enum) bool { return enums[e] })
}

// ApiTypeRoots returns the types sent by API callers and returned to them.
func ApiTypeRoots(data []*schema.Data, services []*schema.Service) []*schema.Type {
	var roots []*schema.Type
	for _, item := range data {
		for _, member := range item.Members {
			roots = append(roots, member.Type)
		}
	}
	for _, service := range services {
		for _, method := range service.Methods {
			roots = append(roots, method.ResultType)
			for _, arg := range method.Arguments {
				if arg.Source == schema.ArgumentSourceDeclared {
					roots = append(roots, arg.Type)
				}
			}
		}
	}
	return roots
}
