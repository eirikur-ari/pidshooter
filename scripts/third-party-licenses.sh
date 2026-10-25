#!/usr/bin/env bash
# Writes the licenses of the Go runtime and of the third-party modules compiled into the binary to PACKAGE_DIR/THIRD_PARTY_LICENSES.
# Requires WORKSPACE_FOLDER and PACKAGE_DIR in the environment.
set -euo pipefail

missing=()
for name in WORKSPACE_FOLDER PACKAGE_DIR; do
  [[ -n "${!name:-}" ]] || missing+=("$name")
done
if (( ${#missing[@]} > 0 )); then
  echo "missing or empty environment variables: ${missing[*]}" >&2
  exit 2
fi

cd "$WORKSPACE_FOLDER"

LICENSES_DIR="$(dirname "$PACKAGE_DIR")/licenses"

mkdir -p "$PACKAGE_DIR"
rm -rf "$LICENSES_DIR"

go run github.com/google/go-licenses/v2@v2.0.1 save ./cmd/pidshooter --ignore github.com/eirikur-ari/pidshooter --save_path="$LICENSES_DIR"

unexpected=$(cd "$LICENSES_DIR" && find . -type f ! \( -name 'LICENSE*' -o -name 'NOTICE*' -o -name 'COPYING*' \))
if [[ -n "$unexpected" ]]; then
  echo "go-licenses saved files that are not licenses:"
  echo "$unexpected"
  exit 1
fi

if [[ -z "$(find "$LICENSES_DIR" -type f)" ]]; then
  echo "go-licenses saved no licenses"
  exit 1
fi

header() {
  printf '================================================================\n%s\n================================================================\n\n' "$1"
}

{
  header "Go standard library and runtime"
  cat "$(go env GOROOT)/LICENSE"
  printf '\n'

  while IFS= read -r file; do
    header "${file#./}"
    cat "$LICENSES_DIR/$file"
    printf '\n'
  done < <(cd "$LICENSES_DIR" && find . -type f | LC_ALL=C sort)
} > "$PACKAGE_DIR/THIRD_PARTY_LICENSES"
