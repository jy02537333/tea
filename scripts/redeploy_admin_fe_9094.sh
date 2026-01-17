#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONTAINER_NAME="tea-admin-fe"
IMAGE_NAME="admin-fe:latest"
PORT="9094"

# Try to reuse current API base from existing container (if any)
API_BASE=""
if docker inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
  API_BASE=$(docker inspect "$CONTAINER_NAME" --format '{{range .Config.Env}}{{println .}}{{end}}' \
    | awk -F= '$1=="VITE_API_BASE"{print substr($0, index($0,"=")+1)}' \
    | tail -n 1 || true)
fi

if [[ -z "$API_BASE" ]]; then
  echo "[redeploy_admin_fe_9094] VITE_API_BASE not found from running container."
  echo "[redeploy_admin_fe_9094] Provide it via env: VITE_API_BASE=http://host:port bash $0"
  if [[ -n "${VITE_API_BASE:-}" ]]; then
    API_BASE="$VITE_API_BASE"
  else
    exit 1
  fi
fi

echo "[redeploy_admin_fe_9094] Building $IMAGE_NAME ..."
docker build -t "$IMAGE_NAME" -f admin-fe/Dockerfile .

echo "[redeploy_admin_fe_9094] Restarting container $CONTAINER_NAME on :$PORT ..."
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

docker run -d \
  --name "$CONTAINER_NAME" \
  --restart unless-stopped \
  -p "$PORT":80 \
  -e "VITE_API_BASE=$API_BASE" \
  "$IMAGE_NAME" >/dev/null

echo "[redeploy_admin_fe_9094] Done. URL: http://127.0.0.1:$PORT/"
echo "[redeploy_admin_fe_9094] Using VITE_API_BASE=$API_BASE"
