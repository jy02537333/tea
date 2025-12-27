#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR=$(dirname "$0")/..
BASE_URL="${API_BASE:-http://127.0.0.1:9292}"
mkdir -p build-ci-logs
RESP_FILE="build-ci-logs/admin_login_response.json"

echo "Calling dev-login to obtain admin token..."
curl -sS -H 'Content-Type: application/json' -d '{"openid":"admin_openid"}' "$BASE_URL/api/v1/user/dev-login" -o "$RESP_FILE" || true
TOKEN=$(jq -r '.data.token // .token // empty' "$RESP_FILE" 2>/dev/null || true)
if [ -z "$TOKEN" ]; then
  echo "ERROR: failed to obtain token. See $RESP_FILE"
  exit 1
fi

echo "Token acquired (len=${#TOKEN})"

# create a withdrawal for user id 3 (adjust USER_ID as needed)
USER_ID=${1:-3}
PAYLOAD='{"bank_account_id":0,"amount_cents":10000,"currency":"CNY","note":"automation test withdraw"}'
OUT_FILE="build-ci-logs/test_withdrawal_response.json"

echo "Creating withdrawal for user $USER_ID..."
HTTP_CODE=$(curl -sS -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "$PAYLOAD" "$BASE_URL/api/v1/users/$USER_ID/withdrawals" -o "$OUT_FILE" -w '%{http_code}') || HTTP_CODE=0

echo "Response saved to $OUT_FILE"
cat "$OUT_FILE"

echo "Done."

# If initial withdrawal returned 400 (likely insufficient funds), credit the wallet and retry
if [ "$HTTP_CODE" = "400" ]; then
  echo "Initial withdrawal attempt failed (HTTP 400). Attempting admin credit to top up wallet..."
  CREDIT_CENTS=${CREDIT_CENTS:-20000}
  CREDIT_PAYLOAD=$(jq -n --argjson v "$CREDIT_CENTS" --arg r "e2e credit" '{amount_cents: $v, remark: $r}')
  CREDIT_OUT="build-ci-logs/test_credit_response.json"
  CREDIT_HTTP=$(curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "$CREDIT_PAYLOAD" "$BASE_URL/api/v1/admin/recharge/users/$USER_ID/credit" -o "$CREDIT_OUT" -w '%{http_code}') || CREDIT_HTTP=0
  echo "Credit response (HTTP $CREDIT_HTTP):" && cat "$CREDIT_OUT"

  echo "Retrying withdrawal..."
  HTTP_CODE=$(curl -sS -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "$PAYLOAD" "$BASE_URL/api/v1/users/$USER_ID/withdrawals" -o "$OUT_FILE" -w '%{http_code}') || HTTP_CODE=0
  echo "Final withdrawal response (HTTP $HTTP_CODE) saved to $OUT_FILE"
  cat "$OUT_FILE"
fi