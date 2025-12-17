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

ORDER_CHECK_JSON="$LOG_DIR/order_detail_156_checked.json"
ORDER_DETAIL_JSON="$LOG_DIR/order_detail_156.json"
ORDER_SUMMARY_JSON="$LOG_DIR/order_amounts_summary.json"

if [ -f "$ORDER_CHECK_JSON" ]; then
  if jq -e '.check==true' "$ORDER_CHECK_JSON" >/dev/null 2>&1; then
    echo "Amount check passed via $ORDER_CHECK_JSON"
  else
    echo "ERROR: Amount check failed (.check!=true) in $ORDER_CHECK_JSON"
    exit 4
  fi
elif [ -f "$ORDER_SUMMARY_JSON" ]; then
  if jq -e '.pay_amount == (.total_amount - .discount_amount)' "$ORDER_SUMMARY_JSON" >/dev/null 2>&1; then
    echo "Amount check passed via $ORDER_SUMMARY_JSON"
  else
    echo "ERROR: Amount check failed via $ORDER_SUMMARY_JSON"
    exit 4
  fi
elif [ -f "$ORDER_DETAIL_JSON" ]; then
  tmp_compute=$(jq '{id: .data.id,
                     store_id: .data.store_id,
                     total_amount: (.data.total_amount|tonumber),
                     discount_amount: (.data.discount_amount|tonumber),
                     pay_amount: (.data.pay_amount|tonumber),
                     check: ((.data.pay_amount|tonumber) == ((.data.total_amount|tonumber) - (.data.discount_amount|tonumber))) }' "$ORDER_DETAIL_JSON")
  echo "$tmp_compute" > "$ORDER_CHECK_JSON" || true
  if echo "$tmp_compute" | jq -e '.check==true' >/dev/null 2>&1; then
    echo "Amount check passed via $ORDER_DETAIL_JSON (computed -> $ORDER_CHECK_JSON)"
  else
    echo "ERROR: Amount check failed via $ORDER_DETAIL_JSON"
    exit 4
  fi
else
  if [ "$REQUIRE_ORDER_CHECK" = "1" ]; then
    echo "ERROR: No order verification files found while REQUIRE_ORDER_CHECK=1"
    echo "Expected one of: $ORDER_CHECK_JSON | $ORDER_SUMMARY_JSON | $ORDER_DETAIL_JSON"
    exit 5
  else
    echo "WARNING: No order verification files found; skipping amount check"
  fi
fi
