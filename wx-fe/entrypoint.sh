#!/bin/sh
set -eu

# Runtime API base URL
API_BASE="${WX_API_BASE_URL:-}"
if [ -z "$API_BASE" ]; then
  API_BASE="http://127.0.0.1:9292"
fi

WEB_ROOT="/usr/share/nginx/html"
PLACEHOLDER="__WX_API_BASE_URL_PLACEHOLDER__"
DEFAULT_HOST_DOCKER_INTERNAL="http://host.docker.internal:9292"
DEFAULT_LOCALHOST="http://127.0.0.1:9292"

# Replace placeholder across built assets (JS/HTML)
# Use POSIX sed for Alpine
echo "[wx-fe] Using API base: $API_BASE"
# Replace placeholder if present
find "$WEB_ROOT" -type f \( -name "*.js" -o -name "*.html" \) -print0 | xargs -0 sed -i "s#${PLACEHOLDER}#${API_BASE}#g" || true
# Also replace common defaults to allow overriding builds without placeholder
find "$WEB_ROOT" -type f \( -name "*.js" -o -name "*.html" \) -print0 | xargs -0 sed -i "s#${DEFAULT_HOST_DOCKER_INTERNAL}#${API_BASE}#g" || true
find "$WEB_ROOT" -type f \( -name "*.js" -o -name "*.html" \) -print0 | xargs -0 sed -i "s#${DEFAULT_LOCALHOST}#${API_BASE}#g" || true

exec "$@"
