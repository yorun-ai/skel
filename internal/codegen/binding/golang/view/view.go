package view

import (
	"fmt"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/schema"
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
	Enums     []*schema.Enum
	Data      []*schema.Data
	Configs   []*schema.Data
	Actors    []*schema.Actor
	Resources []*schema.Resource
	Webs      []*schema.Web
	Events    []*schema.Data
	Services  []*schema.Service
	Tasks     []*schema.Task

	// Reexports retains the selected public contract for regular-package facades.
	// It is populated only in ModeRegular.
	Reexports codegen.Declarations
}

func Full(domain *schema.Domain) *Domain {
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

func Build(mode Mode, domain *schema.Domain, selection codegen.ApiFilter) (*Domain, error) {
	input, err := codegen.Prepare(domain, codegen.Selection{})
	if err != nil {
		return nil, err
	}
	return FromInput(mode, input, selection)
}

func FromInput(mode Mode, input codegen.Input, selection codegen.ApiFilter) (*Domain, error) {
	domain := input.Schema()
	if mode == ModeApi {
		api, err := selectedView(input, codegen.Selection{Surface: codegen.SurfaceAPI, API: selection})
		if err != nil {
			return nil, err
		}
		return &Domain{mode: mode, Enums: api.Enums, Data: api.Data, Services: api.Services}, nil
	}
	if mode == ModePub {
		public, err := selectedView(input, codegen.Selection{Surface: codegen.SurfacePublic})
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
			Webs:      []*schema.Web{},
			Events:    public.Events,
			Services:  public.Services,
			Tasks:     []*schema.Task{},
		}, nil
	}
	if mode == ModeFull {
		return Full(domain), nil
	}
	if mode != ModeRegular {
		return nil, fmt.Errorf("invalid Go generation mode %q", mode)
	}
	public, err := selectedView(input, codegen.Selection{Surface: codegen.SurfacePublic})
	if err != nil {
		return nil, err
	}
	return &Domain{
		mode:      mode,
		Enums:     without(domain.Enums(), public.Enums),
		Data:      without(domain.Data(), public.Data),
		Configs:   without(domain.Configs(), public.Configs),
		Actors:    without(domain.Actors(), public.Actors),
		Resources: without(domain.Resources(), public.Resources),
		Webs:      domain.Webs(),
		Events:    domain.Events(),
		Services:  domain.Services(),
		Tasks:     domain.Tasks(),
		Reexports: public,
	}, nil
}

// New constructs a generation view and reports invalid modes or public views.
func New(mode Mode, domain *schema.Domain) (*Domain, error) {
	return Build(mode, domain, codegen.ApiFilter{})
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

func selectedView(input codegen.Input, selection codegen.Selection) (codegen.Declarations, error) {
	selected, err := input.Select(selection)
	if err != nil {
		return codegen.Declarations{}, err
	}
	return selected.Declarations(), nil
}
