#!/usr/bin/env bash
set -eu

# This script builds and runs the admin-fe Docker container.
# It navigates to the project root, builds the image,
# stops any existing container with the same name,
# and starts a new one with the specified API base URL.

# Usage:
# ./run_admin_fe_docker.sh [API_BASE_URL]
#
# Example:
# ./run_admin_fe_docker.sh http://10.0.0.5:9092

# Get the absolute path of the project root directory
# This script is in doc/docker-file, so we go up two levels
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

# --- Configuration ---
IMAGE_NAME="admin-fe:latest"
CONTAINER_NAME="tea-admin-fe"
HOST_PORT="9094"

# Use the first script argument as the API base URL,
# otherwise, fall back to the default value.
DEFAULT_API_BASE_URL="http://192.168.3.82:9092"
API_BASE_URL=${1:-$DEFAULT_API_BASE_URL}

# --- Build Step ---
echo "==> Building Docker image '$IMAGE_NAME' from Dockerfile in admin-fe/..."
docker build -t "$IMAGE_NAME" -f admin-fe/Dockerfile .
echo "==> Build complete."

# --- Run Step ---
echo "==> Checking for and removing existing container named '$CONTAINER_NAME'..."
docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
echo "==> Old container removed."

echo "==> Starting new container '$CONTAINER_NAME'..."
docker run --rm -d -p "${HOST_PORT}:80" \
  -e VITE_API_BASE="$API_BASE_URL" \
  --name "$CONTAINER_NAME" \
  "$IMAGE_NAME"

# --- Output ---
echo ""
echo "================================================="
echo " Admin FE container started successfully!"
echo "-------------------------------------------------"
echo "  - Container Name: $CONTAINER_NAME"
echo "  - Listening on:   http://localhost:${HOST_PORT}"
echo "  - API Proxy Target: $API_BASE_URL"
echo "-------------------------------------------------"
echo " To view logs, run: docker logs -f $CONTAINER_NAME"
echo " To stop, run:      docker stop $CONTAINER_NAME"
echo "================================================="
