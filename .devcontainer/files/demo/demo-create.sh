#!/usr/bin/env bash
# Creates dummy processes named demo-* to play pidshooter against, for example
# when recording a demo: pidshooter play demo-
# The pattern demo- matches only these processes, never real applications.
# Each one is a copy of sleep, because a process is named after its executable
# (on macOS the copy must be re-signed or the system kills it).
# Remove them with demo-cleanup.sh.
set -euo pipefail

DEMO_DIR=/tmp/pidshooter-demo
LIFETIME_SECONDS=3600
NAMES=(demo-browser demo-editor demo-indexer demo-mailer demo-player demo-syncer demo-backup demo-daemon)

if pgrep -f "^${DEMO_DIR}/" > /dev/null; then
  echo "demo processes are already running, clean them up first with demo-cleanup.sh"
  exit 1
fi

mkdir -p "$DEMO_DIR"

for name in "${NAMES[@]}"; do
  cp "$(command -v sleep)" "${DEMO_DIR}/${name}"
  if command -v codesign > /dev/null; then
    codesign --force --sign - "${DEMO_DIR}/${name}" 2> /dev/null
  fi
  nohup "${DEMO_DIR}/${name}" "$LIFETIME_SECONDS" > /dev/null 2>&1 &
done

echo "created ${#NAMES[@]} demo processes, they exit on their own after ${LIFETIME_SECONDS}s:"
pgrep -fl "^${DEMO_DIR}/" || true
