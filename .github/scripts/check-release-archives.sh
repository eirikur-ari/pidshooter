#!/usr/bin/env bash
# Checks that build/dist/ holds exactly the four archives built for RELEASE_VERSION.
# Requires RELEASE_VERSION in the environment.
set -euo pipefail

count=$(find build/dist -maxdepth 1 -name "pidshooter_${RELEASE_VERSION}_*.tar.gz" | wc -l | tr -d ' ')
total=$(find build/dist -maxdepth 1 -name '*.tar.gz' | wc -l | tr -d ' ')

if [[ "$count" != "4" || "$total" != "4" ]]; then
  echo "expected 4 archives named for version $RELEASE_VERSION, found $count of $total"
  exit 1
fi
