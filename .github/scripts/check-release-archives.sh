#!/usr/bin/env bash
# Checks that build/dist/ holds exactly the four archives built for RELEASE_VERSION, each with the third-party licenses and the README images.
# Requires RELEASE_VERSION in the environment.
set -euo pipefail

if [[ -z "${RELEASE_VERSION:-}" ]]; then
  echo "release version is missing or empty" >&2
  exit 2
fi

count=$(find build/dist -maxdepth 1 -name "pidshooter_${RELEASE_VERSION}_*.tar.gz" | wc -l | tr -d ' ')
total=$(find build/dist -maxdepth 1 -name '*.tar.gz' | wc -l | tr -d ' ')

if [[ "$count" != "4" || "$total" != "4" ]]; then
  echo "expected 4 archives named for version $RELEASE_VERSION, found $count of $total"
  exit 1
fi

for archive in build/dist/*.tar.gz; do
  listing=$(tar -tzf "$archive")
  if ! grep -qx 'THIRD_PARTY_LICENSES' <<< "$listing"; then
    echo "$archive has no THIRD_PARTY_LICENSES"
    exit 1
  fi
  if ! grep -q '^docs/images/.' <<< "$listing"; then
    echo "$archive has no docs/images"
    exit 1
  fi
done
