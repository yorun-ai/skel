package codegen

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

// Surface selects the declarations emitted by a generation task.
type Surface string

const (
	SurfaceFull   Surface = "full"
	SurfacePublic Surface = "public"
	SurfaceAPI    Surface = "api"
)

// Selection describes language-level output selection. The zero value selects
// the whole domain. Target names, import paths and runtime options do not belong here.
type Selection struct {
	Surface Surface
	API     ApiFilter
}

// Declarations is an ordered output set referencing the semantic model. Synthetic
// argument data and actor/resource services remain attached to their owners.
type Declarations struct {
	Enums     []*model.Enum
	Data      []*model.Data
	Configs   []*model.Data
	Events    []*model.Data
	Actors    []*model.Actor
	Resources []*model.Resource
	Webs      []*model.Web
	Services  []*model.Service
	Tasks     []*model.Task
}

// Input is a validated generation view over one semantic graph. All reachable
// model values are borrowed and must remain read-only for its lifetime, including
// between generation calls. Build or edit models before calling Prepare.
// Its zero value is invalid. Select reuses validation and the same graph.
type Input struct {
	domain       *model.Domain
	selection    Selection
	declarations Declarations
}

func Prepare(domain *model.Domain, selection Selection) (Input, error) {
	if err := ValidateDomain(domain); err != nil {
		return Input{}, err
	}
	return (Input{domain: domain}).Select(selection)
}

// Model returns the complete, borrowed semantic domain, including imports.
func (in Input) Model() *model.Domain { return in.domain }
func (in Input) Valid() bool          { return in.domain != nil }
func (in Input) Selection() Selection {
	s := in.selection
	s.API.Actors = append([]string(nil), s.API.Actors...)
	s.API.Types = append([]string(nil), s.API.Types...)
	return s
}

// Select selects another output surface from the complete validated graph.
func (in Input) Select(selection Selection) (Input, error) {
	if !in.Valid() {
		return Input{}, fmt.Errorf("codegen input is uninitialized")
	}
	if selection.Surface == "" {
		selection.Surface = SurfaceFull
	}
	if err := ValidateApiFilterMode(selection.API, selection.Surface == SurfaceAPI); err != nil {
		return Input{}, err
	}
	d := in.domain
	var declarations Declarations
	var v *PublicView
	var err error
	switch selection.Surface {
	case SurfaceFull:
		declarations = Declarations{Enums: d.Enums(), Data: d.Data(), Configs: d.Configs(), Events: d.Events(), Actors: d.Actors(), Resources: d.Resources(), Webs: d.Webs(), Services: d.Services(), Tasks: d.Tasks()}
	case SurfacePublic:
		v, err = BuildPublicView(d)
	case SurfaceAPI:
		selection.API, err = NormalizeApiFilter(selection.API)
		if err == nil {
			v, err = BuildApiView(d, selection.API)
		}
	default:
		return Input{}, fmt.Errorf("unknown codegen surface %q", selection.Surface)
	}
	if err != nil {
		return Input{}, err
	}
	if v != nil {
		declarations = Declarations{Enums: v.Enums, Data: v.Data, Configs: v.Configs, Events: v.Events, Actors: v.Actors, Resources: v.Resources, Services: v.Services}
	}
	return Input{domain: d, selection: selection, declarations: declarations}, nil
}

// Declarations returns a copy of the output slices. Declaration values remain
// shared, read-only semantic objects; source order is preserved.
func (in Input) Declarations() Declarations {
	d := in.declarations
	d.Enums = append([]*model.Enum(nil), d.Enums...)
	d.Data = append([]*model.Data(nil), d.Data...)
	d.Configs = append([]*model.Data(nil), d.Configs...)
	d.Events = append([]*model.Data(nil), d.Events...)
	d.Actors = append([]*model.Actor(nil), d.Actors...)
	d.Resources = append([]*model.Resource(nil), d.Resources...)
	d.Webs = append([]*model.Web(nil), d.Webs...)
	d.Services = append([]*model.Service(nil), d.Services...)
	d.Tasks = append([]*model.Task(nil), d.Tasks...)
	return d
}

// TypeRoots returns member, argument and result types emitted by this selection.
// It excludes imported declarations' members and server-injected API arguments.
func (in Input) TypeRoots() []*model.Type {
	return in.declarations.TypeRoots(in.selection.Surface == SurfaceAPI)
}

// ExternalDomains returns sorted direct type dependencies of the output set.
func (in Input) ExternalDomains() []string { return ExternalDomains(in.TypeRoots()) }

// FindData looks up a declared data by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindData(name string) *model.Data {
	return findDeclaration(in, name, (*model.Domain).Data, func(v *model.Data) string { return v.Name })
}

// FindEnum looks up a declared enum by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindEnum(name string) *model.Enum {
	return findDeclaration(in, name, (*model.Domain).Enums, func(v *model.Enum) string { return v.Name })
}

// FindConfig looks up a declared config by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindConfig(name string) *model.Data {
	return findDeclaration(in, name, (*model.Domain).Configs, func(v *model.Data) string { return v.Name })
}

// FindEvent looks up a declared event by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindEvent(name string) *model.Data {
	return findDeclaration(in, name, (*model.Domain).Events, func(v *model.Data) string { return v.Name })
}

// FindService looks up a declared service by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindService(name string) *model.Service {
	return findDeclaration(in, name, (*model.Domain).Services, func(v *model.Service) string { return v.Name })
}

// FindActor looks up a declared actor by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindActor(name string) *model.Actor {
	return findDeclaration(in, name, (*model.Domain).Actors, func(v *model.Actor) string { return v.Name })
}

// FindResource looks up a declared resource by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindResource(name string) *model.Resource {
	return findDeclaration(in, name, (*model.Domain).Resources, func(v *model.Resource) string { return v.Name })
}

// FindWeb looks up a declared web by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindWeb(name string) *model.Web {
	return findDeclaration(in, name, (*model.Domain).Webs, func(v *model.Web) string { return v.Name })
}

// FindTask looks up a declared task by fully qualified Skel name in the
// complete graph, including imports. It does not imply output selection.
func (in Input) FindTask(name string) *model.Task {
	return findDeclaration(in, name, (*model.Domain).Tasks, func(v *model.Task) string { return v.Name })
}
func findDeclaration[T any](in Input, name string, items func(*model.Domain) []*T, localName func(*T) string) *T {
	for _, d := range in.domains() {
		for _, item := range items(d) {
			if d.Name()+"."+localName(item) == name {
				return item
			}
		}
	}
	return nil
}
func (in Input) domains() []*model.Domain {
	var result []*model.Domain
	seen := map[*model.Domain]bool{}
	var visit func(*model.Domain)
	visit = func(d *model.Domain) {
		if d == nil || seen[d] {
			return
		}
		seen[d] = true
		result = append(result, d)
		for _, im := range d.Imports() {
			visit(im.Domain)
		}
	}
	visit(in.domain)
	return result
}
