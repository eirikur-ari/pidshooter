#!/usr/bin/env bash
# Gathers the binary, license files, readme, its images and changelog in PACKAGE_DIR and archives that into DIST_DIR.
# Requires WORKSPACE_FOLDER, BIN_DIR, BINARY_NAME, PACKAGE_DIR and DIST_DIR in the environment.
set -euo pipefail

# Fails when a required environment variable is missing or empty, then moves to the workspace and creates the output directories
function setup() {
  local missing=() name
  for name in WORKSPACE_FOLDER BIN_DIR BINARY_NAME PACKAGE_DIR DIST_DIR; do
    [[ -n "${!name:-}" ]] || missing+=("$name")
  done
  if (( ${#missing[@]} > 0 )); then
    echo "missing or empty environment variables: ${missing[*]}" >&2
    exit 2
  fi

  cd "$WORKSPACE_FOLDER"
  mkdir -p "$PACKAGE_DIR/docs" "$DIST_DIR"
}

# Prints the version reported by the built binary
function releaseVersion() {
  "$BIN_DIR/$BINARY_NAME" --version | awk '{print $3}'
}

# Prints the path of the archive for the given version on the current platform
function archivePath() {
  local version="$1" os arch
  os=$(go env GOOS) || return
  arch=$(go env GOARCH) || return
  echo "$WORKSPACE_FOLDER/$DIST_DIR/${BINARY_NAME}_${version}_${os}_${arch}.tar.gz"
}

# Gathers the files to ship in PACKAGE_DIR
function stagePackage() {
  cp "$BIN_DIR/$BINARY_NAME" LICENSE README.md CHANGELOG.md "$PACKAGE_DIR/"
  cp -R docs/images "$PACKAGE_DIR/docs/"
}

# Archives the content of PACKAGE_DIR into the given archive path
function createArchive() {
  local archive="$1"

  cd "$PACKAGE_DIR"
  # The tar of macOS adds its extended attributes to the archive as extra ._ files unless COPYFILE_DISABLE is set
  if [[ "$(uname)" == "Darwin" ]]; then
    COPYFILE_DISABLE=1 tar --no-xattrs --exclude .DS_Store -czf "$archive" -- *
  else
    tar -czf "$archive" -- *
  fi
}

# Packages the binary and its accompanying files into a platform archive
function main() {
  local version archive

  setup

  version=$(releaseVersion)
  archive=$(archivePath "$version")

  stagePackage
  createArchive "$archive"
}

main
