# Internal capabilities

Skel is a language. Its compiler, formatter and language server compose reusable
capabilities; using a capability does not make that capability part of a tool.

| Directory | Responsibility |
| --- | --- |
| `parser`, `symbol`, `analyzer` | Syntax, symbols and semantic analysis |
| `loader`, `source`, `location` | Input providers, immutable source revisions and shared positions |
| `hasher` | Compatibility hashes |
| `formatter` | Pure source formatting |
| `codegen` | Go generator SDK inputs, selection, traversal and execution |
| `codegen/binding` | Shared rendering, import mappings and binding option helpers |
| `codegen/output` | Generated-file markers and managed output transactions |
| `codegen/binding/{golang,typescript,skeleton}` | Target language bindings, templates and module metadata |
| `lsp` | Editor protocol, workspace state, scheduling and language features |
| `sourcediff` | Prepare source baselines and compare their semantic schemas |
| `compiler` | Compose source loading, parsing, analysis and hashing |
| `api` | Programmatic input adaptation and cross-capability orchestration |
| `cmd/skelc` | skelc commands, flags, terminal output and execution |
| `cmd/skelc/output` | Command output structures, error codes and exit codes |
| `optionvalidation` | Shared validation errors for source and generator options |
| `util`, `testutil` | Reusable helpers and test infrastructure |

The public `schema` package directly owns semantic declarations and source positions.
`descriptor` owns runtime metadata with named references and no compiler, source or
runtime dependencies. `types` owns portable Go scalar values. These packages have
no matching internal implementations. `schema/diff` compares semantic domains
without loading sources, using named reference identities to avoid expanding
recursive graphs. CLI declaration serialization belongs to `cmd/skelc/output`.
Go's `descriptor` binding builds public descriptors directly from semantic schemas
and renders descriptor registration calls.
The compiler
core must not depend on API adapters, command implementations or LSP.
Capability packages must not depend on compiler orchestration simply because
skelc uses them.

`symbol` owns declaration identities, scopes, syntax reference indexing and name
resolution shared by semantic analysis and LSP. It accepts recovered syntax and
does not require a valid semantic schema. Semantic validation stays in `analyzer`.

The public `api` facade provides programmatic source and generation APIs
through `internal/api`. Capability implementations must not import either API
package. Target generators own option normalization, import validation and rendering.
The public `codegen` facade exposes validated generation inputs, declaration selection,
type traversal and generic substitution through `internal/codegen`. Its shared runner
validates returned files and publishes them through managed output transactions.
Semantic declarations stay in the public `schema` package; target import paths and aliases stay in bindings.
`codegen/binding` provides shared rendering and import helpers for the language
bindings. The SDK remains independent of binding implementations; `codegen/output`
remains independent of both the SDK and bindings. Semantic failures and cycle
detection belong to `analyzer`; terminal log formatting belongs to `cmd/skelc`.
Target-specific naming conflicts, including Go sensitive marker methods, are
validated by the binding against the declarations it actually emits.
The public `cmd/skelc/output` package exposes command result contracts. `cmd/skelc`
is a thin executable entry point into `internal/cmd/skelc`, which can invoke
programmatic APIs, the formatter or the language server as appropriate.
