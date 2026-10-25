#!/usr/bin/env bash
# Reports the vulnerability lint results in the file given as the first argument in
# the run summary, with a warning annotation.
# Requires GITHUB_STEP_SUMMARY in the environment.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "expected exactly one argument: the path of a text file with the vulnerability lint results, which are added to the run summary" >&2
  exit 2
fi

if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
  echo "missing or empty environment variable: GITHUB_STEP_SUMMARY" >&2
  exit 2
fi

report="$1"

{
  echo "### Vulnerability lint"
  echo
  echo '```'
  cat "$report"
  echo '```'
} >> "$GITHUB_STEP_SUMMARY"

echo "::warning title=Vulnerability lint failed::the vulnerability lint failed, see the run summary"
