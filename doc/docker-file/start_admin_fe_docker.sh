#!/usr/bin/env bash
set -eu

# Start admin-fe from an already built Docker image (admin-fe:latest)
# Usage:
#   ./start_admin_fe_docker.sh [API_BASE_URL]
# Example:
#   ./start_admin_fe_docker.sh http://192.168.3.82:9092
# If omitted, defaults to http://127.0.0.1:9292

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

IMAGE_NAME="admin-fe:latest"
CONTAINER_NAME="tea-admin-fe"
HOST_PORT="9094"
DEFAULT_API_BASE_URL="http://127.0.0.1:9292"
API_BASE_URL=${1:-$DEFAULT_API_BASE_URL}

# Ensure image exists
if ! docker image inspect "$IMAGE_NAME" >/dev/null 2>&1; then
  echo "Error: Docker image '$IMAGE_NAME' not found." >&2
  echo "Please build it first or pull it, then re-run this script." >&2
  exit 1
fi

# Stop/remove any existing container
docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true

# Run container
docker run --rm -d -p "${HOST_PORT}:80" \
  -e VITE_API_BASE="$API_BASE_URL" \
  --name "$CONTAINER_NAME" \
  "$IMAGE_NAME"

# Print info
echo ""
echo "================================================="
echo " Admin FE container started (existing image)."
echo "-------------------------------------------------"
echo "  - Image:          $IMAGE_NAME"
echo "  - Container Name: $CONTAINER_NAME"
echo "  - Listening on:   http://localhost:${HOST_PORT}"
echo "  - API Proxy:      $API_BASE_URL"
echo "-------------------------------------------------"
echo " To view logs: docker logs -f $CONTAINER_NAME"
echo " To stop:      docker stop $CONTAINER_NAME"
echo "================================================="
