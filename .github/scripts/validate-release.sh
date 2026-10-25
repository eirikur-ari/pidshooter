#!/usr/bin/env bash
# Checks that RELEASE_VERSION can be released from the current run.
# Requires RELEASE_VERSION, GITHUB_REF, GITHUB_REPOSITORY and GH_TOKEN in the environment.
set -euo pipefail

missing=()
for name in RELEASE_VERSION GITHUB_REF GITHUB_REPOSITORY GH_TOKEN; do
  [[ -n "${!name:-}" ]] || missing+=("$name")
done
if (( ${#missing[@]} > 0 )); then
  echo "missing or empty environment variables: ${missing[*]}" >&2
  exit 2
fi

if [[ "$GITHUB_REF" != "refs/heads/master" ]]; then
  echo "a release must be started from master, not '$GITHUB_REF'"
  exit 1
fi

if [[ ! "$RELEASE_VERSION" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo "version '$RELEASE_VERSION' is not MAJOR.MINOR.PATCH"
  exit 1
fi

if out=$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/v$RELEASE_VERSION" 2>&1); then
  echo "tag v$RELEASE_VERSION already exists"
  exit 1
elif [[ "$out" != *"HTTP 404"* ]]; then
  echo "$out" >&2
  exit 1
fi

make release-notes-content > /dev/null
