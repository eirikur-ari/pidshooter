#!/usr/bin/env bash
# Publishes the archives and checksums in build/dist/ and the notes in
# build/release-notes.md as the GitHub release of RELEASE_VERSION, creating its annotated
# tag when it does not exist.
# Requires RELEASE_VERSION, GITHUB_SHA, GITHUB_REPOSITORY and GH_TOKEN in the environment.
set -euo pipefail

# Fails when a required environment variable is missing or empty
function setup() {
  local missing=() name
  for name in RELEASE_VERSION GITHUB_SHA GITHUB_REPOSITORY GH_TOKEN; do
    [[ -n "${!name:-}" ]] || missing+=("$name")
  done
  if (( ${#missing[@]} > 0 )); then
    echo "missing or empty environment variables: ${missing[*]}" >&2
    exit 2
  fi
}

# Fails when the existing release tag does not point to the commit that was built
function verifyTagCommit() {
  local type="$1" sha="$2"

  if [[ "$type" == "tag" ]]; then
    sha=$(gh api "repos/$GITHUB_REPOSITORY/git/tags/$sha" --jq .object.sha)
  fi
  if [[ "$sha" != "$GITHUB_SHA" ]]; then
    echo "tag v$RELEASE_VERSION already exists and points to a different commit than the one built"
    exit 1
  fi
}

# Creates the annotated release tag on the commit that was built
function createTag() {
  local tag

  tag=$(gh api -X POST "repos/$GITHUB_REPOSITORY/git/tags" \
    -f tag="v$RELEASE_VERSION" -f message="v$RELEASE_VERSION" -f object="$GITHUB_SHA" -f type=commit --jq .sha)
  gh api -X POST "repos/$GITHUB_REPOSITORY/git/refs" -f ref="refs/tags/v$RELEASE_VERSION" -f sha="$tag"
}

# Creates the release tag, or verifies the existing one points to the commit that was built
function ensureTag() {
  local tag_ref type sha

  if tag_ref=$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/v$RELEASE_VERSION" --jq '"\(.object.type) \(.object.sha)"' 2>&1); then
    read -r type sha <<< "$tag_ref"
    verifyTagCommit "$type" "$sha"
  else
    if [[ "$tag_ref" != *"HTTP 404"* ]]; then
      echo "$tag_ref" >&2
      exit 1
    fi
    createTag
  fi
}

# Deletes the leftover draft releases of the release version, failing when one is already published
function deleteDraftReleases() {
  local releases release_id is_draft

  releases=$(gh api "repos/$GITHUB_REPOSITORY/releases" --paginate \
    --jq ".[] | select(.tag_name == \"v$RELEASE_VERSION\") | \"\(.id) \(.draft)\"")
  while read -r release_id is_draft; do
    [[ -n "$release_id" ]] || continue
    if [[ "$is_draft" != "true" ]]; then
      echo "release v$RELEASE_VERSION is already published"
      exit 1
    fi
    gh api -X DELETE "repos/$GITHUB_REPOSITORY/releases/$release_id"
  done <<< "$releases"
}

# Creates the release as a draft with its assets and notes, then publishes it
function publishRelease() {
  gh release create "v$RELEASE_VERSION" --draft --verify-tag \
    --title "v$RELEASE_VERSION" --notes-file build/release-notes.md \
    build/dist/*.tar.gz build/dist/checksums.txt
  gh release edit "v$RELEASE_VERSION" --draft=false
}

# Publishes the release of the release version, creating its tag
function main() {
  setup
  ensureTag
  deleteDraftReleases
  publishRelease
}

main
