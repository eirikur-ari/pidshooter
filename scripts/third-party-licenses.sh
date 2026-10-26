#!/usr/bin/env bash
# Writes the licenses of the Go runtime and of the third-party modules compiled into the binary to PACKAGE_DIR/THIRD_PARTY_LICENSES.
# Requires WORKSPACE_FOLDER and PACKAGE_DIR in the environment.
set -euo pipefail

# Fails when a required environment variable is missing or empty, then moves to the workspace and creates the output directory
function setup() {
  local missing=() name
  for name in WORKSPACE_FOLDER PACKAGE_DIR; do
    [[ -n "${!name:-}" ]] || missing+=("$name")
  done
  if (( ${#missing[@]} > 0 )); then
    echo "missing or empty environment variables: ${missing[*]}" >&2
    exit 2
  fi

  cd "$WORKSPACE_FOLDER"
  mkdir -p "$PACKAGE_DIR"
}

# Saves the license files of the third-party modules compiled into the binary to the given directory
function saveLicenses() {
  local licenses_dir="$1"

  go run github.com/google/go-licenses/v2@v2.0.1 save ./cmd/pidshooter --ignore github.com/eirikur-ari/pidshooter --force --save_path="$licenses_dir"
}

# Fails when the given directory holds no licenses or holds files that are not licenses
function verifyLicenses() {
  local licenses_dir="$1" non_license_files

  non_license_files=$(cd "$licenses_dir" && find . -type f ! \( -name 'LICENSE*' -o -name 'NOTICE*' -o -name 'COPYING*' \))
  if [[ -n "$non_license_files" ]]; then
    echo "go-licenses saved files that are not licenses:"
    echo "$non_license_files"
    exit 1
  fi

  if [[ -z "$(find "$licenses_dir" -type f)" ]]; then
    echo "go-licenses saved no licenses"
    exit 1
  fi
}

# Prints a banner with the given title
function printHeader() {
  local title="$1"

  printf '================================================================\n%s\n================================================================\n\n' "$title"
}

# Writes the Go runtime license and the licenses in the given directory to PACKAGE_DIR/THIRD_PARTY_LICENSES
function writeLicenses() {
  local licenses_dir="$1" license_path

  {
    printHeader "Go standard library and runtime"
    cat "$(go env GOROOT)/LICENSE"
    printf '\n'

    # Sorting in the C locale orders by byte value, so the licenses come in the same order on every machine
    while IFS= read -r license_path; do
      printHeader "${license_path#./}"
      cat "$licenses_dir/$license_path"
      printf '\n'
    done < <(cd "$licenses_dir" && find . -type f | LC_ALL=C sort)
  } > "$PACKAGE_DIR/THIRD_PARTY_LICENSES"
}

# Collects the licenses of the Go runtime and the compiled-in modules into one file
function main() {
  local licenses_dir

  setup

  licenses_dir="$(dirname "$PACKAGE_DIR")/licenses"

  saveLicenses "$licenses_dir"
  verifyLicenses "$licenses_dir"
  writeLicenses "$licenses_dir"
}

main
