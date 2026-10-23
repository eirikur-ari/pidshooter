#!/usr/bin/env bash
# Publishes the archives and checksums in build/dist/ and the notes in
# build/release-notes.md as the GitHub release of RELEASE_VERSION, creating its annotated
# tag last.
# Requires RELEASE_VERSION, GITHUB_SHA, GITHUB_REPOSITORY and GH_TOKEN in the environment.
set -euo pipefail

if ! gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/v$RELEASE_VERSION" > /dev/null 2>&1; then
  tag=$(gh api -X POST "repos/$GITHUB_REPOSITORY/git/tags" \
    -f tag="v$RELEASE_VERSION" -f message="v$RELEASE_VERSION" -f object="$GITHUB_SHA" -f type=commit --jq .sha)
  gh api -X POST "repos/$GITHUB_REPOSITORY/git/refs" -f ref="refs/tags/v$RELEASE_VERSION" -f sha="$tag"
fi

release=$(gh api "repos/$GITHUB_REPOSITORY/releases" --paginate \
  --jq ".[] | select(.tag_name == \"v$RELEASE_VERSION\") | \"\(.id) \(.draft)\"")
if [[ -n "$release" ]]; then
  read -r id draft <<< "$release"
  if [[ "$draft" != "true" ]]; then
    echo "release v$RELEASE_VERSION is already published"
    exit 1
  fi
  gh api -X DELETE "repos/$GITHUB_REPOSITORY/releases/$id"
fi

gh release create "v$RELEASE_VERSION" --draft --verify-tag \
  --title "v$RELEASE_VERSION" --notes-file build/release-notes.md \
  build/dist/*.tar.gz build/dist/checksums.txt
gh release edit "v$RELEASE_VERSION" --draft=false
