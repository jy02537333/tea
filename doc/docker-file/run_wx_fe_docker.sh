#!/bin/sh
set -eu

# One-click build & run for wx-fe (H5)
# Usage:
#   ./doc/docker-file/run_wx_fe_docker.sh http://192.168.3.82:9292
# If omitted, defaults to http://127.0.0.1:9292

API_BASE="${1:-}"
if [ -z "$API_BASE" ]; then
  API_BASE="http://127.0.0.1:9292"
fi

IMAGE_NAME="wx-fe:latest"
CONTAINER_NAME="wx-fe-dev"
CONTEXT_DIR="wx-fe"
DOCKERFILE_PATH="wx-fe/Dockerfile"
HOST_PORT="9093"
CONTAINER_PORT="80"

echo "[wx-fe] Building H5 assets (pnpm build:h5)"
if command -v pnpm >/dev/null 2>&1; then
  (cd "$CONTEXT_DIR" && pnpm -s build:h5 || pnpm -s build)
else
  echo "[wx-fe] pnpm not found; attempting npm/yarn build"
  (cd "$CONTEXT_DIR" && (npm run build:h5 || npm run build))
fi

echo "[wx-fe] Building image: $IMAGE_NAME (context: $CONTEXT_DIR)"
docker build -t "$IMAGE_NAME" -f "$DOCKERFILE_PATH" "$CONTEXT_DIR" --build-arg WX_API_BASE_URL=__WX_API_BASE_URL_PLACEHOLDER__

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
