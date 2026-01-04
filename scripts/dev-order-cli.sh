#!/usr/bin/env bash
set -euo pipefail

# dev-order-cli.sh
# Unified developer CLI for order ops: deliver / receive / cancel
# Usage:
#   scripts/dev-order-cli.sh <deliver|receive|cancel> <ORDER_ID> [options]
# Options:
#   -p, --port <PORT>       API port (default from TEA_SERVER_PORT or 9292)
#   -r, --reason <REASON>   Cancel reason (only for cancel; default "用户取消")
# Env overrides:
#   TEA_SERVER_PORT         API port fallback
# Notes:
# - Reuses token from build-ci-logs/admin_login_response.json when possible
# - Falls back to login via /api/v1/auth/login (admin/pass or openid) or dev-login

usage() {
  cat <<USAGE
Usage:
  $0 <deliver|receive|cancel> <ORDER_ID> [options]

Options:
  -p, --port <PORT>       API port (default from TEA_SERVER_PORT or 9292)
  -r, --reason <REASON>   Cancel reason (only for cancel; default "用户取消")

Examples:
  $0 deliver 123 -p 9292
  $0 receive 123 --port 9393
  $0 cancel 123 -r "超时未支付" -p 9292
USAGE
}

if [[ ${1:-} == "" || ${2:-} == "" ]]; then
  usage
  exit 1
fi

CMD="$1"
ORDER_ID="$2"
shift 2

PORT_USE="${TEA_SERVER_PORT:-9292}"
REASON="用户取消"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -p|--port)
      PORT_USE="$2"; shift 2;;
    -r|--reason)
      REASON="$2"; shift 2;;
    -h|--help)
      usage; exit 0;;
    *)
      echo "Unknown option: $1"; usage; exit 1;;
  esac
done

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
    echo "$t"; return 0
  fi
  # Try openid
  curl -sS -H "Content-Type: application/json" \
    -d '{"openid":"admin_openid"}' \
    "${API_PREFIX}/auth/login" -o "$RESP_FILE" || true
  t="$(get_token_from_file "$RESP_FILE")"
  if [[ -n "$t" ]]; then
    echo "$t"; return 0
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
  echo "[dev-order-cli] No cached token. Trying login..."
  TOKEN="$(login_and_save_token)"
fi

if [[ -z "$TOKEN" ]]; then
  echo "[dev-order-cli] ERROR: Failed to obtain admin token. See $RESP_FILE"
  exit 2
fi

echo "[dev-order-cli] Using token (truncated): ${TOKEN:0:16}..."
HTTP_CODE=""
case "$CMD" in
  deliver)
    URL="${API_PREFIX}/orders/${ORDER_ID}/deliver"
    OUT="$LOG_DIR/deliver_${ORDER_ID}.json"
    HTTP_CODE=$(curl -sS -X POST -H "Authorization: Bearer $TOKEN" -o "$OUT" -w "%{http_code}" "$URL" || true)
    ;;
  receive)
    URL="${API_PREFIX}/orders/${ORDER_ID}/receive"
    OUT="$LOG_DIR/receive_${ORDER_ID}.json"
    HTTP_CODE=$(curl -sS -X POST -H "Authorization: Bearer $TOKEN" -o "$OUT" -w "%{http_code}" "$URL" || true)
    ;;
  cancel)
    URL="${API_PREFIX}/orders/${ORDER_ID}/cancel"
    OUT="$LOG_DIR/cancel_${ORDER_ID}.json"
    BODY=$(printf '{"reason":"%s"}' "${REASON}")
    HTTP_CODE=$(curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$BODY" -o "$OUT" -w "%{http_code}" "$URL" || true)
    ;;
  *)
    echo "Unknown command: $CMD"; usage; exit 1;;

esac

echo "[dev-order-cli] POST $URL -> HTTP $HTTP_CODE"
if [[ -s "$OUT" ]]; then
  echo "[dev-order-cli] Response saved: $OUT"
fi

if [[ "$HTTP_CODE" != "200" ]]; then
  echo "[dev-order-cli] Operation '$CMD' failed (HTTP $HTTP_CODE)."
  exit 3
fi

echo "[dev-order-cli] Operation '$CMD' succeeded for order ${ORDER_ID}."
