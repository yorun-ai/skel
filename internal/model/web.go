package model

// Web describes a web entry point and the actors allowed to access it.
type Web struct {
	// Pos is the web declaration's source position.
	Pos Position
	// Name is the web entry point's local name.
	Name string
	// SkelName is the web entry point's fully qualified Skel name.
	SkelName string
	// Hash is the web entry point's compatibility hash.
	Hash string
	// Description is the web entry point's documentation text.
	Description string
	// Deprecated reports whether the web entry point should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the web entry point is deprecated and what to use instead.
	DeprecatedReason string
	// Audiences lists actors allowed to access the entry point.
	Audiences []*ActorAudience
	// Auth is the declared authentication mode; unset retains runtime defaults.
	Auth AuthMode
	// AuthPos is the auth marker source position, or zero when omitted.
	AuthPos Position
	// MountPath is the optional absolute HTTP mount path; empty means unspecified.
	MountPath string
}
