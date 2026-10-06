// Package model exposes Skel's parser-independent semantic graph. It is shared
// by compilation, schema projection, language tooling and generation.
//
// api.Parse produces resolved semantic values. codegen.Prepare validates a model
// and selects a generation view without copying declaration types. Once prepared,
// the model and all reachable values must be treated as read-only. Build custom
// semantic models with DomainSpec before preparation; constructors do not validate.
//
// Names and source qualifiers describe Skel, never target-language packages.
// Actor authentication/permission services, resource check services and argument
// data are language-defined expansions attached to their owning declarations;
// a binding chooses how to represent them in its target language.
//
// For a versioned interchange and compatibility representation, use
// api.ProjectSchema and the schema package. schema is separate from the semantic
// graph and is not a second code-generation model.
package model
