# CI maintenance

Keep this guide named `CI.md`: GitHub prioritizes `.github/README.md` over the
root README when displaying the repository homepage.

This directory owns repository automation, not public documentation.

| Event | Workflow | Responsibility |
| --- | --- | --- |
| PR targeting any branch | `ci.yml` | Run the checks selected by changed inputs and verify the required gate |
| Push to `main` | `cache.yml` | Populate Go build, module and test caches for later PR runs; no correctness gate |

## Pull Request Gate

`ci.yml` classifies pull request changes with `.github/scripts/ci.sh changes`,
then runs three independent jobs when the change affects code, dependencies,
examples or CI configuration. Markdown, `LICENSE` and issue-template-only
changes skip those jobs while the required gate still completes successfully:

| Job | Checks |
| --- | --- |
| Go static | Module metadata drift and full-repository `go vet` |
| Go race | Full-repository tests with the race detector |
| Examples | Generate both examples, compile their Go modules, and type-check their TypeScript clients against the pinned published vRPC runtime |

All Go commands use `GOWORK=off`, including commands in generated modules.
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
GOWORK=off go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 .github/workflows/ci.yml .github/workflows/cache.yml
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
