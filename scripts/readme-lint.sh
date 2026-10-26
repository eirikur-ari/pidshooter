#!/usr/bin/env bash
# Runs one readme badge lint: go-badge or release-badge.
# release-badge reads RELEASE_VERSION from the environment when it is set.
set -euo pipefail

# Fails when the Go version badge does not match the go version in go.mod
function goBadge() {
  local go_version

  echo "Linting readme Go version badge"
  go_version=$(awk '/^go /{print $2}' go.mod)
  if ! grep -qF "badge/Go-$go_version-" README.md; then
    echo "README.md Go version badge does not match go $go_version in go.mod"
    exit 1
  fi
}

# Prints the version of the latest release tag, or nothing when there is no release tag
function latestTagVersion() {
  local tag

  tag=$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)
  echo "${tag#v}"
}

# Prints the version of the first release in the changelog
function latestChangelogVersion() {
  sed -n '/^## \[[0-9]/{s/^## \[\([^]]*\)\].*/\1/p;q;}' CHANGELOG.md
}

# Prints the release versions the badge may show: the requested version, or else the versions of the latest tag and changelog release
# Fails when no release version is found
function resolveReleaseVersions() {
  local tag_version changelog_version versions

  if [[ -n "${RELEASE_VERSION:-}" ]]; then
    versions="$RELEASE_VERSION"
  else
    tag_version=$(latestTagVersion) || return
    changelog_version=$(latestChangelogVersion) || return
    versions="$tag_version $changelog_version"
  fi

  if [[ -z "${versions// /}" ]]; then
    echo "release version not found, it was neither provided as input nor found in a tag or the changelog" >&2
    exit 1
  fi

  echo "$versions"
}

# Succeeds when the readme release badge shows any of the given release versions
function isValidReleaseBadge() {
  local versions="$1" version

  for version in $versions; do
    if grep -qF "badge/release-v$version-" README.md; then
      return 0
    fi
  done
  return 1
}

# Fails when the release badge shows none of the accepted release versions
function releaseBadge() {
  local versions

  echo "Linting readme release badge"
  versions=$(resolveReleaseVersions)
  if ! isValidReleaseBadge "$versions"; then
    echo "README.md release badge does not show any of: $versions"
    exit 1
  fi
}

# Runs the readme lint named by the first argument
function main() {
  case "${1:-}" in
    go-badge) goBadge ;;
    release-badge) releaseBadge ;;
    *)
      echo "expected one argument, the readme lint to run: go-badge or release-badge"
      exit 2
      ;;
  esac
}

main "$@"
