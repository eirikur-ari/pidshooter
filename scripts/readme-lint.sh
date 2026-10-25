#!/usr/bin/env bash
# Runs one README check: go-badge or release-badge.
# release-badge reads RELEASE_VERSION from the environment when it is set.
set -euo pipefail

go_badge() {
  echo "Checking README Go version badge"
  local go_version
  go_version=$(awk '/^go /{print $2}' go.mod)
  if ! grep -qF "badge/Go-$go_version-" README.md; then
    echo "README.md Go version badge does not match go $go_version in go.mod"
    exit 1
  fi
}

release_badge() {
  echo "Checking README release badge"
  local expected tag changelog_version
  if [[ -n "${RELEASE_VERSION:-}" ]]; then
    expected="$RELEASE_VERSION"
  else
    tag=$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || true)
    changelog_version=$(sed -n '/^## \[[0-9]/{s/^## \[\([^]]*\)\].*/\1/p;q;}' CHANGELOG.md)
    expected="${tag#v} $changelog_version"
  fi

  if [[ -z "${expected// /}" ]]; then
    echo "release version not found, it was neither provided as input nor found in a tag or the changelog"
    exit 1
  fi

  local version
  for version in $expected; do
    if grep -qF "badge/release-v$version-" README.md; then
      return 0
    fi
  done
  echo "README.md release badge does not show any of: $expected"
  exit 1
}

case "${1:-}" in
  go-badge) go_badge ;;
  release-badge) release_badge ;;
  *)
    echo "expected one argument, the README check to run: go-badge or release-badge"
    exit 2
    ;;
esac
