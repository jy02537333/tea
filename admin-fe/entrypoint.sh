#!/bin/sh
set -eu

PLACEHOLDER="__VITE_API_BASE_PLACEHOLDER__"
TARGET_DIR="/usr/share/nginx/html"
JS_FILES_PATH="$TARGET_DIR/assets/*.js"
HTML_FILE="$TARGET_DIR/index.html"

if [ -z "${VITE_API_BASE:-}" ]; then
  echo "Error: VITE_API_BASE environment variable is not set." >&2
  echo "Set it via: docker run -e VITE_API_BASE=http://host:port ..." >&2
  exit 1
fi

echo "[entrypoint] Injecting API base URL into built assets..."
# Replace in JS bundle files
for file in $JS_FILES_PATH; do
  if [ -f "$file" ]; then
    sed -i "s|$PLACEHOLDER|$VITE_API_BASE|g" "$file"
  fi
done
# Optionally replace in index.html if placeholder appears
if [ -f "$HTML_FILE" ]; then
  sed -i "s|$PLACEHOLDER|$VITE_API_BASE|g" "$HTML_FILE" || true
fi

echo "[entrypoint] Injection done. Starting Nginx..."
exec "$@"
