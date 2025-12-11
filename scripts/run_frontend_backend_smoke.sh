#!/usr/bin/env bash
set -euo pipefail

API_BASE="${WX_API_BASE_URL:-${VITE_API_BASE_URL:-http://127.0.0.1:9292}}"
FE_PORT="${WX_FE_PORT:-9093}"

echo "[info] Using API base: ${API_BASE}"

check() {
  local url="$1"; shift
  echo "[check] GET ${url}"
  curl -fsSL "${url}" | head -c 300 || {
    echo "[error] request failed: ${url}" >&2
    exit 1
  }
}

post_json() {
  local url="$1"; local data="$2";
  echo "[check] POST ${url}"
  curl -fsSL -H 'Content-Type: application/json' -d "${data}" "${url}" | head -c 300 || {
    echo "[error] request failed: ${url}" >&2
    exit 1
  }
}

# backend health checks
check "${API_BASE}/health" || true
check "${API_BASE}/api/v1/categories" || true

# optional login smoke
post_json "${API_BASE}/api/v1/auth/login" '{"username":"admin","password":"admin123"}' || true

# front-end dev server hint
cat <<EOF
[next]
To run wx-fe against ${API_BASE}:

export WX_API_BASE_URL="${API_BASE}"
cd /home/frederic/project/tea/wx-fe
pnpm install
pnpm dev --port ${FE_PORT}

Then open http://127.0.0.1:${FE_PORT} and verify XHR to ${API_BASE}.
EOF
