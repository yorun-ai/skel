package schema

import (
	"cmp"
	"slices"
)

// DeclarationType identifies the kind of a top-level Skel declaration.
type DeclarationType string

const (
	// DeclarationTypeActor identifies an actor declaration.
	DeclarationTypeActor DeclarationType = "actor"
	// DeclarationTypeConfig identifies a config declaration.
	DeclarationTypeConfig DeclarationType = "config"
	// DeclarationTypeData identifies a data declaration.
	DeclarationTypeData DeclarationType = "data"
	// DeclarationTypeEnum identifies an enum declaration.
	DeclarationTypeEnum DeclarationType = "enum"
	// DeclarationTypeEvent identifies an event declaration.
	DeclarationTypeEvent DeclarationType = "event"
	// DeclarationTypeResource identifies a resource declaration.
	DeclarationTypeResource DeclarationType = "resource"
	// DeclarationTypeService identifies a service declaration.
	DeclarationTypeService DeclarationType = "service"
	// DeclarationTypeTask identifies a task declaration.
	DeclarationTypeTask DeclarationType = "task"
	// DeclarationTypeWeb identifies a web declaration.
	DeclarationTypeWeb DeclarationType = "web"
)

// Declaration is a lookup view of one top-level semantic declaration.
// Exactly one payload is set; it points to the original declaration in Domain.
// The common fields describe that declaration at the time the view is created.
type Declaration struct {
	Kind     DeclarationType
	Name     string
	SkelName string
	// Pub reports whether the declaration uses the pub modifier.
	Pub              bool
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Pos              Position
	Enum             *Enum
	Data             *Data
	Actor            *Actor
	Resource         *Resource
	Service          *Service
	Web              *Web
	Task             *Task
}

// Declarations returns lookup views sorted by kind and fully qualified name.
// Imported and generated declarations are not included.
func (d *Domain) Declarations() []*Declaration {
	result := make([]*Declaration, 0)
	for _, value := range d.Enums() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeEnum, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Enum: value,
		}))
	}
	for _, value := range d.Data() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeData, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Data: value,
		}))
	}
	for _, value := range d.Configs() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeConfig, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Data: value,
		}))
	}
	for _, value := range d.Events() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeEvent, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Data: value,
		}))
	}
	for _, value := range d.Actors() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeActor, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Actor: value,
		}))
	}
	for _, value := range d.Resources() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeResource, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Resource: value,
		}))
	}
	for _, value := range d.Services() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeService, Name: value.Name, SkelName: value.SkelName, Pub: value.Pub,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Service: value,
		}))
	}
	for _, value := range d.Webs() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeWeb, Name: value.Name, SkelName: value.SkelName, Pub: false,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Web: value,
		}))
	}
	for _, value := range d.Tasks() {
		result = append(result, new(Declaration{
			Kind: DeclarationTypeTask, Name: value.Name, SkelName: value.SkelName, Pub: false,
			Description: value.Description, Deprecated: value.Deprecated, DeprecatedReason: value.DeprecatedReason, Pos: value.Pos,
			Task: value,
		}))
	}

	slices.SortFunc(result, func(a, b *Declaration) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.SkelName, b.SkelName))
	})
	return result
}

// Find returns a lookup view or nil when the kind and name are absent.
func (d *Domain) Find(kind DeclarationType, skelName string) *Declaration {
	for _, declaration := range d.Declarations() {
		if declaration.Kind == kind && declaration.SkelName == skelName {
			return declaration
		}
	}
	return nil
}
