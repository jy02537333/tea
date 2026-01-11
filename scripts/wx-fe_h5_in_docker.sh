#!/usr/bin/env bash
set -euo pipefail

# wx-fe H5 in Docker: one-click dev or static preview
# Usage:
#   bash scripts/wx-fe_h5_in_docker.sh dev
#   bash scripts/wx-fe_h5_in_docker.sh preview
# Env vars:
#   PORT                 Port to expose (default: 10088)
#   WX_API_BASE_URL      API base url (default: http://host.docker.internal:9292)
#   AUTO_VERIFY           If set to 1 (preview only), run container detached, wait for preview ready, run Playwright verify, then cleanup
#   NAME                 Container name (default: tea-wxfe)
#   IMAGE                Base image (default: node:18-alpine)

MODE="${1:-}"
PORT="${PORT:-10088}"
# Note: For H5, API calls are issued by the host browser, so 127.0.0.1 is typically the correct default.
# The container still gets host.docker.internal mapped via --add-host, but the verifier runs on the host.
API_BASE="${WX_API_BASE_URL:-http://127.0.0.1:9292}"
AUTO_VERIFY="${AUTO_VERIFY:-0}"
OUTPUT_ROOT="dist-fixed"
NAME="${NAME:-tea-wxfe}"
IMAGE="${IMAGE:-node:18-alpine}"
WORKDIR="/app/wx-fe"

fail_port_in_use() {
  local port="$1"
  echo "[wx-fe] ERROR: port $port is still in use on the host." >&2
  echo "[wx-fe] Hint: try running this script with sudo, or stop the process manually." >&2
  echo "[wx-fe] Containers exposing $port:" >&2
  docker ps --format 'table {{.Names}}\t{{.Ports}}\t{{.Status}}' | (head -n 1; grep -F ":${port}->" || true) >&2
  if command -v ss >/dev/null 2>&1; then
    echo "[wx-fe] ss listeners for :$port:" >&2
    ss -ltnp 2>/dev/null | grep -E "(:|\[::\]:)${port}\b" >&2 || true
  fi
  exit 2
}

kill_port_users() {
  local port="$1"

  # 1) Prefer stopping docker containers that publish this port
  local ids=""
  ids="$(docker ps --filter "publish=${port}" -q 2>/dev/null || true)"
  if [[ -n "$ids" ]]; then
    echo "[wx-fe] Port $port is in use; stopping docker containers publishing it..." >&2
    # shellcheck disable=SC2086
    docker rm -f $ids >/dev/null 2>&1 || true
  fi

  # 2) If still in use, kill host processes (best-effort)
  local pids=""
  if command -v lsof >/dev/null 2>&1; then
    pids="$(lsof -t -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | tr '\n' ' ' || true)"
  elif command -v ss >/dev/null 2>&1; then
    # Parse pid=NNN from ss output (may require sudo to see other users' pids)
    pids="$(ss -ltnpH 2>/dev/null | awk -v p=":${port}" '$4 ~ (p"$") {print $0}' | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p' | sort -u | tr '\n' ' ' || true)"
  fi

  if [[ -n "$pids" ]]; then
    echo "[wx-fe] Port $port still in use; killing host listener PID(s): $pids" >&2
    # shellcheck disable=SC2086
    kill -TERM $pids >/dev/null 2>&1 || true
    sleep 0.5
    # shellcheck disable=SC2086
    kill -KILL $pids >/dev/null 2>&1 || true
  fi
}

check_port_available() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    if ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "(:|\[::\]:)${port}$"; then
      kill_port_users "$port"
      if ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "(:|\[::\]:)${port}$"; then
        fail_port_in_use "$port"
      fi
    fi
  elif command -v lsof >/dev/null 2>&1; then
    if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
      kill_port_users "$port"
      if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
        fail_port_in_use "$port"
      fi
    fi
  else
    # Fallback: detect via docker ps only (won't catch non-docker listeners)
    if docker ps --format '{{.Ports}}' | grep -Fq ":${port}->"; then
      kill_port_users "$port"
      if docker ps --format '{{.Ports}}' | grep -Fq ":${port}->"; then
        fail_port_in_use "$port"
      fi
    fi
  fi
}

usage() {
  echo "Usage: $0 [dev|preview]" >&2
  echo "Env: PORT (default 10088), WX_API_BASE_URL (default http://host.docker.internal:9292), NAME, IMAGE" >&2
  exit 1
}

[[ -z "$MODE" ]] && usage
if [[ "$MODE" != "dev" && "$MODE" != "preview" ]]; then
  usage
fi

# Stop previous container if exists
docker rm -f "$NAME" >/dev/null 2>&1 || true

check_port_available "$PORT"

if [[ "$MODE" == "dev" ]]; then
  # Dev server with HMR on PORT
  CMD="apk add --no-cache git && npm i -g pnpm && pnpm install && pnpm run dev:h5"
else
  # Static preview: build then serve SPA on PORT
  CMD="apk add --no-cache git && npm i -g pnpm && pnpm install && pnpm run build:h5 && npx --yes serve -s $OUTPUT_ROOT -l $PORT"
fi

echo "[wx-fe] starting in Docker mode=$MODE port=$PORT api=$API_BASE (container: $NAME)" >&2

wait_preview_ready() {
  local url="$1"
  local timeout_s="${2:-120}"
  local start_ts
  start_ts="$(date +%s)"
  echo "[wx-fe] waiting for preview to be ready: $url (timeout ${timeout_s}s)" >&2

  if command -v curl >/dev/null 2>&1; then
    while true; do
      if curl -fsS --max-time 2 "$url" >/dev/null 2>&1; then
        return 0
      fi
      if (( $(date +%s) - start_ts > timeout_s )); then
        return 1
      fi
      sleep 1
    done
  elif command -v wget >/dev/null 2>&1; then
    while true; do
      if wget -q -T 2 -O /dev/null "$url" >/dev/null 2>&1; then
        return 0
      fi
      if (( $(date +%s) - start_ts > timeout_s )); then
        return 1
      fi
      sleep 1
    done
  else
    echo "[wx-fe] WARN: neither curl nor wget found; skipping readiness check" >&2
    return 0
  fi
}

run_verify() {
  local preview_url="http://127.0.0.1:${PORT}/"
  if ! wait_preview_ready "$preview_url" 180; then
    echo "[wx-fe] ERROR: preview not ready on $preview_url" >&2
    return 3
  fi
  echo "[wx-fe] running Playwright verification against $preview_url" >&2

  # Playwright verifier runs on the host. On Linux, host.docker.internal may not resolve on the host,
  # so rewrite it to 127.0.0.1 for the verifier's API probes.
  local verify_api_base="$API_BASE"
  if [[ "$verify_api_base" == *"host.docker.internal"* ]]; then
    verify_api_base="${verify_api_base//host.docker.internal/127.0.0.1}"
  fi

  PREVIEW_URL="$preview_url" WX_API_BASE_URL="$verify_api_base" node tools/automation/wx-fe-verify.playwright.js
}

if [[ "$MODE" == "preview" && "$AUTO_VERIFY" == "1" ]]; then
  # Run detached, verify from host, then cleanup.
  docker run --rm -d \
    --name "$NAME" \
    -p "$PORT":"$PORT" \
    --add-host=host.docker.internal:host-gateway \
    -e WX_API_BASE_URL="$API_BASE" \
    -e CHOKIDAR_USEPOLLING=1 \
    -v "$PWD/wx-fe":"$WORKDIR" \
    -w "$WORKDIR" \
    "$IMAGE" sh -lc "$CMD" >/dev/null

  cleanup() {
    docker rm -f "$NAME" >/dev/null 2>&1 || true
  }
  trap cleanup EXIT

  run_verify
else
  exec docker run --rm -it \
    --name "$NAME" \
    -p "$PORT":"$PORT" \
    --add-host=host.docker.internal:host-gateway \
    -e WX_API_BASE_URL="$API_BASE" \
    -e CHOKIDAR_USEPOLLING=1 \
    -v "$PWD/wx-fe":"$WORKDIR" \
    -w "$WORKDIR" \
    "$IMAGE" sh -lc "$CMD"
fi
