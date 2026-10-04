# Changelog

All notable changes to skelc are documented in this file.

The project follows [Semantic Versioning](https://semver.org/). The public version history starts at `v0.9.0`; versions from the former private repository are not part of the public release history.

## [Unreleased]

## [0.24.2] - 2026-10-04

### Fixed

- Format each generated Go API `MethodSpec` on one line.
- Separate service registration and method initialization groups with one
  blank line, and remove the trailing blank line inside `init()`.

### Upgrade Notes

- Regenerate Go API clients to obtain the formatting changes. Contract
  semantics, initialization order, generated client APIs and runtime dependency
  requirements are unchanged; existing generated clients remain usable.

## [0.24.1] - 2026-10-04

### Changed

- Simplify generated Go API client initialization with
  `vrpc.MustGetMethodInfo`, replacing repeated lookup and panic blocks with
  one assignment per method.
- Raise the default and minimum Go API client runtime to
  `go.yorun.ai/vrpc` v0.13.0. Reject older explicitly selected versions before
  modifying generated output.

### Upgrade Notes

- Regenerate Go API clients to obtain the simplified initialization. Update
  their vRPC dependency to v0.13.0 or later; generated standalone modules
  select it automatically. Existing generated clients can continue to be used.
- Skel contracts, generated client method signatures and wire behavior are
  unchanged. Backend Vine requirements are unchanged.

## [0.24.0] - 2026-10-04

### Added

- Add repeatable `--actor <domain.Actor>` filters to Go and TypeScript API
  client generation. Multiple actors select the union of their API services;
  omitting the flag continues to generate all API services.
- Add `ext service` for contracts implemented by other domains: public Go
  output exports servers, while the defining domain's regular output provides
  clients.
- Add `ext event` for events emitted by other domains: public Go output exports
  emitters, while the defining domain's regular output provides listeners.
  Both outputs share the public payload type, including imported generic data.
- Preserve extension markers in public Skel contracts and normalized and Vine
  runtime schemas. Formatting and language-server features recognize `ext`;
  compatibility hashes and schema diffs account for extension direction changes.

### Changed

- Replace `open service` with `ext service` and remove the `open` modifier.
- Require every `api service` to declare at least one actor audience with `for`.
- Raise both the default and minimum supported Vine version for generated Go
  modules to v0.25.1.

### Upgrade Notes

- Replace `open service` with `ext service` and regenerate public and regular
  packages together. Extension clients now belong to the defining domain's
  regular package; external implementations use the public server contract.
- Add `for Actor` to API services that lack an audience. Anonymous API services
  still declare an actor and use `noauth` to allow anonymous calls.
- Use fully qualified actor names with `--actor`; the flag requires `--api`.
- Upgrade Vine to v0.25.1 or later before using regenerated Go packages.
  Normalized service schemas now use `ext` in place of `open`.

## [0.23.3] - 2026-10-04

### Added

- Add `skelc schema import --skel-in <file-or-directory>` to query direct
  domain imports without loading dependency contracts or supplying import
  mappings. JSON results include domain names, explicit aliases and source
  locations, preserve declarations across files, and use stable ordering.
- Expose the import-query response type through the public `command` package.

### Upgrade Notes

- Existing commands and generated output remain unchanged. No migration or
  regeneration is required.

## [0.23.2] - 2026-10-04

### Fixed

- Derive Go module domain dependencies from each generation view instead of
  including every supplied import mapping. Public, regular, full, and API
  outputs retain the packages they reference and the specified versions,
  including dependencies inferred from a module prefix.
- Keep transitive contract-resolution mappings separate from direct generated
  package dependencies, using shared external-domain analysis for Go and
  TypeScript. Preserve backend and API dependency-conflict validation behavior.

### Upgrade Notes

- Regenerate Go modules to obtain the filtered dependency metadata. No Skel,
  CLI, or generated API migration is required. Continue using `go mod tidy`
  to maintain indirect dependencies and checksums.

## [0.23.1] - 2026-10-02

### Fixed

- Prevent language-server crashes while editing incomplete declarations and
  permission expressions. Syntax recovery keeps inline decorators attached to
  their declarations and continues analyzing later valid declarations.
- Reject malformed programmatic models and schema permission expressions before
  generation or encoding, including missing generic arguments, invalid generic
  arity, and cyclic structural types. Valid recursive data remains supported.
- Detect previously missed hard-reference cycles and apply directory domain
  validation consistently in editor analysis.
- Keep navigation, references, and rename within the correct source input and
  resolved symbol scope. Reject invalid names and generic-parameter conflicts,
  and use document versions when the editor supports them.
- Prevent stale diagnostics from replacing newer results, preserve remote
  document URI identities, and keep document updates responsive with slow
  clients or canceled analysis.
- Preserve path token boundaries during formatting so repeated formatting
  produces stable output.
- Keep target-specific generation state out of shared semantic models so one
  parsed domain can safely be reused across generators.
- Read Git schema baselines from one pinned commit and handle filenames
  containing whitespace correctly.

### Changed

- Reuse incremental analysis results, workspace snapshots, and source indexes
  to reduce editor latency and allocation overhead.
- Preserve generated files and their timestamps when their contents are
  unchanged.
- Strengthen static analysis and regression coverage for malformed input,
  compatibility behavior, and generated Go and TypeScript code.

### Upgrade Notes

- Valid contracts, public Go APIs, CLI options, and generated APIs remain
  compatible with v0.23.0. No migration or mandatory regeneration is required;
  Go and runtime dependency requirements are unchanged.
- Previously missed invalid hard-reference cycles and directory domain
  violations now produce diagnostics. Malformed programmatic models return
  validation errors instead of crashing or recursing indefinitely.
- Upgrade the skelc executable used by the editor as well as the CLI to receive
  the language-server fixes.

## [0.23.0] - 2026-10-02

### Added

- Config fields support nested ordinary data, generic data instantiations,
  binary values, nullable values, and list/map combinations. Public config
  generation includes its reachable local data declarations.

### Changed

- Config and event declarations are final and cannot be referenced as value
  types, including through containers or generic arguments.
- Removed the `@noTrim` config decorator and its generated tag, canonical schema,
  compatibility diff and LSP metadata. Vine v0.25.0 preserves config whitespace
  without an opt-in decorator.
- Updated CLI and Go tooling dependencies: urfave/cli v3.13.0, x/mod v0.41.0
  and x/sys v0.48.0.

### Fixed

- Nullable generic parameter references such as `TValue?` preserve nullability
  in generated Go pointer fields, TypeScript unions and vRPC wire schemas.

### Upgrade Notes

- Remove `@noTrim` annotations before compiling existing contracts.
- Replace config/event value references with ordinary data declarations.
- Regenerate affected Go/TypeScript contracts. Go fields previously emitted as
  `TValue` for `TValue?` now use `*TValue`; update assignments and nil handling.
  Plain generic parameters and nullable collection wrappers retain their types.

## [0.22.1] - 2026-09-23

### Fixed

- TypeScript API generation emits a minimal usable package for contracts with
  no API declarations, including an empty module entry point instead of exports
  referencing files that were not generated.
- Generated Go enum JSON marshalers append formatted bytes directly with
  `fmt.Appendf`, eliminating the intermediate string conversion and its editor
  diagnostic without changing serialized output.

## [0.22.0] - 2026-09-21

### Changed

- Unaliased imports now require the full domain name in references. For example,
  `import ws.sandbox` uses `ws.sandbox.SandboxActor`; add `as sandbox` to keep
  using `sandbox.SandboxActor`. This applies to data and enum types, Actor
  audiences and permission resources. LSP completion, navigation and rename
  support full-domain references. Canonical schema identities and compatibility
  hashes remain independent of the chosen import spelling.
- Go and TypeScript generators allocate deterministic aliases for colliding
  import names, using lowercase Go package aliases and camelCase TypeScript
  namespace aliases. Available explicit aliases retain priority, and generated
  aliases avoid reserved names. Update implicit short references or add explicit
  aliases, then regenerate affected Go/TypeScript packages and public Skel
  contracts. Upgrade the compiler used by editors as well as the CLI.
- Release preparation requires coordinated documentation edits and validation in
  the owning workspace, without requiring documentation-site commits or merges;
  those Git operations still require separate explicit authorization.

### Fixed

- Public Skel exports no longer retain unrelated dependencies when an import
  alias matches another domain name. Actor and type dependencies are selected
  using their resolved domain identity.
- Update the commerce example's imported type and permission references to use
  the full domain name.

## [0.21.0] - 2026-09-17

### Removed

- Generated Go code no longer declares clone methods. Data types drop `Clone()`
  and `CloneBy(...)`, and Rpc method specs drop `CloneArguments` and
  `CloneResult`; the Vine runtime clones in-process arguments and results from
  their declared Go types since Vine v0.20.2. Regenerate every generated Go
  package with this compiler and require Vine v0.20.2 or later, and drop calls
  to `Clone()` or `CloneBy(...)` in application code or clone with
  `vine/util/vbean.DeepClone`. A data member may now be named `clone` or
  `cloneBy`, because no generated method collides with it.

### Changed

- Raise the default and minimum Vine dependency for generated Go code from
  v0.19.0 to v0.20.2.
- Render generated Vine schemas one declaration per line: an enum, data, config,
  web, event, actor, resource, service, task or method keeps its own fields on one
  line, and members, arguments and nested declarations stay one per line. The
  runtime values are unchanged, but generated schema files no longer re-align
  when a single field changes, and hub's admin schema drops from 6170 to 825 lines
  and from 160 KB to 118 KB. Keyed fields inside a generated schema literal may
  appear in a different order than in earlier releases.
- Inline single-line descriptions in generated TypeScript data and service files
  as `/** ... */` instead of a three-line comment block.

## [0.20.0] - 2026-09-16

### Changed

- Raise the default and minimum Vine dependency for generated Go code to v0.19.0.

### Added

- Support optional `mount /path` declarations inside Web objects, including
  path validation, formatting, LSP completion, schema snapshots and compatibility
  diffs. Generated Web specs and runtime schemas both carry `MountPath`.
  Use it when a frontend build needs a stable, immutable public URL prefix for
  its entry point and static assets. Omitting `mount` leaves the Web unrestricted;
  `mount /` explicitly limits it to the root path.
  Mounted Go output requires Vine source with these fields; the generated Go
  dependency now targets Vine v0.19.0.

## [0.19.3] - 2026-09-14

### Removed

- Remove `MinimumGolangApiServiceVineVersion` and the
  `minimumApiServiceVineVersion` field from `skelc version`. All backend Go
  output now uses the same minimum and default Vine versions; integrations
  should use `MinimumGolangVineVersion` and `minimumVineVersion` instead.

### Changed

- Raise the default and minimum Vine dependency for generated Go code to v0.15.7.
- Backend Go generation through the programmatic API now requires an explicit,
  valid `CompilerVersion` of v0.17.1 or later (or the explicit development
  version `v0.0.0-dev`); the CLI continues to populate it automatically.

### Fixed

- Validate generated Go module identities, including public modules and paths
  derived from module prefixes, before committing output.
- Encode TypeScript package metadata with a JSON encoder and reject malformed
  npm package names instead of writing invalid package.json files.
- Report invalid Go and TypeScript import mappings in deterministic domain order.
- Reject non-standard runtime and imported Go dependency versions, including
  versions incompatible with their Go module paths, before committing output.
- Resolve only the selected Go runtime version once per generation request
  and pass validated options through a single typed generation entry point.
- Classify invalid runtime versions, compiler versions, and dependency mappings
  consistently as invalid arguments.
- Reject conflicting Go dependency versions deterministically, including imports
  that would override the selected runtime dependency.
- Reject conflicting explicit TypeScript dependency constraints deterministically
  and preserve explicit constraints when other imports infer wildcard versions.
- Verify imported `Clone` and `CloneBy` compatibility against an unchanged
  fixture generated by skelc v0.17.1 and linked with Vine v0.15.7.

## [0.19.2] - 2026-09-13

### Removed

- Remove `skelc version --features`, the optional `features` response field, and
  the Go `VersionFeaturesResult` type. Integrations should use compiler version
  requirements instead; the removed flag now returns `INVALID_ARGUMENT`.

### Changed

- Stop generating YAML tags on Go config fields. Vine loads config objects as
  JSON; Hub seed YAML handling does not depend on these tags.

### Fixed

- Allow enum values in config maps, including nullable enum values, consistently
  with enum fields and list elements.
- Release archives embed the exact tagged module version and clean source
  revision, and release builds refuse to overwrite existing binary assets.

## [0.19.1] - 2026-09-12

### Fixed

- Generate YAML tags for config object fields so their YAML names match the
  corresponding JSON tags, preserving camelCase field names during decoding.

## [0.19.0] - 2026-09-09

### Added

- Add `open service` for public client and server contracts. Go public packages include server interfaces and default implementations; split regular packages reuse them. `open`, `pub`, and `api` are mutually exclusive.

### Changed

- Raise the default and minimum Vine dependency for generated Go code to v0.15.5.

## [0.18.1] - 2026-09-09

### Changed

- Require explicit `api service` names to end with `ApiService`. Rename existing API contracts and regenerate server/client code together because service wire names change.

## [0.18.0] - 2026-09-09

### Added

- Add global `--strict`, `Input.Strict`, and LSP strict settings to reject migration warnings while preserving ordinary warnings.
- Add explicit `api service` contracts, mutually exclusive with `pub`, Portal-only invocation metadata, and migration warnings for unmodified services and legacy client rules.
- Add `--api` Go clients using `go.yorun.ai/vrpc` v0.12.0 and its shared Skel types. Prefix-derived modules use an `api` suffix; cross-domain client types reference API packages.
- Require `--api` for TypeScript generation and reject `--pub`. Both client targets include API dependencies and standalone public types.
- Collect public contract data dependencies without requiring local `pub` data/enum markers, preserving explicit cross-domain Skel visibility.

### Changed

- Raise both the default and minimum supported Vine versions for generated Go
  code to v0.15.4. Explicit `--go-vine-version` values below v0.15.4 are rejected.
- Generate typed Vine client calls with `InvokeAs`, removing intermediate result and error assertions.

### Fixed

- Detect nested generic binary payloads consistently in Go and TypeScript clients.
- Avoid local-name collisions with declared arguments in generated Go API and Vine clients.
- LSP refreshes sibling Skel files when opening a document or receiving file
  change notifications, recovering missing type definitions after partial
  generation notifications. Delayed deletion events retain files already
  recreated on disk, and unsaved editor contents remain authoritative.

## [0.17.1] - 2026-09-07

### Added

- Actor credential fields accept `string?` as well as `string`. Generated Go
  and public Skel preserve optionality. Credential blocks still
  require at least one non-nullable `string` field; definitions containing only
  `string?` fields and other field types are rejected. Portal
  credential parsing must be updated separately before optional fields can
  be omitted from gateway requests.

### Fixed

- LSP semantic diagnostics now analyze standalone Skel files independently when
  their directory has no `domain.skel`, avoiding false duplicate declarations
  between files with the same domain. Directories containing `domain.skel`
  continue to combine same-domain files. Schema compatibility checks use the
  matching standalone file at Git `HEAD` and resolve relative baseline paths
  from its containing directory.

## [0.17.0] - 2026-09-07

### Upgrade notes

- Generated Go code now requires Vine v0.15.1 or later; upgrade the runtime
  before regenerating contracts. Existing Skel declarations remain valid.
- Regenerate actor Info types and schemas to use `@identifier`. The marker is
  optional and supports one non-nullable string, uuid, or int field per actor.
- Schema consumers using strict decoding must be updated before reading the new
  `identifierField` or `noTrim` metadata emitted when those markers are used.

### Added

- Add optional `@identifier` on actor info string/uuid/int fields, schema and hash metadata, public Skel preservation, generated Go Info field tags, and LSP completion. Generated Go now requires Vine v0.15.1.

- Generated permission-check invocations now explicitly set `CodeArgumentName`.
  Business arguments may be named `code`; the injected string argument uses
  the first available name among `code`, `code1`, `code2`, and so on. Public
  contracts continue to omit the injected argument. Requires Vine v0.15.0.

- Config fields accept the argument-free `@noTrim` decorator to preserve string
  whitespace, including nullable strings, list elements, and map values.
  Generated Go uses `skel:"noTrim"` or `skel:"sensitive,noTrim"`. Public Skel,
  schema snapshots, hashes, and config-field completion preserve the marker;
  changing it is reported as `DANGEROUS` by schema diff. Runtime support requires
  Vine v0.15.0 or later.

### Changed

- Omit unused nil validation hooks from generated Go method registrations.

- Raised the minimum and default Vine dependency for generated Go modules to
  v0.15.1 for actor identity metadata, `index(n)`, and `noTrim` support.

- Generated Go service and resource-check argument fields use `skel:"index(n)"`
  instead of `arg:"n"`. Sensitive arguments combine attributes as
  `skel:"index(n),sensitive"`. This requires Vine v0.15.0 or later.

## [0.16.0] - 2026-09-07

### Changed

- Generated permission constants and resource-check code parameters now use Go
  `string`; generated runtime schemas describe permission-code values as strings.
  Permission codes use ordinary string types throughout generation.
- Added `model.Argument.Source` to distinguish runtime-injected permission codes
  from caller-supplied check arguments without relying on their type.

### Removed

- Removed the built-in `PermissionCode` contract type, its model/schema type kinds,
  and editor completion. Use `string` in contract fields; resource-check `code`
  parameters remain implicit.
- Removed generated resource `XxxPermissionCodes()` helpers, including public
  facade forwarding functions. Individual permission constants remain available.

### Upgrade Notes

- Regenerate affected packages together and update resource-check implementations
  to accept `string` instead of `skel.PermissionCode`. Remove calls to generated
  permission-list helpers; use schema metadata for permission discovery. Vine's
  existing `PermissionCode` type remains available for older generated packages.

## [0.15.0] - 2026-09-04

### Changed

- Raised the minimum Go version for skelc and generated Go modules to 1.27.0,
  configured CI checks to follow the latest Go 1.27 patch, and pinned release
  binaries to Go 1.27.1
- Generated nullable Go lists and maps as pointers so null remains distinct
  from an empty collection, and raised the minimum and default Vine version for
  generated Go modules to v0.14.0

### Removed

- Removed rolling compatibility with Go packages generated by skelc v0.11.x;
  imported generated data must now provide `Clone()` or `CloneBy(...)`
- Removed migration support for the legacy `.skelc-manifest.json` sidecar;
  generated output ownership is now determined only by in-file markers
- Removed generated data `Validate` methods and Rpc argument and result
  validation functions; collection nullability is now expressed by Go types

### Upgrade Notes

- Upgrade Go to 1.27.0 or later and Vine to v0.14.0 or later before consuming
  newly generated Go code. Existing modules must update their own dependencies;
  generated standalone modules default to Vine v0.14.0.
- Regenerate Go packages and coordinate updates to imported generated packages.
  Nullable `list<T>?` and `map<K, V>?` now map to `*[]T` and `*map[K]V`.
  Update assignments, collection access, and service implementations, and remove
  calls to generated `Validate` methods while retaining business validation.
- With the v0.15 encoding contract, nil collection pointers encode as `null`,
  while nil slices and maps encode as empty collections in JSON and CBOR.
  Non-nullable collection input `null` is no longer rejected by generated
  validation and encodes back as an empty collection. Skel declarations and
  TypeScript collection types do not need to change.
- For output still tracked only by `.skelc-manifest.json`, regenerate with
  skelc v0.14.x first so generated files carry in-file ownership markers.
  Review generated diffs and run consumer tests before deploying the upgrade.

## [0.14.1] - 2026-09-03

### Fixed

- Removed an unintended blank line before the `Hash` field in generated Vine
  runtime schema literals
- Release builds now validate the JSON-formatted `skelc version` result before
  uploading archives

## [0.14.0] - 2026-08-17

### Changed

- All non-LSP commands now emit one pretty-printed JSON result on stdout; text
  result modes and `--output-format` have been removed. `check` embeds all
  diagnostics in its result, generation reports completion while writing
  non-fatal diagnostics to stderr, and `schema get` returns JSON `null` when absent.
  Exit codes distinguish success (`0`), a completed unsatisfied check (`1`),
  and command failure (`2`); failures emit a stable `{code,message}` JSON object
  while stderr remains reserved for logs and diagnostics. stderr logs now
  default to JSONL; `--log-format text` selects human-readable text
- Added the public `go.yorun.ai/skelc/command` facade for command results, error
  codes, and exit-code constants

### Removed

- Removed the deprecated `symbol list/get` commands. Use `schema list/get` for
  declaration inspection

## [0.13.0] - 2026-08-17

### Added

- Added a unified `schema` command group for listing and inspecting declarations,
  producing deterministic versioned JSON schema snapshots, and diffing two Skel
  source inputs with `COMPATIBLE`, `DANGEROUS`, and `BREAKING` change classes.
  Diff items independently identify `ADDED`, `REMOVED`, and `MODIFIED`
  operations. Every schema subcommand emits pretty-printed JSON, and diff always
  returns the complete report without applying a failure threshold. When an
  explicit baseline is omitted, diff reads the candidate path from Git `HEAD`.
  The public `go.yorun.ai/skelc/schema` facade exposes strict snapshot
  `Encode`, `Decode`, and `Validate` functions plus typed constants for all
  normalized wire enums over the implementation in `internal/schema`. The root
  Go facade remains focused on parsing and generation.
  The existing `symbol list/get` commands remain available as deprecated
  compatibility entry points.
- Added schema compatibility support to `skelc lsp`: clients can enable live
  impact diagnostics against Git `HEAD` or an explicit source baseline, request
  CodeLens actions, and execute `skel.schema.diff` to retrieve the complete
  structured report for the current in-memory domain.

### Changed

- Extended rolling compatibility with Go packages generated by skelc v0.11.x
  through the v0.14.x release line. Regenerate imported packages before
  upgrading to v0.15.0, when imported generated data must provide `Clone()` or
  `CloneBy(...)`.
- Generated Vine runtime schema now adapts the same normalized projection used
  by schema inspection and compatibility diffing. Resolved local type
  references distinguish data, config, and event declarations, while Vine-only
  hashes and generated-service metadata remain runtime extensions.

## [0.12.0] - 2026-08-16

### Changed

- Generated non-generic Go data types now expose `Clone()`, while generic data
  types expose `CloneBy(...)` with one typed clone callback per type parameter.
  Rpc service specs compose these methods into typed request and result clone
  hooks that provide value isolation for Vine's in-process transport without
  promising JSON/CBOR or custom marshaling equivalence. Soft-recursive data is
  cloned structurally. During the v0.12.0 rolling upgrade, imported data uses
  its generated clone method when available; older non-generic packages fall
  back to serialization, while older generic packages use a typed structural
  clone so type-parameter callbacks still define isolation. Data fields named
  `clone` and `cloneBy` are now reserved for the stable generated methods
- skelc and generated Go modules now require Go 1.26.6; generated code now
  requires and defaults to `go.yorun.ai/vine v0.13.1`
- Updated `golang.org/x/mod` to v0.40.0
- Generated Go package documentation now declares the package fully managed by
  skelc and warns that unmanaged Go files are outside skelc and Vine's
  compilation, runtime, and compatibility guarantees

## [0.11.1] - 2026-08-03

### Changed

- Generated files now carry an in-file `Code generated by skelc. DO NOT EDIT.`
  ownership marker; stale output cleanup no longer creates or depends on a
  `.skelc-manifest.json` sidecar, and the first generation automatically
  migrates and removes a manifest created by v0.9.3 through v0.11.0

## [0.11.0] - 2026-08-02

### Added

- Public diagnostic types and stable code constants through
  `go.yorun.ai/skelc/diagnostic` and the root API
- `format --check` and machine-readable format results through
  `--output-format json`
- A comprehensive commerce example covering imports, generic data,
  configuration, events, resources and permissions, Web capabilities, and tasks

### Changed

- Compiler validation and compatibility-hash failures now propagate explicit
  errors instead of relying on panic-based checks
- Formatter lexer failures now propagate explicitly to CLI, LSP, and generated
  Skel callers

### Fixed

- Semantic diagnostic codes and naming suggestions are now assigned from
  structured analyzer metadata instead of parsed from human-readable messages
- `format` stages all changed files before committing them and restores files
  already replaced if a later write fails
- Formatter transactions preserve ownership, modes, supported extended
  metadata, and synchronize parent directories before reporting success
- Multiline block comments retain their relative indentation when rebased
- Multi-target generation stages every output before committing and restores
  all affected targets when a commit fails
- `@deprecated` completion inserts its required reason and uses a snippet
  placeholder when the language client supports snippets
- `skelc lsp` now uses normal CLI dispatch so global flags, help output,
  injected standard streams, and server failures behave consistently

## [0.10.3] - 2026-07-30

### Fixed

- Language-server semantic analysis now matches `check`: it merges same-domain
  files only within one source directory, leaves imports unresolved, and keeps
  same-named domain instances in separate directories independent
- Decorator completion now filters and replaces prefixes after `@` correctly
  and only suggests decorators accepted for the following declaration, block,
  field, or argument; decorators already present on the same target are omitted

## [0.10.2] - 2026-07-30

### Added

- Context-aware completion for actor transports and config lifecycle values
- Remote workspace URI preservation and dynamic workspace-folder indexing in the language server

## [0.10.1] - 2026-07-29

### Added

- `@deprecated("reason")` metadata for declarations and their deprecatable
  elements, including generated Vine schemas, Go docs, TypeScript JSDoc,
  public Skel output, and LSP presentation
- UUID map keys in contracts, generated Go maps, TypeScript records, and public Skel output

### Changed

- Generated Go code now requires at least `go.yorun.ai/vine v0.10.1`, which is
  also the default dependency version for generated Go modules
- Documentation links now target the public Skel site and editor-support repository

## [0.10.0] - 2026-07-26

### Added

- `@sensitive` metadata across Skel contracts, generated schemas and code, and
  public contract projection

### Fixed

- Release binaries are built from the checked-out release tag

## [0.9.5] - 2026-07-23

### Added

- Complete transitive Skel import-graph analysis

### Fixed

- Normalize imported enum, data, generic, and named-type references before code generation

## [0.9.4] - 2026-07-22

### Fixed

- Harden formatter and diagnostic edge cases
- Suggest corrected names for invalid identifiers

## [0.9.3] - 2026-07-22

### Added

- LSP document formatting, keyword and type completion, and declaration hover details
- Hierarchical document symbols, workspace symbol search, and top-level declaration rename
- Best-effort domain, import, and top-level declaration indexing while a document has syntax errors
- Debounced workspace-wide semantic diagnostics over unsaved documents, including same-domain file merging and cross-domain validation
- Collection of multiple compiler diagnostics in a single run

### Changed

- Compiler validation aborts now carry structured error codes, source positions, and wrapped causes through centralized API and CLI recovery boundaries
- Analyzer validation now reports errors explicitly instead of using panic/recover control flow; `check` and LSP collect up to 50 independent diagnostics per domain while suppressing errors that only depend on invalid declarations

### Fixed

- Prevent analyzer panics when syntax recovery produces an incomplete nested permission expression
- Preserve formatter idempotence and relative indentation for multiline comments and triple-quoted strings
- Calculate parser diagnostic and LSP ranges correctly when non-ASCII characters precede a token on the same line

## [0.9.2] - 2026-07-22

### Added

- Automated publication of release binaries

### Changed

- Documentation now links to the independent editor-support repository

## [0.9.1] - 2026-07-21

### Added

- Language Server Protocol support for syntax diagnostics, document symbols,
  definitions, and references

## [0.9.0] - 2026-07-21

Initial public release.

### Included

- Skel parsing, validation, formatting, and symbol inspection
- Go source and standalone Go module generation
- TypeScript type, package, and vRPC service client generation
- Public Skel contract extraction for cross-domain sharing
- Binary-aware sparse vRPC wire-schema generation for TypeScript clients

[Unreleased]: https://github.com/yorun-ai/skelc/compare/v0.24.2...HEAD
[0.24.2]: https://github.com/yorun-ai/skelc/compare/v0.24.1...v0.24.2
[0.24.1]: https://github.com/yorun-ai/skelc/compare/v0.24.0...v0.24.1
[0.24.0]: https://github.com/yorun-ai/skelc/compare/v0.23.3...v0.24.0
[0.23.3]: https://github.com/yorun-ai/skelc/compare/v0.23.2...v0.23.3
[0.23.2]: https://github.com/yorun-ai/skelc/compare/v0.23.1...v0.23.2
[0.23.1]: https://github.com/yorun-ai/skelc/compare/v0.23.0...v0.23.1
[0.23.0]: https://github.com/yorun-ai/skelc/compare/v0.22.1...v0.23.0
[0.22.1]: https://github.com/yorun-ai/skelc/compare/v0.22.0...v0.22.1
[0.22.0]: https://github.com/yorun-ai/skelc/compare/v0.21.0...v0.22.0
