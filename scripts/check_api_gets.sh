#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/build-ci-logs"
mkdir -p "$OUT_DIR"
OUT_FILE="$OUT_DIR/api_get_check.txt"
: > "$OUT_FILE"
BASE="http://127.0.0.1:9292"

paths=(
  "/api/v1/health"
  "/api/v1/products"
  "/api/v1/products/1"
  "/api/v1/categories"
  "/api/v1/user/info"
  "/api/v1/user/1"
  "/api/v1/user/interest-records"
  "/api/v1/cart"
  "/api/v1/orders"
  "/api/v1/orders/1"
  "/api/v1/stores"
  "/api/v1/stores/1"
  "/api/v1/coupons"
  "/api/v1/user/coupons"
  "/api/v1/admin/users"
  "/api/v1/admin/stores/1/orders/stats"
  "/api/v1/admin/stores/1/products"
  "/api/v1/admin/orders"
  "/api/v1/admin/orders/1"
  "/api/v1/admin/rbac/roles"
  "/api/v1/admin/rbac/permissions"
  "/api/v1/admin/logs/operations"
  "/api/v1/admin/logs/access"
  "/api/v1/admin/refunds"
  "/api/v1/admin/payments"
  "/api/v1/admin/withdraws"
  "/api/v1/admin/finance/summary"
  "/api/v1/admin/finance/reconcile/diff"
  "/api/v1/auth/me"
  "/api/v1/auth/captcha"
)

echo "API GET reachability check: $(date -Iseconds)" > "$OUT_FILE"
for p in "${paths[@]}"; do
  url="$BASE$p"
  echo -n "Checking $url ... " | tee -a "$OUT_FILE"
  # try GET with 3s timeout, save body small preview
  status=$(curl -sS -o /tmp/_api_check_body.tmp -w "%{http_code}" --max-time 3 "$url" 2>/dev/null || echo "000")
  if [ -s /tmp/_api_check_body.tmp ]; then
    preview=$(head -n 5 /tmp/_api_check_body.tmp | tr '\n' ' ' | sed 's/\s\+/ /g')
  else
    preview="<no-body>"
  fi
  echo "HTTP $status - $preview" | tee -a "$OUT_FILE"
done

rm -f /tmp/_api_check_body.tmp

echo "Done. Results saved to $OUT_FILE"
