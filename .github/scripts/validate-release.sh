#!/usr/bin/env bash
# Checks that RELEASE_VERSION can be released from the current run.
# Requires RELEASE_VERSION, GITHUB_REF, GITHUB_REPOSITORY and GH_TOKEN in the environment.
set -euo pipefail

# Fails when a required environment variable is missing or empty
function setup() {
  local missing=() name
  for name in RELEASE_VERSION GITHUB_REF GITHUB_REPOSITORY GH_TOKEN; do
    [[ -n "${!name:-}" ]] || missing+=("$name")
  done
  if (( ${#missing[@]} > 0 )); then
    echo "missing or empty environment variables: ${missing[*]}" >&2
    exit 2
  fi
}

# Fails unless the run was started from master
function checkBranch() {
  if [[ "$GITHUB_REF" != "refs/heads/master" ]]; then
    echo "a release must be started from master, not '$GITHUB_REF'"
    exit 1
  fi
}

# Fails unless the release version is MAJOR.MINOR.PATCH
function checkVersionFormat() {
  if [[ ! "$RELEASE_VERSION" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
    echo "version '$RELEASE_VERSION' is not MAJOR.MINOR.PATCH"
    exit 1
  fi
}

# Fails when the release tag already exists or its lookup fails for any reason other than the tag not being found
function checkTagAbsent() {
  local tag_ref

  if tag_ref=$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/v$RELEASE_VERSION" 2>&1); then
    echo "tag v$RELEASE_VERSION already exists"
    exit 1
  elif [[ "$tag_ref" != *"HTTP 404"* ]]; then
    echo "$tag_ref" >&2
    exit 1
  fi
}

# Fails when the changelog has no release notes for the release version
function checkReleaseNotes() {
  make release-notes-content > /dev/null
}

# Checks that the release version can be released from the current run
function main() {
  setup
  checkBranch
  checkVersionFormat
  checkTagAbsent
  checkReleaseNotes
}

main
