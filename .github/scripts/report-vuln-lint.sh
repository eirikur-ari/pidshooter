#!/usr/bin/env bash
# Reports the vulnerability lint results in the file given as the first argument in
# the run summary, with a warning annotation.
# Requires GITHUB_STEP_SUMMARY in the environment.
set -euo pipefail

# Fails when the argument count is not one or the required environment variable is missing or empty
function setup() {
  if [[ $# -ne 1 ]]; then
    echo "expected exactly one argument: the path of a text file with the vulnerability lint results, which are added to the run summary" >&2
    exit 2
  fi

  if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
    echo "missing or empty environment variable: GITHUB_STEP_SUMMARY" >&2
    exit 2
  fi
}

# Adds the vulnerability lint results in the given file to the run summary
function writeSummary() {
  local report="$1"

  {
    echo "### Vulnerability lint"
    echo
    echo '```'
    cat "$report"
    echo '```'
  } >> "$GITHUB_STEP_SUMMARY"
}

# Prints the warning annotation of the failed vulnerability lint
function annotateWarning() {
  echo "::warning title=Vulnerability lint failed::the vulnerability lint failed, see the run summary"
}

# Reports the vulnerability lint results in the run summary
function main() {
  setup "$@"

  writeSummary "$1"
  annotateWarning
}

main "$@"
