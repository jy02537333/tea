#!/usr/bin/env bash
set -euo pipefail

LOG_DIR=build-ci-logs
ADMIN_LOGIN_JSON="$LOG_DIR/admin_login_response.json"

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
