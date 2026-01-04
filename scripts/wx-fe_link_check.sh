#!/usr/bin/env bash
set -euo pipefail

# wx-fe 首页与分类联调验收（小流量）
# - 启动 tea-api（9292）并获取 token
# - 启动 wx-fe H5 预览（默认 10088）并检查就绪
# - 以 token 调用 stores/products 接口，生成摘要

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

LOG_DIR="$REPO_ROOT/build-ci-logs"
OUT_DIR="/tmp/wx-fe-link-check.$(date +%s)"
mkdir -p "$OUT_DIR"

WX_API_BASE_URL_DEFAULT="http://127.0.0.1:9292"
WX_API_BASE_URL="${WX_API_BASE_URL:-$WX_API_BASE_URL_DEFAULT}"
WX_H5_PORT="${WX_H5_PORT:-10088}"
WX_H5_URL="http://127.0.0.1:${WX_H5_PORT}"

echo "[wx-fe] Using WX_API_BASE_URL=$WX_API_BASE_URL; H5 dev port=$WX_H5_PORT"

# 1) 启动后端（9292）并获取 token
echo "[wx-fe] Starting tea-api via run-tea-api.sh..."
"$REPO_ROOT/run-tea-api.sh" || {
  echo "[wx-fe] ERROR: tea-api failed to start"; exit 2;
}

ADMIN_LOGIN_JSON="$LOG_DIR/admin_login_response.json"
TOKEN=""
if [ -s "$ADMIN_LOGIN_JSON" ]; then
  TOKEN=$(python3 - "$ADMIN_LOGIN_JSON" <<'PY'
import sys, json
try:
    obj = json.load(open(sys.argv[1]))
    data = obj.get('data') or obj
    print((data.get('token') if isinstance(data, dict) else '') or '')
except Exception:
    print('')
PY
  )
fi

if [ -z "$TOKEN" ]; then
  echo "[wx-fe] WARN: token not found in $ADMIN_LOGIN_JSON, retry login"
  LOGIN_URL="${WX_API_BASE_URL}/api/v1/auth/login"
  curl -sS -H "Content-Type: application/json" -d '{"username":"admin","password":"pass"}' "$LOGIN_URL" -o "$ADMIN_LOGIN_JSON" || true
  TOKEN=$(python3 - "$ADMIN_LOGIN_JSON" <<'PY'
import sys, json
try:
    obj = json.load(open(sys.argv[1]))
    data = obj.get('data') or obj
    print((data.get('token') if isinstance(data, dict) else '') or '')
except Exception:
    print('')
PY
  )
fi

if [ -z "$TOKEN" ]; then
  echo "[wx-fe] ERROR: failed to obtain token"; exit 3;
fi
echo "[wx-fe] Obtained token (truncated): ${TOKEN:0:16}..."

# 2) 启动 wx-fe H5 预览（后台运行）
echo "[wx-fe] Starting H5 dev preview (port ${WX_H5_PORT})..."
H5_LOG="$OUT_DIR/h5-preview.out"
nohup env WX_API_BASE_URL="$WX_API_BASE_URL" pnpm -C "$REPO_ROOT/wx-fe" run dev:h5 > "$H5_LOG" 2>&1 &
H5_PID=$!
echo "$H5_PID" > "$OUT_DIR/h5-preview.pid"
echo "[wx-fe] H5 dev PID: $H5_PID (logs: $H5_LOG)"

# 3) 等待 H5 就绪
MAX_WAIT=60
WAIT_IDX=0
READY=0
while [ $WAIT_IDX -lt $MAX_WAIT ]; do
  if curl -sSf "$WX_H5_URL" >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 1
  WAIT_IDX=$((WAIT_IDX+1))
done

if [ "$READY" -ne 1 ]; then
  echo "[wx-fe] ERROR: H5 dev not ready after ${MAX_WAIT}s (see $H5_LOG)"; exit 4;
fi
echo "[wx-fe] H5 dev is ready at $WX_H5_URL"

# 4) API 小流量验收：stores/products
SUMMARY_TXT="$OUT_DIR/summary.txt"
STORES_JSON="$OUT_DIR/stores.json"
PRODUCTS_JSON="$OUT_DIR/products.json"

echo "[wx-fe] Fetching stores..."
curl -sS -H "Authorization: Bearer $TOKEN" "${WX_API_BASE_URL}/api/v1/stores?page=1&limit=10" -o "$STORES_JSON" || true

echo "[wx-fe] Fetching products (first page)..."
curl -sS -H "Authorization: Bearer $TOKEN" "${WX_API_BASE_URL}/api/v1/products?page=1&limit=10" -o "$PRODUCTS_JSON" || true

STORES_COUNT=$(python3 - "$STORES_JSON" <<'PY'
import json,sys
def count_items(obj):
    if isinstance(obj, dict):
        if 'data' in obj and isinstance(obj['data'], list):
            return len(obj['data'])
        if 'items' in obj and isinstance(obj['items'], list):
            return len(obj['items'])
    if isinstance(obj, list):
        return len(obj)
    return 0
try:
    stores = json.load(open(sys.argv[1]))
    print(count_items(stores))
except Exception:
    print(0)
PY
  )

PRODUCTS_COUNT=$(python3 - "$PRODUCTS_JSON" <<'PY'
import json,sys
def count_items(obj):
    if isinstance(obj, dict):
        if 'data' in obj and isinstance(obj['data'], list):
            return len(obj['data'])
        if 'items' in obj and isinstance(obj['items'], list):
            return len(obj['items'])
    if isinstance(obj, list):
        return len(obj)
    return 0
try:
    products = json.load(open(sys.argv[1]))
    print(count_items(products))
except Exception:
    print(0)
PY
  )

{
  echo "wx-fe 联调验收（首页/分类）"
  echo "H5 dev: $WX_H5_URL (PID=$H5_PID)"
  echo "API base: $WX_API_BASE_URL"
  echo "Stores count: $STORES_COUNT"
  echo "Products count: $PRODUCTS_COUNT"
} > "$SUMMARY_TXT"

echo "[wx-fe] Done. Summary -> $SUMMARY_TXT"
echo "[wx-fe] You can stop H5 dev with: kill \$(cat $OUT_DIR/h5-preview.pid)"

exit 0
