#!/bin/sh
set -eu

# Run pre-built wx-fe:latest with configurable API base
# Usage:
#   ./doc/docker-file/start_wx_fe_docker.sh http://192.168.3.82:9292
# If omitted, defaults to http://127.0.0.1:9292

API_BASE="${1:-}"
if [ -z "$API_BASE" ]; then
  API_BASE="http://127.0.0.1:9292"
fi

IMAGE_NAME="wx-fe:latest"
CONTAINER_NAME="wx-fe-dev"
HOST_PORT="9093"
CONTAINER_PORT="80"

# Ensure image exists
if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
  echo "Image $IMAGE_NAME not found. Build it first with:"
  echo "  ./doc/docker-file/run_wx_fe_docker.sh <API_BASE>"
  exit 1
fi

# Restart container
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  echo "[wx-fe] Removing existing container: $CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
fi

echo "[wx-fe] Starting container: $CONTAINER_NAME -> http://localhost:${HOST_PORT}"
docker run -d --name "$CONTAINER_NAME" -p "${HOST_PORT}:${CONTAINER_PORT}" \
  --add-host=host.docker.internal:host-gateway \
  -e WX_API_BASE_URL="$API_BASE" \
  "$IMAGE_NAME"

echo "[wx-fe] Done. Open http://127.0.0.1:${HOST_PORT}"
