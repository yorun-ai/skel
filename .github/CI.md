# CI maintenance

Keep this guide named `CI.md`: GitHub prioritizes `.github/README.md` over the
root README when displaying the repository homepage.

This directory owns repository automation, not public documentation.

| Event | Workflow | Responsibility |
| --- | --- | --- |
| PR targeting any branch | `ci.yml` | Run the checks selected by changed inputs and verify the required gate |
| Push a `v*` tag / manual tag input | `release.yml` | Build and verify archives, then publish GitHub Release |
| Push to `main` | `cache.yml` | Populate Go build, module and test caches for later PR runs; no correctness gate |

## Pull Request Gate

`ci.yml` classifies pull request changes with `.github/scripts/ci.sh changes`,
then runs three independent jobs when the change affects code, dependencies,
examples or CI configuration. Markdown, `LICENSE` and issue-template-only
changes skip those jobs while the required gate still completes successfully:

| Job | Checks |
| --- | --- |
| Go static | Module metadata drift, `go vet`, Staticcheck correctness/simplification/unused-code checks and nilness |
| Go race | Full-repository tests with the race detector |
| Examples | Generate both examples and type-check their TypeScript clients against the pinned published vRPC runtime |

Generated backend Go modules are not compiled in Skel CI or cache warmup.
Generator tests check emitted declarations and module requirements without
loading Vine. Go API client compilation and checks against Skel's own public
types remain covered. Consumers compile and test generated backend code with
their selected runtime dependencies.

All Go commands use `GOWORK=off`. CI and cache warmup follow the latest Go
1.27 patch; release builds remain pinned to their validated toolchain.
Run `bash .github/scripts/ci.sh static` from the repository root to reproduce
the static gate locally. Both CI and cache warmup use this script, which pins
Staticcheck and the Go analysis tools through the separate `.github/go-tools`
module and its checksum file, without adding application dependencies. The
shared `golang.org/x/tools` dependency is pinned explicitly so both analyzers
can read compiler export data from current Go 1.27 patches.
The nilness analyzer catches redundant nil comparisons as well as definite nil
dereferences. Legacy LSP initialization fields have local, explained
Staticcheck exemptions so support for older clients remains covered by tests.

Race tests disable Go's automatic vet pass because the static job runs the full
vet check. `GORACE=atexit_sleep_ms=0` removes the race runtime's exit delay; it
does not disable race detection or change its failure exit code.

Keep `CI / Required Checks` as the single required gate. It runs even when
another job fails and requires all three jobs to succeed; failures,
cancellations and skipped jobs must not pass. `bash .github/scripts/ci.sh verify`
implements that contract, so keep the job list in the script aligned with the
jobs here.

There are no path filters or reduced package lists, so changes to templates and
fixtures retain full coverage. PR updates cancel older runs for that PR and do
not run again after merge. Each check has a 15-minute timeout; the final gate
has a 5-minute timeout.

## Main Cache Warmup

`cache.yml` runs on pushes to `main`. It selects its work with the same
classifier, then warms a `standard` and a `race` Go cache by running the
ordinary test suite, the static checks, the quickstart example and the race
suite. It publishes no artifacts and reports no required status.

Pull request jobs only restore caches; the warmup is the only writer, so PR runs
do not create cache entries for every commit. `./.github/actions/go-cache` owns
the cache layout: the key combines the toolchain version, the mode (`standard`
or `race`), `go.sum` and the commit, and the restore keys fall back to older
entries with the same toolchain and mode. The action restores both `GOMODCACHE`
and `GOCACHE`, so one warmup covers module downloads, test results and vet
analysis.

## Change Classification

`.github/scripts/ci.sh changes` reads the changed paths of the event range and
writes a `run-ci` flag to `$GITHUB_OUTPUT`. `pull_request` events diff from the
merge base; `push` events diff from the previous commit and fall back to the
empty tree for a first push.

Documentation-only changes (`*.md`, `*.mdx`, `LICENSE`, issue templates) select
no jobs. Everything else, including workflow and script changes, selects the
full job set.

## Validating Workflow Edits

Validate workflow changes with:

```bash
GOWORK=off go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 .github/workflows/*.yml
shellcheck .github/scripts/release*.sh
bash .github/scripts/release_test.sh
git diff --check
```

Exercise the classifier locally before relying on it:

```bash
CHANGE_BASE=$(git rev-parse HEAD~1) CHANGE_HEAD=$(git rev-parse HEAD) \
  GITHUB_EVENT_NAME=push GITHUB_OUTPUT=/dev/stdout bash .github/scripts/ci.sh changes
```

When editing a job's commands, also execute those commands locally and keep
generated example output outside the repository.

The Examples job uses Node.js 24 and `.github/scripts/typecheck-generated.mjs`.
Its isolated temporary workspace installs the exact versions and integrity hashes
from `.github/typescript/package-lock.json`; it never substitutes sibling runtime
source. Cross-domain example imports resolve through TypeScript paths. Update the
fixture lockfile deliberately when changing the tested runtime or TypeScript version.

## Release Publication and Recovery

Merge release preparation through PR CI, sync `main`, and push the version tag
from the reviewed commit. The workflow checks tag identity, main ancestry, and a
dated, nonempty CHANGELOG entry. Release publication events do not start builds.
The Go static job also runs release lifecycle regression tests and workflow lint.

Builds use a clean tag checkout and retain version, revision, and dirty-state
checks. All four Linux/macOS AMD64/ARM64 archives and `checksums.txt` must be ready
before a Draft Release is created. Notes come from the matching CHANGELOG entry;
prerelease tags produce prereleases. The workflow uploads the files, checks the
exact asset list, downloads them, compares the checksum file and verifies SHA-256,
then publishes the Release. PR build outputs are not release attachments.

Retry a failed run or dispatch using its existing tag. Builds failing before
upload do not create a new Release; upload/verification failures leave a Draft.
A retry rebuilds the full set and replaces only the expected Draft attachments.
Unexpected attachments fail verification and require inspection. Already
published Releases are rejected without modifying their attachments. Do not
manually publish the Draft or move the tag to recover. A real version-tag run is
required to validate hosted permissions and cross-platform publication end to end.
