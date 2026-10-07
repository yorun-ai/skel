package schema

// Domain is the semantic schema for one Skel domain.
//
// Domain exposes ordered declaration collections through accessor methods so
// generators can preserve deterministic source semantics.
type Domain struct {
	name        string
	description string
	hash        string

	imports   []*Import
	enums     []*Enum
	data      []*Data
	configs   []*Data
	events    []*Data
	actors    []*Actor
	resources []*Resource
	webs      []*Web
	services  []*Service
	tasks     []*Task
}

// DomainSpec contains the values used to construct a [Domain].
// Declaration slices retain their supplied order.
type DomainSpec struct {
	// Name is the fully qualified domain name.
	Name string
	// Description is the domain's documentation text.
	Description string
	// Hash is the domain's compatibility hash.
	Hash string
	// Imports lists imported domains in deterministic order.
	Imports []*Import
	// Enums lists enum declarations in deterministic order.
	Enums []*Enum
	// Data lists data declarations in deterministic order.
	Data []*Data
	// Configs lists config declarations in deterministic order.
	Configs []*Data
	// Events lists event payload declarations in deterministic order.
	Events []*Data
	// Actors lists actor declarations in deterministic order.
	Actors []*Actor
	// Resources lists resource declarations in deterministic order.
	Resources []*Resource
	// Webs lists web entry-point declarations in deterministic order.
	Webs []*Web
	// Services lists service declarations in deterministic order.
	Services []*Service
	// Tasks lists task declarations in deterministic order.
	Tasks []*Task
}

// Import describes one domain imported by a Skel contract.
type Import struct {
	// Pos is the import declaration's source position.
	Pos Position
	// Domain is the parsed semantic schema of the imported domain.
	// It is nil when inspection leaves dependencies unresolved.
	Domain *Domain
	// Name is the imported domain's fully qualified name.
	Name string
	// Alias is the qualifier used to reference the imported domain.
	Alias string
	// ExplicitAlias reports whether Alias was declared explicitly in source.
	ExplicitAlias bool
}

// NewDomainFromSpec constructs a domain from already validated semantic data.
// It does not validate, normalize, copy, or hash the supplied values.
func NewDomainFromSpec(spec DomainSpec) *Domain {
	return new(Domain{
		name:        spec.Name,
		description: spec.Description,
		hash:        spec.Hash,
		imports:     spec.Imports,
		enums:       spec.Enums,
		data:        spec.Data,
		configs:     spec.Configs,
		events:      spec.Events,
		actors:      spec.Actors,
		resources:   spec.Resources,
		webs:        spec.Webs,
		services:    spec.Services,
		tasks:       spec.Tasks,
	})
}

// Name returns the fully qualified domain name.
func (d *Domain) Name() string { return d.name }

// Description returns the domain's documentation text.
func (d *Domain) Description() string { return d.description }

// Hash returns the domain's compatibility hash.
func (d *Domain) Hash() string { return d.hash }

// Imports returns imported domains in deterministic order.
func (d *Domain) Imports() []*Import { return d.imports }

// Enums returns enum declarations in deterministic order.
func (d *Domain) Enums() []*Enum { return d.enums }

// Data returns data declarations in deterministic order.
func (d *Domain) Data() []*Data { return d.data }

// Configs returns config declarations in deterministic order.
func (d *Domain) Configs() []*Data { return d.configs }

// Events returns event payload declarations in deterministic order.
func (d *Domain) Events() []*Data { return d.events }

// Actors returns actor declarations in deterministic order.
func (d *Domain) Actors() []*Actor { return d.actors }

// Resources returns resource declarations in deterministic order.
func (d *Domain) Resources() []*Resource { return d.resources }

// Webs returns web entry-point declarations in deterministic order.
func (d *Domain) Webs() []*Web { return d.webs }

// Services returns service declarations in deterministic order.
func (d *Domain) Services() []*Service { return d.services }

// Tasks returns task declarations in deterministic order.
func (d *Domain) Tasks() []*Task { return d.tasks }
