// Package features computes language-server responses from immutable workspace
// snapshots. It does not own mutable server or client state.
package features

import "go.yorun.ai/skel/internal/lsp/workspace"

// Service evaluates language features against one workspace snapshot.
type Service struct {
	Snapshot               workspace.Snapshot
	SnippetSupport         bool
	DocumentChangesSupport bool
}
