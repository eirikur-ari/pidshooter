#!/usr/bin/env bash
# Manages the devcontainer of this workspace: start, rebuild, stop, shell, destroy or status.
# Requires WORKSPACE_FOLDER in the environment. DC_SHELL selects the shell of the shell command and defaults to zsh.
set -euo pipefail

if [[ -z "${WORKSPACE_FOLDER:-}" ]]; then
  echo "missing or empty environment variable: WORKSPACE_FOLDER" >&2
  exit 2
fi

LABEL_FILTER="label=devcontainer.local_folder=$WORKSPACE_FOLDER"
DC_SHELL="${DC_SHELL:-zsh}"

start() {
  devcontainer up --workspace-folder "$WORKSPACE_FOLDER"
}

rebuild() {
  devcontainer up --workspace-folder "$WORKSPACE_FOLDER" --remove-existing-container --build-no-cache
}

stop() {
  local id
  id=$(docker ps -q --filter "$LABEL_FILTER")
  if [[ -n "$id" ]]; then
    # shellcheck disable=SC2086
    docker stop $id
  else
    echo "No running devcontainer found."
  fi
}

shell() {
  devcontainer exec --workspace-folder "$WORKSPACE_FOLDER" "$DC_SHELL"
}

destroy() {
  local id img=""
  id=$(docker ps -aq --filter "$LABEL_FILTER")
  if [[ -n "$id" ]]; then
    # shellcheck disable=SC2086
    img=$(docker inspect --format '{{.Image}}' $id | sort -u)
    # shellcheck disable=SC2086
    docker rm -f $id
  else
    echo "No devcontainer found."
  fi
  if [[ -n "$img" ]]; then
    # shellcheck disable=SC2086
    docker rmi -f $img
  fi
}

status() {
  docker ps -a --filter "$LABEL_FILTER" \
    --format 'table {{.ID}}\t{{.Status}}\t{{.Image}}'
}

case "${1:-}" in
  start) start ;;
  rebuild) rebuild ;;
  stop) stop ;;
  shell) shell ;;
  destroy) destroy ;;
  status) status ;;
  *)
    echo "expected one argument, the devcontainer action to run: start, rebuild, stop, shell, destroy or status"
    exit 2
    ;;
esac
