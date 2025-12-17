#!/usr/bin/env bash
set -euo pipefail

###############################
# E2E integration for wx-fe pages and tea-api membership flow
#
# Steps:
# 1) List membership packages
# 2) Create membership order (user token)
# 3) Create unified pay order
# 4) Simulate payment callback
# 5) List membership orders for user
# 6) Fetch /api/v1/users/me/summary and check membership level
#    -> 输出证据文件，供 CI 严格校验 Sprint B 关键路径
#
# Requirements:
# - tea-api running locally (e.g., http://127.0.0.1:9292)
# - curl & jq installed
# - USER_TOKEN（JWT）可以通过环境变量传入，或由本脚本自动从
#   build-ci-logs/api_validation_stateful 及 admin_login_response.json 中提取
#
# Usage (local):
#   export API_BASE="http://127.0.0.1:9292"
#   export USER_TOKEN="<JWT>"   # 可选，若不提供则尝试自动发现
#   export ADMIN_TOKEN="<JWT>"  # 可选，仅在部分环境下模拟回调需要
#   ./scripts/run_membership_integration.sh
###############################

API_BASE=${API_BASE:-"http://127.0.0.1:9292"}
USER_TOKEN=${USER_TOKEN:-""}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT_DIR/build-ci-logs"
mkdir -p "$LOG_DIR"
SUMMARY_FILE="$LOG_DIR/local_api_summary.txt"
touch "$SUMMARY_FILE"

have_jq=1
command -v jq >/dev/null 2>&1 || have_jq=0

function pp_json() {
  if [[ $have_jq -eq 1 ]]; then
    jq .
  else
    cat
  fi
}

# 自动从已有日志中发现 USER_TOKEN（优先使用 stateful/dev-login 日志）
if [[ -z "$USER_TOKEN" ]]; then
  if [[ $have_jq -eq 1 ]]; then
    CANDIDATES=(
      "$ROOT_DIR/build-ci-logs/api_validation_stateful/POST__api_v1_auth_login.json"
      "$ROOT_DIR/build-ci-logs/api_validation_stateful/POST__api_v1_auth_login_openid.json"
      "$ROOT_DIR/build-ci-logs/admin_login_response.json"
    )
    for f in "${CANDIDATES[@]}"; do
      if [[ -f "$f" ]]; then
        USER_TOKEN=$(jq -r '.data.token // .token // empty' "$f" 2>/dev/null || echo "")
        if [[ -n "$USER_TOKEN" ]]; then
          echo "[membership] USER_TOKEN discovered from $f" >&2
          break
        fi
      fi
    done
  fi
fi

if [[ -z "$USER_TOKEN" ]]; then
  echo "ERROR: USER_TOKEN is required and could not be discovered from logs. Export USER_TOKEN or run stateful login first." >&2
  exit 1
fi

function api_get() {
  local path="$1"
  curl -sS -H "Authorization: Bearer $USER_TOKEN" "$API_BASE$path"
}

function api_post_json() {
  local path="$1"; shift
  local json="$1"; shift || true
  curl -sS -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" -X POST -d "$json" "$API_BASE$path"
}

function api_post_admin_json() {
  local path="$1"; shift
  local json="$1"; shift || true
  curl -sS -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" -X POST -d "$json" "$API_BASE$path"
}

echo "[1] List membership packages"
packages_json=$(api_get "/api/v1/membership-packages")
echo "$packages_json" | pp_json

if [[ $have_jq -eq 1 ]]; then
  package_id=$(echo "$packages_json" | jq -r '.data[0].id // empty')
else
  # Fallback: naive grep (expects id in response). Adjust as needed.
  package_id=$(echo "$packages_json" | grep -Eo '"id"\s*:\s*[0-9]+' | head -n1 | grep -Eo '[0-9]+')
fi

if [[ -z "$package_id" ]]; then
  echo "ERROR: No membership package found. Please seed a package before testing." >&2
  exit 1
fi

echo "Selected package_id=$package_id"

echo "[2] Create membership order"
create_order_resp=$(api_post_json "/api/v1/membership-orders" "{\"package_id\": $package_id, \"remark\": \"E2E test\"}")
echo "$create_order_resp" | pp_json

if [[ $have_jq -eq 1 ]]; then
  order_id=$(echo "$create_order_resp" | jq -r '.data.order_id // empty')
else
  order_id=$(echo "$create_order_resp" | grep -Eo '"order_id"\s*:\s*[0-9]+' | head -n1 | grep -Eo '[0-9]+')
fi

if [[ -z "$order_id" ]]; then
  echo "ERROR: Failed to create membership order (no order_id)." >&2
  exit 1
fi

echo "Created order_id=$order_id"

echo "[3] Create unified pay order"
unified_order_resp=$(api_post_json "/api/v1/payments/unified-order" "{\"order_id\": $order_id, \"payment_method\": \"weapp\"}")
echo "$unified_order_resp" | pp_json

if [[ $have_jq -eq 1 ]]; then
  callback_url=$(echo "$unified_order_resp" | jq -r '.data.mock_callback_url // empty')
else
  callback_url=$(echo "$unified_order_resp" | grep -Eo '"mock_callback_url"\s*:\s*"[^"]+"' | head -n1 | sed -E 's/.*:\s*"([^"]+)"/\1/')
fi

if [[ -z "$callback_url" ]]; then
  echo "WARNING: No mock_callback_url returned. If your environment uses real payment, skip callback simulation." >&2
else
  echo "[4] Simulate payment callback"
  # Some setups may require admin token; try user first, fallback to admin if provided
  callback_resp=$(curl -sS -H "Authorization: Bearer $USER_TOKEN" -X POST "$callback_url" || true)
  if [[ -n "$ADMIN_TOKEN" ]]; then
    # Attempt with admin if user fails silently
    callback_resp_admin=$(api_post_admin_json "${callback_url#${API_BASE}}" "{}" || true)
    echo "$callback_resp_admin" | pp_json
  else
    echo "$callback_resp" | pp_json
  fi
fi

echo "[5] List membership orders (user)"
orders_json=$(api_get "/api/v1/orders?order_type=4")
echo "$orders_json" | pp_json

echo "[6] Fetch user summary after membership purchase"
summary_json=$(api_get "/api/v1/users/me/summary")
echo "$summary_json" | pp_json

if [[ $have_jq -eq 1 ]]; then
  echo "$summary_json" > "$LOG_DIR/membership_summary_after_purchase.json" || true
  level=$(echo "$summary_json" | jq -r '.data.membership // .membership // empty')
  # 约定：非空且不为 visitor/gues t 视为已升级
  ok=false
  if [[ -n "$level" && "$level" != "visitor" && "$level" != "guest" ]]; then
    ok=true
  fi
  check_obj=$(jq -n --arg level "$level" --argjson ok "$([[ "$ok" == "true" ]] && echo true || echo false)" '{membership_level: $level, ok: $ok, source: "membership_summary_after_purchase.json"}')
  echo "$check_obj" > "$LOG_DIR/membership_b_flow_checked.json" || true
  echo "membership_flow: level=${level:-""}, ok=$ok" >> "$SUMMARY_FILE" || true
else
  echo "WARNING: jq not available; cannot persist membership evidence JSON" >&2
fi

echo "Done. Membership flow executed; evidence (if jq available) stored under $LOG_DIR."