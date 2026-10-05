#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=.github/scripts/release.sh
source "$(dirname "${BASH_SOURCE[0]}")/release.sh"
test_dir=$(mktemp -d)
trap 'rm -rf "$test_dir"' EXIT
export RUNNER_TEMP="$test_dir" GITHUB_REPOSITORY=example/skelc RELEASE_TAG=v1.2.3
export RELEASE_DIST="$test_dir/dist"
mkdir -p "$RELEASE_DIST" "$test_dir/remote"
cd "$test_dir"
cat > CHANGELOG.md <<'NOTES'
## [Unreleased]
## [1.2.3] - 2026-10-05
### Changed
- Release fixture.
## [1.2.2] - 2026-10-04
- Must not leak.
NOTES

# GitHub and Git are replaced; no network access or real release writes.
git() {
  case "$1" in
    rev-parse) echo abc ;;
    merge-base) [[ "${OFF_MAIN:-false}" == false ]] ;;
    *) return 1 ;;
  esac
}
gh() {
  local state
  state=$(cat "$test_dir/state")
  case "$1 $2" in
    'api --paginate')
      if [[ "${API_FAIL:-false}" == true ]]; then return 1; fi
      case "$state" in
        missing) echo '[[]]' ;;
        draft) echo '[[{"tag_name":"v1.2.3","draft":true}]]' ;;
        published) echo '[[{"tag_name":"v1.2.3","draft":false}]]' ;;
      esac ;;
    'release create') echo draft > "$test_dir/state" ;;
    'release edit')
      if [[ " $* " == *' --draft=false '* ]]; then echo published > "$test_dir/state"; fi ;;
    'release upload')
      [[ "$state" == draft ]] || return 1
      cp "$RELEASE_DIST"/* "$test_dir/remote/"
      if [[ "${UPLOAD_FAIL:-false}" == true ]]; then return 1; fi ;;
    'release view')
      printf '%s\n' "$test_dir/remote/"* | sed 's|.*/||' ;;
    'release download')
      local dest="${!#}"
      cp "$test_dir/remote/"* "$dest/"
      if [[ "${CORRUPT:-false}" == true ]]; then echo corrupt >> "$dest/skelc_1.2.3_linux_amd64.tar.gz"; fi ;;
    *) echo "Unexpected gh call: $*" >&2; return 1 ;;
  esac
}
expect_failure() {
  local result
  set +e
  (set -e; release_main "$1") > "$test_dir/failure.log" 2>&1
  result=$?
  set -e
  if [[ "$result" == 0 ]]; then
    echo "Expected $1 to fail" >&2; exit 1
  fi
}

echo missing > "$test_dir/state"
release_main prepare
grep -Fx -- '- Release fixture.' "$RUNNER_TEMP/skelc-release-notes.md"
if grep -q 'Must not leak' "$RUNNER_TEMP/skelc-release-notes.md"; then exit 1; fi
OFF_MAIN=true expect_failure prepare
API_FAIL=true expect_failure prepare
RELEASE_TAG=bad expect_failure prepare
RELEASE_TAG=v9.9.9 expect_failure prepare
echo published > "$test_dir/state"
expect_failure prepare
expect_failure publish

for target in darwin_amd64 darwin_arm64 linux_amd64 linux_arm64; do
  printf '%s\n' "$target" > "$RELEASE_DIST/skelc_1.2.3_$target.tar.gz"
done
(cd "$RELEASE_DIST" && sha256sum ./*.tar.gz > checksums.txt)
echo missing > "$test_dir/state"
release_main prepare
UPLOAD_FAIL=true expect_failure publish
test "$(cat "$test_dir/state")" = draft
CORRUPT=true expect_failure publish
test "$(cat "$test_dir/state")" = draft
touch "$test_dir/remote/unexpected.txt"
expect_failure publish
rm "$test_dir/remote/unexpected.txt"
# An incomplete draft is repaired by re-uploading the complete build set.
rm "$test_dir/remote/skelc_1.2.3_linux_arm64.tar.gz"
release_main publish
test "$(cat "$test_dir/state")" = published
expect_failure publish
echo 'Release lifecycle checks passed'
