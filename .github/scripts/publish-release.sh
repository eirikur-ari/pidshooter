#!/usr/bin/env bash
# Publishes the archives and checksums in build/dist/ and the notes in
# build/release-notes.md as the GitHub release of RELEASE_VERSION, creating its annotated
# tag last.
# Requires RELEASE_VERSION, GITHUB_SHA, GITHUB_REPOSITORY and GH_TOKEN in the environment.
set -euo pipefail

missing=()
for name in RELEASE_VERSION GITHUB_SHA GITHUB_REPOSITORY GH_TOKEN; do
  [[ -n "${!name:-}" ]] || missing+=("$name")
done
if (( ${#missing[@]} > 0 )); then
  echo "missing or empty environment variables: ${missing[*]}" >&2
  exit 2
fi

if out=$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/v$RELEASE_VERSION" --jq '"\(.object.type) \(.object.sha)"' 2>&1); then
  read -r type sha <<< "$out"
  if [[ "$type" == "tag" ]]; then
    sha=$(gh api "repos/$GITHUB_REPOSITORY/git/tags/$sha" --jq .object.sha)
  fi
  if [[ "$sha" != "$GITHUB_SHA" ]]; then
    echo "tag v$RELEASE_VERSION already exists and points to a different commit than the one built"
    exit 1
  fi
else
  if [[ "$out" != *"HTTP 404"* ]]; then
    echo "$out" >&2
    exit 1
  fi
  tag=$(gh api -X POST "repos/$GITHUB_REPOSITORY/git/tags" \
    -f tag="v$RELEASE_VERSION" -f message="v$RELEASE_VERSION" -f object="$GITHUB_SHA" -f type=commit --jq .sha)
  gh api -X POST "repos/$GITHUB_REPOSITORY/git/refs" -f ref="refs/tags/v$RELEASE_VERSION" -f sha="$tag"
fi

releases=$(gh api "repos/$GITHUB_REPOSITORY/releases" --paginate \
  --jq ".[] | select(.tag_name == \"v$RELEASE_VERSION\") | \"\(.id) \(.draft)\"")
while read -r id draft; do
  [[ -n "$id" ]] || continue
  if [[ "$draft" != "true" ]]; then
    echo "release v$RELEASE_VERSION is already published"
    exit 1
  fi
  gh api -X DELETE "repos/$GITHUB_REPOSITORY/releases/$id"
done <<< "$releases"

gh release create "v$RELEASE_VERSION" --draft --verify-tag \
  --title "v$RELEASE_VERSION" --notes-file build/release-notes.md \
  build/dist/*.tar.gz build/dist/checksums.txt
gh release edit "v$RELEASE_VERSION" --draft=false
