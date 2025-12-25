#!/usr/bin/env bash
set -euo pipefail

LOG_DIR=build-ci-logs
ADMIN_LOGIN_JSON="$LOG_DIR/admin_login_response.json"
REQUIRE_ORDER_CHECK=${REQUIRE_ORDER_CHECK:-0}

if [ ! -f "$ADMIN_LOGIN_JSON" ]; then
  echo "Admin login response not found: $ADMIN_LOGIN_JSON"
  exit 2
fi

# try common token fields
TOKEN=$(jq -r '.data.token // .token // .access_token // empty' "$ADMIN_LOGIN_JSON") || true
if [ -z "$TOKEN" ]; then
  echo "Admin token not found in $ADMIN_LOGIN_JSON"
  exit 2
fi

BASE=${API_BASE_URL:-http://127.0.0.1:9292}

endpoints=(
  "/api/v1/admin/orders"
  "/api/v1/admin/rbac/roles"
  "/api/v1/admin/stores/1/orders/stats"
)

FAIL=0
for ep in "${endpoints[@]}"; do
  url="$BASE$ep"
  echo "Checking $url"
  status=$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$url" || true)
  echo "$ep -> $status"
  if [ "$status" != "200" ]; then
    echo "ERROR: $url returned $status"
    FAIL=1
  fi
done

if [ $FAIL -ne 0 ]; then
  echo "One or more critical admin endpoints returned non-200"
  exit 3
fi

echo "All critical admin endpoints returned 200"

# --- Sprint A: Order amount deduction assertion ---
echo "Asserting Sprint A order amount deduction..."

# Prefer checked evidence file, then summary, then any order_detail file
ORDER_SUMMARY_JSON="$LOG_DIR/order_amounts_summary.json"

# Helper: find first matching file for a glob pattern
find_first() {
  local pattern="$1"
  local match
  match=$(compgen -G "$pattern" | head -n1 || true)
  if [ -n "$match" ]; then echo "$match"; fi
}

ORDER_CHECK_ANY=$(find_first "$LOG_DIR/order_detail_*_checked.json")
ORDER_DETAIL_ANY=$(find_first "$LOG_DIR/order_detail_[0-9]*.json")

if [ -n "$ORDER_CHECK_ANY" ] && [ -f "$ORDER_CHECK_ANY" ]; then
  if jq -e '.check==true' "$ORDER_CHECK_ANY" >/dev/null 2>&1; then
    echo "Amount check passed via $ORDER_CHECK_ANY"
  else
    echo "ERROR: Amount check failed (.check!=true) in $ORDER_CHECK_ANY"
    exit 4
  fi
elif [ -f "$ORDER_SUMMARY_JSON" ]; then
  if jq -e '.pay_amount == (.total_amount - .discount_amount)' "$ORDER_SUMMARY_JSON" >/dev/null 2>&1; then
    echo "Amount check passed via $ORDER_SUMMARY_JSON"
  else
    echo "ERROR: Amount check failed via $ORDER_SUMMARY_JSON"
    exit 4
  fi
elif [ -n "$ORDER_DETAIL_ANY" ] && [ -f "$ORDER_DETAIL_ANY" ]; then
  OID=$(basename "$ORDER_DETAIL_ANY" | sed -E 's/order_detail_([0-9]+)\.json/\1/')
  ORDER_CHECK_OUT="$LOG_DIR/order_detail_${OID}_checked.json"
  # Evidence JSON may have order fields under .data (older shape) or .data.order (newer shape).
  tmp_compute=$(jq '{id: (.data.id // .data.order.id),
                     store_id: (.data.store_id // .data.order.store_id),
                     total_amount: ((.data.total_amount // .data.order.total_amount)|tonumber),
                     discount_amount: ((.data.discount_amount // .data.order.discount_amount)|tonumber),
                     pay_amount: ((.data.pay_amount // .data.order.pay_amount)|tonumber),
                     check: (((.data.pay_amount // .data.order.pay_amount)|tonumber) == (((.data.total_amount // .data.order.total_amount)|tonumber) - ((.data.discount_amount // .data.order.discount_amount)|tonumber))) }' "$ORDER_DETAIL_ANY")
  echo "$tmp_compute" > "$ORDER_CHECK_OUT" || true
  if echo "$tmp_compute" | jq -e '.check==true' >/dev/null 2>&1; then
    echo "Amount check passed via $ORDER_DETAIL_ANY (computed -> $ORDER_CHECK_OUT)"
  else
    echo "ERROR: Amount check failed via $ORDER_DETAIL_ANY"
    exit 4
  fi
else
  if [ "$REQUIRE_ORDER_CHECK" = "1" ]; then
    echo "ERROR: No order verification files found while REQUIRE_ORDER_CHECK=1"
    echo "Expected one of: $LOG_DIR/order_detail_*_checked.json | $ORDER_SUMMARY_JSON | $LOG_DIR/order_detail_[0-9]*.json"
    exit 5
  else
    echo "WARNING: No order verification files found; skipping amount check"
  fi
fi
