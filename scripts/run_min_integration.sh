#!/usr/bin/env bash
# Minimal integration flows for Sprint C evidence generation
# Produces build-ci-logs/** artifacts without failing the pipeline.

set -uo pipefail

API_BASE="${API_BASE:-http://127.0.0.1:9292}"
OUT_DIR="build-ci-logs"
mkdir -p "$OUT_DIR"

log() { echo "[min-int] $*" | tee -a "$OUT_DIR/min-integration.log"; }

# 1) Ensure tokens are prepared (admin/user)
log "Preparing tokens via scripts/prepare_tokens.sh"
API_BASE="$API_BASE" bash ./scripts/prepare_tokens.sh >> "$OUT_DIR/min-integration.log" 2>&1 || true

if [[ -f "$OUT_DIR/tokens.env" ]]; then
  # shellcheck disable=SC1090
  source "$OUT_DIR/tokens.env"
fi

ADMIN_TOKEN="${ADMIN_TOKEN:-}"
USER_TOKEN="${USER_TOKEN:-}"

log "API_BASE: $API_BASE"
log "ADMIN_TOKEN: ${ADMIN_TOKEN:0:12}...(${#ADMIN_TOKEN} chars)"
log "USER_TOKEN : ${USER_TOKEN:0:12}...(${#USER_TOKEN} chars)"

# Helper curl wrappers (never fail the script)
post_json_auth() {
  local url="$1" token="$2" body="$3" out="$4"
  curl -sS -X POST "$url" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$body" \
    > "$out" 2>> "$OUT_DIR/min-integration.log" || true
}

get_auth() {
  local url="$1" token="$2" out="$3"
  curl -sS -H "Authorization: Bearer $token" "$url" \
    > "$out" 2>> "$OUT_DIR/min-integration.log" || true
}

status_code_auth() {
  local url="$1" token="$2" out="$3"
  curl -sS -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $token" "$url" \
    > "$out" 2>> "$OUT_DIR/min-integration.log" || true
}

status_code_auth_post() {
  local url="$1" token="$2" out="$3" body="${4:-"{}"}"
  curl -sS -o /dev/null -w "%{http_code}\n" -X POST \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "$body" "$url" \
    > "$out" 2>> "$OUT_DIR/min-integration.log" || true
}

# SC1: OSS Policy (Admin)
if [[ -n "$ADMIN_TOKEN" ]]; then
  log "SC1: Requesting OSS policy"
  # 依据接口定义补齐必需字段，提升返回包含 policy/signature/accessKeyId/expire_at 的概率
  post_json_auth "$API_BASE/api/v1/admin/storage/oss/policy" "$ADMIN_TOKEN" \
    '{"business":"product_image","file_name":"demo.jpg","content_type":"image/jpeg","file_size":12345}' \
    "$OUT_DIR/get_oss_policy.json"
else
  log "SC1: Skipped (missing ADMIN_TOKEN)"
fi

# SC2: Referral record and stats (User)
if [[ -n "$USER_TOKEN" ]]; then
  log "SC2: Creating referral record and querying stats"
  post_json_auth "$API_BASE/api/v1/referral/record" "$USER_TOKEN" \
    '{"referrer_id":1001, "referred_user_id":1002, "source":"qr"}' \
    "$OUT_DIR/referral_record.json"
  get_auth "$API_BASE/api/v1/users/1001/referral-stats" "$USER_TOKEN" \
    "$OUT_DIR/referral_stats.json"
  get_auth "$API_BASE/api/v1/users/1001/referred-users?page=1&size=20" "$USER_TOKEN" \
    "$OUT_DIR/referred_users.json"
else
  log "SC2: Skipped (missing USER_TOKEN)"
fi

# SC4: User commissions
if [[ -n "$USER_TOKEN" ]]; then
  log "SC4: Commission calculate preview"
  post_json_auth "$API_BASE/api/v1/commissions/calculate" "$USER_TOKEN" \
    '{"payer_user_id":1002,"referrer_user_id":1001,"items":[{"sku_id":1,"unit_price_cents":9800,"quantity":2,"discount_cents":0}],"shipping_cents":0,"coupon_cents":0,"order_level_discount_cents":0}' \
    "$OUT_DIR/commission_preview.json"

  log "SC4: Commission create record (example)"
  post_json_auth "$API_BASE/api/v1/commissions" "$USER_TOKEN" \
    '{"order_id":156,"breakdown":[]}' \
    "$OUT_DIR/commission_create_resp.json"

  log "SC4: Commission list and summary"
  get_auth "$API_BASE/api/v1/users/1001/commissions?page=1&size=20" "$USER_TOKEN" \
    "$OUT_DIR/commissions_list.json"
  get_auth "$API_BASE/api/v1/users/1001/commissions/summary" "$USER_TOKEN" \
    "$OUT_DIR/commissions_summary.json"
else
  log "SC4: Skipped (missing USER_TOKEN)"
fi

if [[ -n "$ADMIN_TOKEN" ]]; then
  log "SC4: Admin commission release (optional)"
  # 改为 POST 调用，避免 GET 404；若需要 payload，可在此补充字段
  post_json_auth "$API_BASE/api/v1/admin/finance/commission/release" "$ADMIN_TOKEN" '{}' \
    "$OUT_DIR/commission_release_resp.json"
  get_auth "$API_BASE/api/v1/admin/finance/summary" "$ADMIN_TOKEN" \
    "$OUT_DIR/finance_summary_after_release.json"
fi

# SC5: Partner purchase/upgrade (User)
if [[ -n "$USER_TOKEN" ]]; then
  log "SC5: Partner packages, purchase and upgrade"
  get_auth "$API_BASE/api/v1/partner/packages" "$USER_TOKEN" \
    "$OUT_DIR/partner_packages.json"
  post_json_auth "$API_BASE/api/v1/partner/purchase" "$USER_TOKEN" \
    '{"package_id":1,"pay_method":"wechat"}' \
    "$OUT_DIR/partner_purchase_resp.json"
  post_json_auth "$API_BASE/api/v1/partner/upgrade" "$USER_TOKEN" \
    '{"package_id":2}' \
    "$OUT_DIR/partner_upgrade_resp.json"
  get_auth "$API_BASE/api/v1/users/me/summary" "$USER_TOKEN" \
    "$OUT_DIR/membership_upgrade_summary.json"
else
  log "SC5: Skipped (missing USER_TOKEN)"
fi

# SC6: Store backend accept/reject/print (Admin)
if [[ -n "$ADMIN_TOKEN" ]]; then
  log "SC6: Store order list, accept, reject, print"
  get_auth "$API_BASE/api/v1/stores/1/orders?page=1&size=20" "$ADMIN_TOKEN" \
    "$OUT_DIR/store_orders_list.json"
  post_json_auth "$API_BASE/api/v1/stores/1/orders/156/accept" "$ADMIN_TOKEN" \
    '{"store_id":1,"order_id":156}' \
    "$OUT_DIR/store_order_accept_resp.json"
  post_json_auth "$API_BASE/api/v1/stores/1/orders/209/reject" "$ADMIN_TOKEN" \
    '{"store_id":1,"order_id":209,"reason":"out_of_stock"}' \
    "$OUT_DIR/store_order_reject_resp.json"
  post_json_auth "$API_BASE/api/v1/stores/1/print/jobs" "$ADMIN_TOKEN" \
    '{"store_id":1,"order_id":156,"template_type":"receipt","copies":1}' \
    "$OUT_DIR/print_job_resp.json"
else
  log "SC6: Skipped (missing ADMIN_TOKEN)"
fi

# SC7: Router existence checks (Admin)
if [[ -n "$ADMIN_TOKEN" ]]; then
  log "SC7: Router checks for admin endpoints"
  status_code_auth "$API_BASE/api/v1/admin/partner-levels" "$ADMIN_TOKEN" \
    "$OUT_DIR/router_check_partner_levels.txt"
  # 使用 POST 探测需要 POST 的路由，避免 404 误报
  status_code_auth_post "$API_BASE/api/v1/admin/finance/commission/release" "$ADMIN_TOKEN" \
    "$OUT_DIR/router_check_commission_release.txt" '{}'
else
  log "SC7: Skipped (missing ADMIN_TOKEN)"
fi

log "Minimal integration finished. Artifacts in $OUT_DIR/"
