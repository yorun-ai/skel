package common

import "go.yorun.ai/skelc/internal/model"

// BuildApiView selects client services and all locally owned API data dependencies.
// Explicitly public types remain available for clients of other domains.
func BuildApiView(domain *model.Domain) *PublicView {
	result := &PublicView{
		Data:     filter(domain.Data(), func(d *model.Data) bool { return d.Pub }),
		Enums:    filter(domain.Enums(), func(e *model.Enum) bool { return e.Pub }),
		Services: filter(domain.Services(), func(s *model.Service) bool { return s.ClientApi() }),
	}
	collectViewData(domain, result)
	return result
}

func collectViewData(domain *model.Domain, view *PublicView) {
	data := map[*model.Data]bool{}
	enums := map[*model.Enum]bool{}
	var visitType func(*model.Type)
	visitData := func(d *model.Data) {
		if d == nil || data[d] {
			return
		}
		data[d] = true
		for _, member := range d.Members {
			visitType(member.Type)
		}
	}
	visitType = func(t *model.Type) {
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
		case model.TypeKindData:
			visitData(t.Data)
		case model.TypeKindEnum:
			enums[t.Enum] = true
		case model.TypeKindList:
			if t.List != nil {
				visitType(t.List.Value)
			}
		case model.TypeKindMap:
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
		visitData(a.AuthCredential)
		visitData(a.AuthInfo)
	}
	services := append([]*model.Service{}, view.Services...)
	for _, a := range view.Actors {
		services = append(services, a.AuthService, a.PermService)
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
	view.Data = filter(domain.Data(), func(d *model.Data) bool { return data[d] })
	view.Enums = filter(domain.Enums(), func(e *model.Enum) bool { return enums[e] })
}

// ApiTypeRoots returns the types sent by API callers and returned to them.
func ApiTypeRoots(data []*model.Data, services []*model.Service) []*model.Type {
	var roots []*model.Type
	for _, item := range data {
		for _, member := range item.Members {
			roots = append(roots, member.Type)
		}
	}
	for _, service := range services {
		for _, method := range service.Methods {
			roots = append(roots, method.ResultType)
			for _, arg := range method.Arguments {
				if arg.Source == model.ArgumentSourceDeclared {
					roots = append(roots, arg.Type)
				}
			}
		}
	}
	return roots
}
