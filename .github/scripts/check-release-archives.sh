#!/usr/bin/env bash
# Checks that build/dist/ holds exactly the EXPECTED_ARCHIVE_COUNT archives built for RELEASE_VERSION, each with the third-party licenses and the readme images.
# Requires RELEASE_VERSION and EXPECTED_ARCHIVE_COUNT in the environment.
set -euo pipefail

# Fails when a required environment variable is missing or empty
function setup() {
  local missing=() name
  for name in RELEASE_VERSION EXPECTED_ARCHIVE_COUNT; do
    [[ -n "${!name:-}" ]] || missing+=("$name")
  done
  if (( ${#missing[@]} > 0 )); then
    echo "missing or empty environment variables: ${missing[*]}" >&2
    exit 2
  fi
}

# Fails unless build/dist holds exactly the expected number of archives, all of them named for the release version
function checkArchiveCount() {
  local count total

  count=$(find build/dist -maxdepth 1 -name "pidshooter_${RELEASE_VERSION}_*.tar.gz" | wc -l | tr -d ' ')
  total=$(find build/dist -maxdepth 1 -name '*.tar.gz' | wc -l | tr -d ' ')

  if [[ "$count" != "$EXPECTED_ARCHIVE_COUNT" || "$total" != "$EXPECTED_ARCHIVE_COUNT" ]]; then
    echo "expected $EXPECTED_ARCHIVE_COUNT archives named for version $RELEASE_VERSION, found $count of $total"
    exit 1
  fi
}

# Fails when an archive in build/dist lacks the third-party licenses or the readme images
function checkArchiveContents() {
  local archive entries

  for archive in build/dist/*.tar.gz; do
    entries=$(tar -tzf "$archive")
    if ! grep -qx 'THIRD_PARTY_LICENSES' <<< "$entries"; then
      echo "$archive has no THIRD_PARTY_LICENSES"
      exit 1
    fi
    if ! grep -q '^docs/images/.' <<< "$entries"; then
      echo "$archive has no docs/images"
      exit 1
    fi
  done
}

# Checks the release archives in build/dist
function main() {
  setup
  checkArchiveCount
  checkArchiveContents
}

main
