# Internal capabilities

Skel is a language. Its compiler, formatter and language server compose reusable
capabilities; using a capability does not make that capability part of a tool.

| Directory | Responsibility |
| --- | --- |
| `parser`, `binding`, `analyzer`, `model` | Syntax, symbols, semantic analysis and semantic data |
| `schema` | Shared contract types, validation, codecs, queries and contract comparisons |
| `projection` | Convert semantic models into shared contracts without loading sources |
| `loader`, `source`, `location` | Input providers, immutable source revisions and shared positions |
| `hasher`, `skelmeta` | Compatibility hashes and generated-language conventions |
| `formatter` | Pure source formatting |
| `codegen` | Target generators and output transactions |
| `lsp` | Editor protocol, workspace state, scheduling and language features |
| `sourcediff` | Prepare source baselines and compare their projected contracts |
| `compiler` | Compose source loading, parsing, analysis and hashing |
| `api` | Programmatic input adaptation and cross-capability orchestration |
| `cmd/skelc` | skelc commands, flags, terminal output and execution |
| `cmd/skelc/output` | Command output structures, error codes and exit codes |
| `optionvalidation` | Shared validation errors for source and generator options |
| `util`, `testutil` | Reusable helpers and test infrastructure |

`schema` must not import semantic models or tool implementations. `projection`
may use models and schema, but must not load or compile sources. The compiler
core must not depend on API adapters, command implementations or LSP.
Capability packages must not depend on compiler orchestration simply because
skelc uses them.

The public `api` facade provides programmatic source and generation APIs
through `internal/api`. Capability implementations must not import either API
package. Target generators own option normalization, import validation and managed output transactions.
The public `cmd/skelc/output` package exposes command result contracts. `cmd/skelc`
is a thin executable entry point into `internal/cmd/skelc`, which can invoke
programmatic APIs, the formatter or the language server as appropriate.
