package schema

// Enum describes an enum declaration.
type Enum struct {
	// Pos is the enum declaration's source position.
	Pos Position
	// Name is the enum's local name.
	Name string
	// SkelName is the enum's fully qualified Skel name.
	SkelName string
	// Hash is the enum's compatibility hash.
	Hash string
	// Domain is the fully qualified name of the owning domain.
	Domain string
	// Description is the enum's documentation text.
	Description string
	// Deprecated reports whether the enum should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the enum is deprecated and what to use instead.
	DeprecatedReason string
	// Pub reports whether the enum belongs to the public contract.
	Pub bool
	// UnspecifiedItem is the required fallback item.
	UnspecifiedItem *EnumItem
	// Items lists explicitly declared enum items in source order.
	Items []*EnumItem
}

// EnumItem describes one item in an enum declaration.
type EnumItem struct {
	// Pos is the item's source position.
	Pos Position
	// Name is the item's local name.
	Name string
	// Description is the item's documentation text.
	Description string
	// Deprecated reports whether the item should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the item is deprecated and what to use instead.
	DeprecatedReason string
}
