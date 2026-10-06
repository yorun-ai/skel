// Package model exposes the supported parser-independent semantic data model
// for Skel contracts. It is a public facade over the compiler semantic model
// implementation.
//
// Values in this package are produced by go.yorun.ai/skel/api.Parse
// function and are the supported input for custom code generators. Names in
// SkelName fields are fully qualified, while Name fields contain the
// declaration's local name. Hash fields are deterministic compatibility hashes
// calculated by skelc. The implementation remains internal so only names
// deliberately re-exported by this package form part of the public API.
//
// Model values retain the resolved references and compiler state needed by
// custom generators. For a normalized, versioned compatibility representation
// derived from this model, call go.yorun.ai/skel/api.ProjectSchema and use
// go.yorun.ai/skel/schema for validation, encoding and contract comparisons.
package model
