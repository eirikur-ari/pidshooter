#!/usr/bin/env bash
# Prints the CHANGELOG.md section of RELEASE_VERSION, failing if it is missing or empty.
# Requires RELEASE_VERSION in the environment.
set -euo pipefail

if [[ -z "${RELEASE_VERSION:-}" ]]; then
  echo "release version is missing or empty" >&2
  exit 2
fi

awk -v version="$RELEASE_VERSION" '
  /^## \[/ { if (found) exit; found = (index($0, "## [" version "]") == 1); next }
  found && /^\[[^]]+\]: / { exit }
  found { body = body $0 "\n" }
  END {
    if (body !~ /[^[:space:]]/) { print "CHANGELOG.md has no entries for " version > "/dev/stderr"; exit 1 }
    printf "%s", body
  }' CHANGELOG.md
