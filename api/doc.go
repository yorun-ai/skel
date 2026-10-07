// Package api provides the supported programmatic API for inspecting Skel
// sources, querying schemas, comparing contracts, and generating Go,
// TypeScript, and public Skel output.
//
// Use [Parse] when several generators or a custom generator need to share one
// validated schema. The Compile functions combine parsing and one generation
// step. [Check] and [ScanImports] inspect sources without loading dependencies.
// [QuerySchema] returns normalized schema documents for inspection or comparison.
//
// Generated files carry an ownership marker. Existing files without that
// marker are preserved.
package api
