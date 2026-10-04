package view

import (
	"fmt"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
)

type Mode string

const (
	ModeApi     Mode = "api"
	ModeFull    Mode = "full"
	ModePub     Mode = "pub"
	ModeRegular Mode = "regular"
)

type Domain struct {
	mode      Mode
	Enums     []*model.Enum
	Data      []*model.Data
	Configs   []*model.Data
	Actors    []*model.Actor
	Resources []*model.Resource
	Webs      []*model.Web
	Events    []*model.Data
	Services  []*model.Service
	Tasks     []*model.Task
}

func Full(domain *model.Domain) *Domain {
	return &Domain{
		mode:      ModeFull,
		Enums:     domain.Enums(),
		Data:      domain.Data(),
		Configs:   domain.Configs(),
		Actors:    domain.Actors(),
		Resources: domain.Resources(),
		Webs:      domain.Webs(),
		Events:    domain.Events(),
		Services:  domain.Services(),
		Tasks:     domain.Tasks(),
	}
}

func Build(mode Mode, domain *model.Domain, selection common.ApiFilter) (*Domain, error) {
	if mode == ModeApi {
		api, err := common.BuildApiView(domain, selection)
		if err != nil {
			return nil, err
		}
		return &Domain{mode: mode, Enums: api.Enums, Data: api.Data, Services: api.Services}, nil
	}
	if mode == ModePub {
		public, err := common.BuildPublicView(domain)
		if err != nil {
			return nil, err
		}
		return &Domain{
			mode:      mode,
			Enums:     public.Enums,
			Data:      public.Data,
			Configs:   public.Configs,
			Actors:    public.Actors,
			Resources: public.Resources,
			Webs:      []*model.Web{},
			Events:    public.Events,
			Services:  public.Services,
			Tasks:     []*model.Task{},
		}, nil
	}
	if mode == ModeFull {
		return Full(domain), nil
	}
	if mode != ModeRegular {
		return nil, fmt.Errorf("invalid Go generation mode %q", mode)
	}
	public, err := common.BuildPublicView(domain)
	if err != nil {
		return nil, err
	}
	return &Domain{
		mode:      mode,
		Enums:     without(domain.Enums(), public.Enums),
		Data:      without(domain.Data(), public.Data),
		Configs:   filterNonPubData(domain.Configs()),
		Actors:    filterNonPubActors(domain.Actors()),
		Resources: filterNonPubResources(domain.Resources()),
		Webs:      domain.Webs(),
		Events:    domain.Events(),
		Services:  domain.Services(),
		Tasks:     domain.Tasks(),
	}, nil
}

// New constructs a generation view and reports invalid modes or public views.
func New(mode Mode, domain *model.Domain) (*Domain, error) {
	return Build(mode, domain, common.ApiFilter{})
}

func without[T any](all, excluded []*T) []*T {
	seen := map[*T]bool{}
	for _, value := range excluded {
		seen[value] = true
	}
	result := make([]*T, 0)
	for _, value := range all {
		if !seen[value] {
			result = append(result, value)
		}
	}
	return result
}
