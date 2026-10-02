package model

// DataKind identifies the Skel declaration represented by a [Data] value.
type DataKind string

const (
	// DataKindData identifies a data declaration.
	DataKindData DataKind = "data"
	// DataKindConfig identifies a config declaration.
	DataKindConfig DataKind = "config"
	// DataKindEvent identifies an event payload declaration.
	DataKindEvent DataKind = "event"
)

// ConfigLifecycle controls the lifetime of generated configuration values.
type ConfigLifecycle string

const (
	// ConfigLifecycleEternal identifies the eternal config lifecycle.
	ConfigLifecycleEternal ConfigLifecycle = "eternal"
	// ConfigLifecycleInstant identifies the instant config lifecycle.
	ConfigLifecycleInstant ConfigLifecycle = "instant"
)

// Data describes a data, config, or event payload declaration.
type Data struct {
	// Pos is the declaration's source position.
	Pos Position
	// Name is the declaration's local name.
	Name string
	// SkelName is the declaration's fully qualified Skel name.
	SkelName string
	// Hash is the declaration's compatibility hash.
	Hash string
	// Domain is the fully qualified name of the owning domain.
	Domain string
	// Description is the declaration's documentation text.
	Description string
	// Deprecated reports whether the declaration should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the declaration is deprecated and what to use instead.
	DeprecatedReason string
	// Kind identifies whether this value represents data, config, or an event.
	Kind DataKind
	// Lifecycle is set for config declarations.
	Lifecycle ConfigLifecycle
	// Pub reports whether the declaration belongs to the public contract.
	Pub bool
	// Sensitive reports whether values of this data type are sensitive as a whole.
	Sensitive bool
	// TypeParameters lists generic type parameters in declaration order.
	TypeParameters []*TypeParameter
	// Members lists the declaration's fields in source order.
	Members []*DataMember
}

// IsGeneric reports whether d declares one or more type parameters.
func (d *Data) IsGeneric() bool { return len(d.TypeParameters) > 0 }

// DataMember describes one field of a data, config, event payload, or generated
// argument data model.
type DataMember struct {
	// Pos is the member's source position.
	Pos Position
	// Name is the member's local name.
	Name string
	// Description is the member's documentation text.
	Description string
	// Deprecated reports whether the member should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the member is deprecated and what to use instead.
	DeprecatedReason string
	// Example is the member's example value as source text.
	Example string
	// Sensitive reports whether generated Go values mark this member for redaction.
	Sensitive bool
	// Type is the member's resolved semantic type.
	Type *Type
}
