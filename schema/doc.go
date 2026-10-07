// Package schema defines Skel's parser-independent semantic graph, shared by
// compilation, language tooling, queries and code generation.
//
// A schema retains source positions and links between declarations. Imported
// types may remain unresolved when callers inspect source without dependencies;
// api.Parse resolves dependencies, and codegen.Prepare requires a complete graph.
// Build custom schemas with DomainSpec before preparation; constructors do not
// validate. After preparation, treat the graph and all reachable values as read-only.
// Analysis and codegen.Prepare populate method EffectiveAuthMode/EffectiveRequire.
// ComputeEffectivePolicy computes a policy without mutation; PopulateEffectivePolicies
// refreshes a domain's derived values, and ValidateEffectivePolicy detects stale values.
// Effective policies include inheritance and composition, independently of whether
// imported check bindings have been resolved. Type links in policies remain borrowed.
//
// Names and source qualifiers describe Skel, never target-language packages.
// Actor authentication/permission services, resource check services and argument
// data are language-defined expansions attached to their owning declarations;
// a binding chooses how to represent them in its target language.
//
// Package descriptor defines runtime metadata without source locations or graph
// links. Package schema/diff compares semantic domains directly. Domain.Declarations
// and Domain.Find return views that borrow the original declaration nodes.
package schema
