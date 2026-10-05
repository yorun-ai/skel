#!/usr/bin/env bash
set -euo pipefail

# Listing releases includes drafts; API/authentication failures must not be
# mistaken for a release that does not exist.
release_state() {
  gh api --paginate --slurp "repos/$GITHUB_REPOSITORY/releases?per_page=100" |
    jq -er --arg tag "$RELEASE_TAG" '
      [.[][] | select(.tag_name == $tag)] |
      if length == 0 then "missing"
      elif length != 1 then error("Duplicate release tag")
      elif .[0].draft then "draft" else "published" end'
}

release_main() {
  : "${RELEASE_TAG:?}" "${GITHUB_REPOSITORY:?}" "${RUNNER_TEMP:?}"
  [[ "$RELEASE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]] || {
    echo 'Invalid release tag' >&2; return 1;
  }
  local notes="$RUNNER_TEMP/skelc-release-notes.md" state
  case "${1:-}" in
    prepare)
      test "$(git rev-parse "refs/tags/$RELEASE_TAG^{commit}")" = "$(git rev-parse HEAD)"
      git merge-base --is-ancestor HEAD origin/main
      awk -v prefix="## [${RELEASE_TAG#v}] - " '
        index($0, prefix) == 1 {
          date = substr($0, length(prefix) + 1)
          if (date !~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]$/) exit 1
          found = 1; next
        }
        found && /^## / { exit }
        found { print; if ($0 ~ /[^[:space:]]/) body = 1 }
        END { if (!found || !body) exit 1 }
      ' CHANGELOG.md > "$notes"
      state=$(release_state)
      if [[ "$state" == published ]]; then
        echo 'Release is already published; refusing to replace assets' >&2
        return 1
      fi
      ;;
    publish)
      : "${RELEASE_DIST:?}"
      test -s "$notes"
      local version="${RELEASE_TAG#v}" target prerelease=false
      local assets=() expected downloaded
      for target in darwin_amd64 darwin_arm64 linux_amd64 linux_arm64; do
        assets+=("skelc_${version}_${target}.tar.gz")
      done
      expected=$(printf '%s\n' "${assets[@]}" checksums.txt | sort)
      (cd "$RELEASE_DIST" && sha256sum -c checksums.txt)
      state=$(release_state)
      if [[ "$state" == published ]]; then
        echo 'Release is already published; refusing to replace assets' >&2
        return 1
      fi
      [[ "$RELEASE_TAG" != *-* ]] || prerelease=true
      if [[ "$state" == missing ]]; then
        gh release create "$RELEASE_TAG" --repo "$GITHUB_REPOSITORY" --verify-tag \
          --draft --prerelease="$prerelease" --title "$RELEASE_TAG" --notes-file "$notes"
      else
        gh release edit "$RELEASE_TAG" --repo "$GITHUB_REPOSITORY" \
          --prerelease="$prerelease" --title "$RELEASE_TAG" --notes-file "$notes"
      fi
      # Only an unpublished draft can reach here. Replace its incomplete assets
      # as one build set, then download and verify before making it public.
      local paths=()
      for target in "${assets[@]}" checksums.txt; do paths+=("$RELEASE_DIST/$target"); done
      gh release upload "$RELEASE_TAG" "${paths[@]}" --clobber --repo "$GITHUB_REPOSITORY"
      downloaded=$(gh release view "$RELEASE_TAG" --repo "$GITHUB_REPOSITORY" \
        --json assets --jq '.assets[].name' | sort)
      test "$downloaded" = "$expected"
      local verify
      verify=$(mktemp -d "$RUNNER_TEMP/skelc-release-verify.XXXXXX")
      gh release download "$RELEASE_TAG" --repo "$GITHUB_REPOSITORY" --dir "$verify"
      cmp "$RELEASE_DIST/checksums.txt" "$verify/checksums.txt"
      (cd "$verify" && sha256sum -c checksums.txt)
      rm -rf "$verify"
      gh release edit "$RELEASE_TAG" --repo "$GITHUB_REPOSITORY" --verify-tag --draft=false
      ;;
    *) echo 'Usage: release.sh prepare|publish' >&2; return 1 ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then release_main "$@"; fi
