#!/usr/bin/env bash
set -euo pipefail

# Stop wx-fe Docker container(s)
# Usage:
#   bash scripts/wx-fe_stop_docker.sh            # stop default name
#   bash scripts/wx-fe_stop_docker.sh <name>     # stop specific name
# Env:
#   NAME  Container name to stop (default: tea-wxfe)
#   ALL   If set to 1, stop all containers whose names contain NAME pattern

NAME_ARG="${1:-}"
NAME_PATTERN="${NAME:-tea-wxfe}"
ALL_MODE="${ALL:-0}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found" >&2
  exit 127
fi

stop_one() {
  local n="$1"
  docker rm -f "$n" >/dev/null 2>&1 && echo "stopped: $n" || true
}

if [[ -n "$NAME_ARG" ]]; then
  stop_one "$NAME_ARG"
  exit 0
fi

if [[ "$ALL_MODE" == "1" ]]; then
  mapfile -t names < <(docker ps -a --format '{{.Names}}' | grep -E "$NAME_PATTERN" || true)
  if [[ ${#names[@]} -eq 0 ]]; then
    echo "no containers matched pattern: $NAME_PATTERN" >&2
    exit 0
  fi
  for n in "${names[@]}"; do
    stop_one "$n"
  done
else
  # default single name
  stop_one "$NAME_PATTERN"
fi
