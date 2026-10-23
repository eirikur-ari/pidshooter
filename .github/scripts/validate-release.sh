#!/usr/bin/env bash
# Checks that RELEASE_VERSION can be released from the current run.
# Requires RELEASE_VERSION, GITHUB_REF, GITHUB_REPOSITORY and GH_TOKEN in the environment.
set -euo pipefail

if [[ "$GITHUB_REF" != "refs/heads/master" ]]; then
  echo "a release must be started from master, not '$GITHUB_REF'"
  exit 1
fi

if ! echo "$RELEASE_VERSION" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  echo "version '$RELEASE_VERSION' is not MAJOR.MINOR.PATCH"
  exit 1
fi

if gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/v$RELEASE_VERSION" > /dev/null 2>&1; then
  echo "tag v$RELEASE_VERSION already exists"
  exit 1
fi

make release-notes-content > /dev/null
