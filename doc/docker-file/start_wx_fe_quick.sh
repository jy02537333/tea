#!/bin/sh
set -eu

# Quick start wx-fe from prebuilt image (no rebuild)
# Usage:
#   ./doc/docker-file/start_wx_fe_quick.sh <API_BASE_URL> [HOST_PORT]
# Examples:
#   ./doc/docker-file/start_wx_fe_quick.sh http://host.docker.internal:9292
#   ./doc/docker-file/start_wx_fe_quick.sh http://192.168.3.82:9292 9095

API_BASE="${1:-}"
HOST_PORT="${2:-9093}"

if [ -z "$API_BASE" ]; then
  echo "API_BASE_URL is required. Example: ./doc/docker-file/start_wx_fe_quick.sh http://192.168.3.82:9292"
  exit 1
fi

IMAGE_NAME="wx-fe:latest"
CONTAINER_NAME="wx-fe-dev"
CONTAINER_PORT="80"

# Ensure image exists
if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
  echo "Image $IMAGE_NAME not found. Please build first:"
  echo "  ./doc/docker-file/run_wx_fe_docker.sh <API_BASE_URL>"
  exit 1
fi

# Restart container if exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  echo "[wx-fe] Removing existing container: $CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
fi

echo "[wx-fe] Starting container: $CONTAINER_NAME -> http://localhost:${HOST_PORT}"
docker run -d --name "$CONTAINER_NAME" \
  -p "${HOST_PORT}:${CONTAINER_PORT}" \
  --add-host=host.docker.internal:host-gateway \
  -e WX_API_BASE_URL="$API_BASE" \
  "$IMAGE_NAME"

echo "[wx-fe] Started. Open http://127.0.0.1:${HOST_PORT}"
