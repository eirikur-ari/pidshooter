#!/usr/bin/env bash
# Gathers the binary, license files, readme, its images and changelog in PACKAGE_DIR and archives that into DIST_DIR.
# Requires WORKSPACE_FOLDER, BIN_DIR, BINARY_NAME, PACKAGE_DIR and DIST_DIR in the environment.
set -euo pipefail

missing=()
for name in WORKSPACE_FOLDER BIN_DIR BINARY_NAME PACKAGE_DIR DIST_DIR; do
  [[ -n "${!name:-}" ]] || missing+=("$name")
done
if (( ${#missing[@]} > 0 )); then
  echo "missing or empty environment variables: ${missing[*]}" >&2
  exit 2
fi

cd "$WORKSPACE_FOLDER"

version=$("$BIN_DIR/$BINARY_NAME" --version | awk '{print $3}')
archive="$WORKSPACE_FOLDER/$DIST_DIR/${BINARY_NAME}_${version}_$(go env GOOS)_$(go env GOARCH).tar.gz"

mkdir -p "$PACKAGE_DIR" "$DIST_DIR"

cp "$BIN_DIR/$BINARY_NAME" LICENSE README.md CHANGELOG.md "$PACKAGE_DIR/"
rm -rf "$PACKAGE_DIR/docs"
mkdir -p "$PACKAGE_DIR/docs"
cp -R docs/images "$PACKAGE_DIR/docs/images"

cd "$PACKAGE_DIR"
COPYFILE_DISABLE=1 tar --no-xattrs --exclude .DS_Store -czf "$archive" -- *
