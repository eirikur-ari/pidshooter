#!/usr/bin/env bash
# Manages the devcontainer of this workspace: start, rebuild, stop, shell, copy, destroy or status.
# Requires WORKSPACE_FOLDER in the environment. DC_SHELL selects the shell of the shell command and defaults to zsh.
# The copy command takes the path of a file or directory and an optional destination directory inside the running devcontainer.
# A file is copied into the destination, a directory has its contents copied into it.
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

copy() {
  local source="${1:-}"
  local destination="${2:-/home/sandbox/tmp}"
  local id

  if [[ -z "$source" ]]; then
    echo "expected the path of the file or directory to copy into the devcontainer" >&2
    exit 2
  fi
  if [[ ! -e "$source" ]]; then
    echo "the file or directory to copy into the devcontainer does not exist: $source" >&2
    exit 2
  fi

  id=$(docker ps -q --filter "$LABEL_FILTER")
  if [[ -z "$id" ]]; then
    echo "No running devcontainer found." >&2
    exit 1
  fi

  docker exec "$id" mkdir -p "$destination"
  if [[ -d "$source" ]]; then
    COPYFILE_DISABLE=1 tar -C "$source" -cf - . | docker exec -i "$id" tar -C "$destination" -xf -
  else
    COPYFILE_DISABLE=1 tar -C "$(dirname "$source")" -cf - "$(basename "$source")" | docker exec -i "$id" tar -C "$destination" -xf -
  fi
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
  copy) copy "${2:-}" "${3:-}" ;;
  destroy) destroy ;;
  status) status ;;
  *)
    echo "expected the devcontainer action to run as the first argument: start, rebuild, stop, shell, copy, destroy or status"
    exit 2
    ;;
esac
