#!/usr/bin/env bash
# Reports the vulnerability check output in the file given as the first argument in
# the run summary, with a warning annotation.
# Requires GITHUB_STEP_SUMMARY in the environment.
set -euo pipefail

report="$1"

{
  echo "### Vulnerability check"
  echo
  echo '```'
  cat "$report"
  echo '```'
} >> "$GITHUB_STEP_SUMMARY"

echo "::warning title=Vulnerabilities found::the vulnerability check reported findings, see the run summary"
