#!/usr/bin/env bash
# Prints the CHANGELOG.md section of RELEASE_VERSION, failing if it is missing or empty.
# Requires RELEASE_VERSION in the environment.
set -euo pipefail

# Fails when the release version is missing or empty
function setup() {
  if [[ -z "${RELEASE_VERSION:-}" ]]; then
    echo "release version is missing or empty" >&2
    exit 2
  fi
}

# Prints the changelog section of the release version, failing when it has no entries
function printReleaseNotes() {
  awk -v version="$RELEASE_VERSION" '
    # True for a line that starts with: ## [
    function isVersionHeader(line)          { return line ~ /^## \[/ }

    # True for a line that starts with the requested version, for example: ## [0.3.0]
    function isRequestedVersionHeader(line) { return index(line, "## [" version "]") == 1 }

    # True for a link reference line from the footer of the changelog, for example:
    #   [0.3.0]: https://repourl/compare/v0.2.0...v0.3.0
    function isLinkReference(line)          { return line ~ /^\[[^]]+\]: / }

    # True for text that is empty or holds only whitespace
    function isBlank(text)                  { return text !~ /[^[:space:]]/ }

    # A version header ends the section being read, and starts reading when it is the requested version
    isVersionHeader($0) {
      if (inSection) exit
      inSection = isRequestedVersionHeader($0)
      next
    }

    # The link references of the footer come after the last section, so reading ends there
    inSection && isLinkReference($0) { exit }

    # Every other line of the requested section is part of the release notes
    inSection { notes = notes $0 "\n" }

    # Fails when the requested section is missing or empty, otherwise prints the release notes without surrounding blank lines
    END {
      if (isBlank(notes)) { print "CHANGELOG.md has no entries for " version > "/dev/stderr"; exit 1 }
      sub(/^[[:space:]]+/, "", notes)
      sub(/[[:space:]]+$/, "", notes)
      print notes
    }' CHANGELOG.md
}

# Prints the release notes of the release version from the changelog
function main() {
  setup
  printReleaseNotes
}

main
