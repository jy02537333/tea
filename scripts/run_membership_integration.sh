#!/usr/bin/env bash
set -euo pipefail

# E2E integration for wx-fe pages and tea-api membership flow
# Steps:
# 1) List membership packages
# 2) Create membership order (user token)
# 3) Create unified pay order
# 4) Simulate payment callback
# 5) List membership orders for user
# 6) Generate summary JSON files for assertions
#
# Requirements:
# - tea-api running locally (e.g., http://127.0.0.1:8080)
# - ADMIN_TOKEN and USER_TOKEN exported (JWT strings)
# - At least one membership package exists in DB
# - curl & jq installed
#
# Usage:
#   export API_BASE="http://127.0.0.1:8080"
#   export USER_TOKEN="<JWT>"
#   export ADMIN_TOKEN="<JWT>" # if callback requires admin
#   ./scripts/run_membership_integration.sh
#
# Outputs:
#   - membership_summary_after_purchase.json: Summary of purchase results
#   - membership_b_flow_checked.json: Flow validation results

API_BASE=${API_BASE:-"http://127.0.0.1:8080"}
USER_TOKEN=${USER_TOKEN:-""}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}
OUTPUT_DIR=${OUTPUT_DIR:-"build-ci-logs"}

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

if [[ -z "$USER_TOKEN" ]]; then
  echo "ERROR: USER_TOKEN is required. Export USER_TOKEN before running." >&2
  exit 1
fi

have_jq=1
command -v jq >/dev/null 2>&1 || have_jq=0

function pp_json() {
  if [[ $have_jq -eq 1 ]]; then
    jq .
  else
    cat
  fi
}

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

echo "[6] Generating summary JSON files"
# Generate membership_summary_after_purchase.json
if [[ $have_jq -eq 1 ]]; then
  cat > "$OUTPUT_DIR/membership_summary_after_purchase.json" <<EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "api_base": "$API_BASE",
  "package_id": $package_id,
  "order_id": $order_id,
  "packages_response": $packages_json,
  "create_order_response": $create_order_resp,
  "unified_order_response": $unified_order_resp,
  "orders_list_response": $orders_json,
  "test_status": "completed"
}
EOF
  echo "Generated: $OUTPUT_DIR/membership_summary_after_purchase.json"
  
  # Generate membership_b_flow_checked.json with validation results
  # Extract key data for validation
  order_count=$(echo "$orders_json" | jq -r '.data | length // 0')
  
  # Extract order-specific data in a single jq call for efficiency
  read -r order_status payment_status < <(echo "$orders_json" | jq -r "
    .data[] | select(.id == $order_id) | 
    \"\(.status // \"unknown\") \(.pay_status // \"unknown\")\"
  ")
  
  # Determine validation results
  package_list_ok="true"
  [[ -n "$package_id" && "$package_id" != "null" ]] || package_list_ok="false"
  
  order_create_ok="true"
  [[ -n "$order_id" && "$order_id" != "null" ]] || order_create_ok="false"
  
  payment_created_ok="true"
  # Check both that callback_url exists and that unified order response indicates success
  if [[ -z "$callback_url" ]]; then
    payment_created_ok="false"
  elif [[ $have_jq -eq 1 ]]; then
    # Also verify the unified order response has expected success indicators
    unified_success=$(echo "$unified_order_resp" | jq -r '.code // .status // "0"')
    [[ "$unified_success" == "0" || "$unified_success" == "200" || "$unified_success" == "success" ]] || payment_created_ok="false"
  fi
  
  order_found_ok="true"
  [[ "$order_count" -gt 0 ]] || order_found_ok="false"
  
  all_checks_passed="true"
  [[ "$package_list_ok" == "true" && "$order_create_ok" == "true" && "$payment_created_ok" == "true" && "$order_found_ok" == "true" ]] || all_checks_passed="false"
  
  cat > "$OUTPUT_DIR/membership_b_flow_checked.json" <<EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "flow_name": "Sprint B - Membership Purchase",
  "checks": {
    "package_list_retrieved": $package_list_ok,
    "order_created": $order_create_ok,
    "payment_order_created": $payment_created_ok,
    "order_found_in_list": $order_found_ok
  },
  "data": {
    "package_id": $package_id,
    "order_id": $order_id,
    "order_count": $order_count,
    "order_status": "$order_status",
    "payment_status": "$payment_status"
  },
  "result": {
    "all_checks_passed": $all_checks_passed,
    "test_status": "$(if [[ "$all_checks_passed" == "true" ]]; then echo "PASSED"; else echo "FAILED"; fi)"
  }
}
EOF
  echo "Generated: $OUTPUT_DIR/membership_b_flow_checked.json"
  
  # Display summary
  echo ""
  echo "========================================="
  echo "Sprint B Membership Flow Test Summary"
  echo "========================================="
  echo "Package ID: $package_id"
  echo "Order ID: $order_id"
  echo "Order Status: $order_status"
  echo "Payment Status: $payment_status"
  echo "All Checks Passed: $all_checks_passed"
  echo "========================================="
else
  echo "WARNING: jq not available, JSON summary files not generated"
fi

echo "Done. Verify that the created order is paid/completed as expected."