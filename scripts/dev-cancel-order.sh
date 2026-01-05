#!/usr/bin/env bash
set -euo pipefail

# dev-cancel-order.sh
# Usage:
#   scripts/dev-cancel-order.sh <ORDER_ID> [REASON] [PORT]
# Env overrides:
#   TEA_SERVER_PORT: set API port (default 9292)
#   CANCEL_REASON: optional cancel reason if not provided as 2nd arg
# Notes:
# - Attempts to reuse token from build-ci-logs/admin_login_response.json
# - Falls back to login via /api/v1/auth/login (admin/pass) or dev-login

if [[ ${1:-} == "" ]]; then
  echo "Usage: $0 <ORDER_ID> [REASON] [PORT]"
  exit 1
fi

ORDER_ID="$1"
REASON="${2:-${CANCEL_REASON:-用户取消}}"
PORT_USE="${3:-${TEA_SERVER_PORT:-9292}}"
BASE_URL="http://127.0.0.1:${PORT_USE}"
API_PREFIX="${BASE_URL}/api/v1"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$REPO_ROOT/build-ci-logs"
mkdir -p "$LOG_DIR"
RESP_FILE="$LOG_DIR/admin_login_response.json"

get_token_from_file() {
  local file="$1"
  if [[ -s "$file" ]]; then
    python3 - "$file" <<'PY'
import sys, json
try:
    obj = json.load(open(sys.argv[1]))
except Exception:
    print("")
    sys.exit(0)
data = obj.get('data') or obj
print((data.get('token') if isinstance(data, dict) else "") or "")
PY
  else
    echo ""
  fi
}

login_and_save_token() {
  # Try username/password
  curl -sS -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"pass"}' \
    "${API_PREFIX}/auth/login" -o "$RESP_FILE" || true
  local t
  t="$(get_token_from_file "$RESP_FILE")"
  if [[ -n "$t" ]]; then
    echo "$t"
    return 0
  fi
  # Try openid
  curl -sS -H "Content-Type: application/json" \
    -d '{"openid":"admin_openid"}' \
    "${API_PREFIX}/auth/login" -o "$RESP_FILE" || true
  t="$(get_token_from_file "$RESP_FILE")"
  if [[ -n "$t" ]]; then
    echo "$t"
    return 0
  fi
  # Fallback to dev-login
  curl -sS -H "Content-Type: application/json" \
    -d '{"openid":"admin_openid"}' \
    "${API_PREFIX}/user/dev-login" -o "$RESP_FILE" || true
  t="$(get_token_from_file "$RESP_FILE")"
  echo "$t"
}

TOKEN="$(get_token_from_file "$RESP_FILE")"
if [[ -z "$TOKEN" ]]; then
  echo "[dev-cancel] No cached token. Trying login..."
  TOKEN="$(login_and_save_token)"
fi

if [[ -z "$TOKEN" ]]; then
  echo "[dev-cancel] ERROR: Failed to obtain admin token. See $RESP_FILE"
  exit 2
fi

echo "[dev-cancel] Using token (truncated): ${TOKEN:0:16}..."
CANCEL_URL="${API_PREFIX}/orders/${ORDER_ID}/cancel"
HTTP_CODE=$(curl -sS -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"reason\":\"${REASON//\"/\\\"}\"}" \
  -o "$LOG_DIR/cancel_${ORDER_ID}.json" -w "%{http_code}" "$CANCEL_URL" || true)

echo "[dev-cancel] POST $CANCEL_URL -> HTTP $HTTP_CODE"
if [[ -s "$LOG_DIR/cancel_${ORDER_ID}.json" ]]; then
  echo "[dev-cancel] Response saved: $LOG_DIR/cancel_${ORDER_ID}.json"
fi

if [[ "$HTTP_CODE" != "200" ]]; then
  echo "[dev-cancel] Cancel failed (HTTP $HTTP_CODE)."
  exit 3
fi

echo "[dev-cancel] Cancel succeeded for order ${ORDER_ID}."
