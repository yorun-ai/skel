#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'usage: ci.sh <changes|static|verify>\n' >&2
}

# Keep local validation, the PR gate and cache warmup on the same pinned checks.
static_checks() {
  export GOWORK=off
  go mod tidy -diff
  go vet ./...
  go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 -checks='SA*,S1*,U1000' ./...
  go run golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness@v0.47.0 ./...
}

# classify_changes reads NUL-delimited changed paths from stdin and reports the
# checks the change selects. Documentation-only changes skip the Go jobs.
classify_changes() {
  jq -Rse '
    split("\u0000") | map(select(length > 0)) |
    any(.[];
      (test("\\.(md|mdx)$") | not) and
      . != "LICENSE" and
      (startswith(".github/ISSUE_TEMPLATE/") | not)
    ) as $run_ci |
    {"run-ci": $run_ci}'
}

# verify_ci_results checks the required gate: every selected job succeeded and
# every unselected job is skipped.
verify_ci_results() {
  jq -e '
    . as $needs |
    all(["changes"][]; . as $job | $needs[$job].result == "success") and
    ($needs.changes.outputs["run-ci"] == "true" or $needs.changes.outputs["run-ci"] == "false") and
    all(["go-static", "go-race", "examples"][];
      . as $job |
      if $needs.changes.outputs["run-ci"] == "true"
      then $needs[$job].result == "success"
      else $needs[$job].result == "skipped"
      end) |
    if . then true else error("CI failed or contains an unexpected job result") end'
}

changes() {
  local base="${CHANGE_BASE:-}" head="${CHANGE_HEAD:-}"
  case "${GITHUB_EVENT_NAME:-}" in
    pull_request)
      [[ "$base" =~ ^[0-9a-f]{40}$ && "$head" =~ ^[0-9a-f]{40}$ ]] || {
        echo "Invalid pull request change range: base=$base head=$head" >&2
        return 1
      }
      base=$(git merge-base "$base" "$head")
      ;;
    push)
      [[ "$head" =~ ^[0-9a-f]{40}$ ]] || {
        echo "Invalid push change head: $head" >&2
        return 1
      }
      if [[ ! "$base" =~ ^[0-9a-f]{40}$ ]] || ! git cat-file -e "${base}^{commit}" 2>/dev/null; then
        base=$(git hash-object -t tree /dev/null)
      fi
      ;;
    *)
      echo "Unsupported event: ${GITHUB_EVENT_NAME:-}" >&2
      return 1
      ;;
  esac
  git diff --name-only --no-renames -z "$base" "$head" | classify_changes |
    jq -r 'to_entries[] | "\(.key)=\(.value)"' | tee -a "$GITHUB_OUTPUT"
}

case "${1:-}" in
  changes)
    changes
    ;;
  static)
    static_checks
    ;;
  verify)
    verify_ci_results <<< "${NEEDS:-}"
    ;;
  *)
    usage
    exit 2
    ;;
esac
