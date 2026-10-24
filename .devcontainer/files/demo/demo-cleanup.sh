#!/usr/bin/env bash
# Removes the demo processes created by demo-create.sh, and their files.
# Safe to run when none are running.
set -euo pipefail

DEMO_DIR=/tmp/pidshooter-demo

pkill -f "^${DEMO_DIR}/" || true
rm -rf "$DEMO_DIR"

echo "demo processes cleaned up"
