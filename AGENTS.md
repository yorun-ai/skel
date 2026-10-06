# Skel Agent Guidelines

## Go Version and Syntax

- Target Go 1.27 syntax. Prefer `new` with a composite literal when creating a pointer, for example: `option := new(SomeOption{Field: value})`.
- Use `kind` when `type` would otherwise be the natural local variable name.
- Prefix unexported package-local production type declarations with `_`, such as `_Parser` and `_Option`. This applies only to types; do not prefix unexported constants, variables, functions, or methods with `_`. Test fixture types may use descriptive lowercase names.
- Use `Rpc`, not `RPC`, in identifiers and generated Go APIs.

## Architecture Boundaries

- `cmd/skelc/output` exposes CLI result, error and exit-code contracts through `internal/cmd/skelc/output`; keep command execution in `internal/cmd/skelc`.
- `cmd/skelc` is the executable entry point; keep it thin and delegate CLI behavior to `internal/cmd/skelc`.
- `internal/cmd/skelc` owns command definitions, flag-specific validation, terminal output, and exit codes. Generation commands call the public `api` package; input normalization, target-option normalization, and output-directory lifecycle must not be duplicated in CLI code.
- `internal/compiler` is the source-compilation orchestration package; CLI implementation and output contracts belong to `internal/cmd/skelc`, while shared option validation lives in `internal/optionvalidation`. LSP, formatting and generation remain independent capabilities.
- Keep source loading, syntax parsing, semantic analysis and compatibility hashing separate. `internal/compiler` coordinates them and owns recovery, diagnostics, imports and incremental analysis; `internal/model` remains parser-independent.
- `internal/symbol` owns shared declaration identities, scopes and name resolution for semantic analysis and language tooling, including recovered syntax. Keep semantic validation in `internal/analyzer`; target-language bindings belong to `internal/codegen/binding`.
- `internal/source` owns immutable document revisions and byte locations; `internal/loader` owns input discovery and filesystem, memory, and Git providers.
- `internal/schema` owns shared contract types, querying, encoding, validation and contract diffing; it must not depend on compiler models, parsing or input loading. `internal/projection` converts semantic models into the shared contract. `internal/sourcediff` coordinates compilation and source baselines for CLI and LSP comparisons. `internal/codegen/binding/golang/vineschema` adapts the pure projection to Vine with runtime metadata only.
- Target generators live in `internal/codegen/binding/{golang,skeleton,typescript}`. `internal/codegen` owns SDK inputs, declaration selection, type traversal and execution. `internal/codegen/binding` owns shared rendering and import helpers and must not depend on a target generator. `internal/codegen/output` owns generated-file markers and managed multi-target transactions; it must not depend on the SDK or binding implementations.
- Target-language naming constraints belong to the corresponding binding. In particular, Go validates sensitive marker field/method collisions only for emitted structures; `skelSensitive` is not a language-level reserved field.
- Keep module metadata generation in its target generator package. TypeScript rendering and template payloads belong to `internal/codegen/binding/typescript`; Go source rendering and Vine schema adaptation retain their separate boundaries.
- `internal/lsp/workspace` owns document indexing, workspace state and immutable snapshots; analysis scheduling and language features consume those snapshots.
- `internal/location` owns shared source-position values used by schema, semantic models and diagnostics. Reusable helpers stay in `internal/util`; language capabilities remain peer packages under `internal`, regardless of which tools consume them.
- Semantic failure values and reference-cycle algorithms belong to `internal/analyzer`; terminal log formatting belongs to `internal/cmd/skelc`. Do not extract capability-local implementation details into generic utility packages.
- Prefer cohesive packages with responsibility-specific files. Add a package only for a meaningful dependency, ownership or reuse boundary, not for each implementation stage or helper.
- `internal/formatter` owns pure Skel source formatting. The CLI owns in-place formatting and must validate all applicable inputs before writing files so a failed operation does not leave a partially updated source tree.
- Keep implementation packages under `internal` unless they form part of the supported programmatic API. The public `api` package is a facade over `internal/api`, which owns source-input adaptation, source inspection and cross-capability orchestration. Target-option normalization belongs to target generators; `codegen` exposes the shared generator SDK and managed execution boundary. The API implementation composes the core pipeline and independent generators; the core must not depend on this adapter. The root `skel` package does not expose toolchain APIs. `model` exposes parser-independent semantic data. `codegen.Input` borrows that graph read-only and supplies generation selection, lookup and traversal; do not duplicate declarations in a separate generator model or write target-language metadata back to the graph. Public facade packages remain limited to aliases, constants, and narrowly scoped function forwarding to their matching implementation package.

## Language and Compatibility

- Treat the Skel grammar, accepted legacy syntax, diagnostics, CLI flags, exit codes, JSON/JSONL fields, generated filenames, generated APIs, and generated module metadata as public compatibility boundaries.
- When changing Skel syntax, coordinate affected grammar, semantic model, formatter, generators and tests. Check the language and CLI references for descriptions made inaccurate by the change.
- When changing generated code, update every affected language backend and golden or structural tests. Confirm that generated Go code remains compatible with the declared Vine version.
- Keep deterministic behavior: input discovery, symbols, imports, dependencies, diagnostics, and generated files must have stable ordering.
- Do not add silent recovery for invalid contracts. Diagnostics should identify the relevant source path and location whenever available.

## Generated Artifacts

- Modify generator templates under the relevant `internal/codegen/binding/{golang,skeleton,typescript}` package rather than patching expected generated output behavior elsewhere.
- Editor integrations live in the independent `yorun-ai/skel-editor-support` repository. Keep editor client code and Marketplace packaging out of skelc; coordinate LSP compatibility across the two repositories.

## Documentation

- Use lowercase `portal`, `web`, `hub`, and `link` in prose, comments, and CLI help. Preserve the spelling of actual code identifiers, GoDoc declaration names, and language examples.

- Read the relevant `skel-site` references before changing syntax, CLI behavior or generated output. Correct existing descriptions made inaccurate by a change and document new user-facing features; internal changes and fixes restoring documented behavior need no new site content.
- Keep `README.md` and `README.zh-CN.md` synchronized, including language-switch links, commands, compatibility notes, and license information.
- `skel-site/docs/language/syntax.md` is the detailed English Skel language
  reference; `skel-site/docs/reference/cli.md` is the detailed English CLI
  reference. Keep their Simplified Chinese translations under
  `skel-site/i18n/zh-CN/docusaurus-plugin-content-docs/current` synchronized.
- Keep examples executable against the current CLI and syntax. Avoid documenting planned commands or unsupported flags.

- Before release, complete and validate coordinated documentation edits in the
  owning workspace. Do not commit or merge documentation-site changes as part
  of a code release without separate explicit authorization; documentation-site
  commits and merges are not release prerequisites.

## Changelog

- Do not edit `CHANGELOG.md` in ordinary commits or pull requests, even when the change is user-visible. Leave the changelog untouched during implementation, fixes, refactors and follow-up work.
- Write entries only in the release preparation pull request (`chore(release): prepare vX.Y.Z`). That pull request adds the dated `## [X.Y.Z] - YYYY-MM-DD` heading and records every user-visible change merged since the previous release.
- Derive the entries from the merged commits and pull requests in the release range. Writing them at release time keeps reverted or reworked changes from leaving stale entries in the changelog.

## Release Publication

- Merge the release-preparation PR after required CI passes, sync local `main`, then create and push the version tag from that reviewed commit.
- Pushing a `v*` tag triggers binary builds; do not manually publish a GitHub Release first. Release events do not trigger builds.
- The workflow validates main ancestry and the dated CHANGELOG entry, builds four Linux/macOS AMD64/ARM64 archives from the clean tag checkout, creates a Draft Release with changelog notes, and verifies downloaded attachments and SHA-256 before publishing.
- Retry failed runs or dispatch with the same existing tag. Unpublished Draft attachments may be replaced as a complete build set; published Releases must never be overwritten. Do not move a tag to repair a failed publication.
- Report publication complete only after the workflow succeeds and the Release and verified attachments are available. See `.github/CI.md` for validation and recovery.

## Tests

- Keep implementation tests paired with their source files. Shared setup may live in a narrowly scoped test helper file.
- Name unit tests `<source>_test.go`. For larger suites or integration scenarios, use `<entrypoint>_<scenario>_test.go`; name tests after the current entrypoint rather than a removed API.
- Use `<subject>_benchmark_test.go` and `<subject>_fuzz_test.go` for dedicated benchmark and fuzz files. Keep shared setup in `test_helper_test.go` or `<subject>_helper_test.go`; avoid vague names such as `performance_test.go` and `api_util_test.go`.
- Preserve tests for currently supported compatibility behavior. Remove retired-feature assertions only after confirming the support boundary, and exercise production entrypoints instead of recreating removed production pipelines in test helpers.
- Use `t.TempDir` for filesystem tests and `t.Cleanup` to restore modified globals or environment variables.
- Do not write test output into repository source directories.
- Add parser and formatter coverage for whitespace, comments, source locations, invalid input, and round trips when relevant.
- Add generator coverage for deterministic output and all affected declaration kinds when changing templates or rendering behavior.

## Validation

- Run `gofmt` on changed Go files and run `git diff --check`.
- Run targeted package tests while iterating, then run `GOWORK=off go test ./...` for repository-wide Go changes so an enclosing workspace cannot replace published dependencies.
- Run `GOWORK=off go vet ./...` after changes involving exported APIs, reflection, filesystem safety, or CLI/runtime wiring.
- Run `pnpm build` in `skel-site` after changing skelc user-facing documentation there.
- For CLI, syntax, or generator changes, exercise at least one representative `skelc check` or `skelc gen` flow in addition to automated tests.
