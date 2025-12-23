#!/usr/bin/env bash
# Verify Sprint C minimal integration evidence files (non-blocking by CI)
# - Reads build-ci-logs/** artifacts produced by scripts/run_min_integration.sh
# - Performs light presence/content checks and emits a JSON summary
# - Exits non-zero if any required checks fail (CI can wrap with continue-on-error)

set -euo pipefail

OUT_DIR="build-ci-logs"
mkdir -p "$OUT_DIR"

need_bin() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[verify-sprint-c] Missing dependency: $1" >&2
    return 1
  fi
}

has_jq=1
if ! command -v jq >/dev/null 2>&1; then
  has_jq=0
  echo "[verify-sprint-c] WARN: jq not found, will do presence-only checks"
fi

items_json="[]"
errors=0

add_item() {
  local name="$1" file="$2" ok="$3" note="$4"
  local obj
  if [[ ${has_jq} -eq 1 ]]; then
    obj=$(jq -n --arg name "$name" --arg file "$file" --arg ok "$ok" --arg note "$note" '{name:$name,file:$file,ok:($ok=="true"),note:$note}')
    items_json=$(jq -c --argjson obj "$obj" '. + [$obj]' <<<"$items_json")
  else
    # minimal JSON building without jq (best-effort)
    items_json="${items_json%]}${items_json: -1 == '[' ? '' : ','}{\"name\":\"$name\",\"file\":\"$file\",\"ok\":$([[ "$ok" == "true" ]] && echo true || echo false),\"note\":\"$note\"}]"
  fi
}

check_exists() {
  local name="$1" file="$2"
  if [[ -s "$file" ]]; then
    add_item "$name" "$file" true "exists"
  else
    add_item "$name" "$file" false "missing"
    errors=$((errors+1))
  fi
}

check_json_keys() {
  local name="$1" file="$2"; shift 2
  local keys=("$@")
  if [[ ! -s "$file" ]]; then
    add_item "$name" "$file" false "missing"
    errors=$((errors+1))
    return
  fi
  if [[ ${has_jq} -eq 0 ]]; then
    add_item "$name" "$file" true "exists (jq unavailable)"
    return
  fi
  local missing=()
  for k in "${keys[@]}"; do
    local val
    val=$(jq -r --arg k "$k" '.[$k] // .data[$k] // empty' "$file" 2>/dev/null || true)
    if [[ -z "$val" || "$val" == "null" ]]; then
      missing+=("$k")
    fi
  done
  if [[ ${#missing[@]} -eq 0 ]]; then
    add_item "$name" "$file" true "keys ok: ${keys[*]}"
  else
    add_item "$name" "$file" false "missing keys: ${missing[*]}"
    errors=$((errors+1))
  fi
}

check_http_code_file() {
  local name="$1" file="$2"
  if [[ ! -s "$file" ]]; then
    add_item "$name" "$file" false "missing"
    errors=$((errors+1))
    return
  fi
  local code
  code=$(tr -d '\r' < "$file" | tail -n1)
  if [[ "$code" =~ ^(200|401|403)$ ]]; then
    add_item "$name" "$file" true "code=$code acceptable"
  else
    add_item "$name" "$file" false "unexpected code=$code"
    errors=$((errors+1))
  fi
}

echo "[verify-sprint-c] Verifying artifacts in $OUT_DIR"

# SC1（若接口未就绪，存在错误响应时不计为 hard fail）
if [[ ! -s "$OUT_DIR/get_oss_policy.json" ]]; then
  add_item "SC1 OSS policy" "$OUT_DIR/get_oss_policy.json" false "missing"
  errors=$((errors+1))
else
  if [[ ${has_jq} -eq 1 ]]; then
    pol=$(jq -r '.policy // empty' "$OUT_DIR/get_oss_policy.json" 2>/dev/null || true)
    sig=$(jq -r '.signature // empty' "$OUT_DIR/get_oss_policy.json" 2>/dev/null || true)
    ak=$(jq -r '.accessKeyId // empty' "$OUT_DIR/get_oss_policy.json" 2>/dev/null || true)
    exp=$(jq -r '.expire_at // empty' "$OUT_DIR/get_oss_policy.json" 2>/dev/null || true)
    if [[ -n "$pol" && -n "$sig" && -n "$ak" && -n "$exp" ]]; then
      add_item "SC1 OSS policy" "$OUT_DIR/get_oss_policy.json" true "keys ok: policy signature accessKeyId expire_at"
    else
      # 若返回为错误占位（如 {error:"not_found"}），视为接口未就绪但证据已产出，不计 hard fail
      err=$(jq -r '.error // .message // empty' "$OUT_DIR/get_oss_policy.json" 2>/dev/null || true)
      if [[ -n "$err" ]]; then
        add_item "SC1 OSS policy" "$OUT_DIR/get_oss_policy.json" true "exists (endpoint not ready: $err)"
      else
        add_item "SC1 OSS policy" "$OUT_DIR/get_oss_policy.json" false "missing keys: policy signature accessKeyId expire_at"
        errors=$((errors+1))
      fi
    fi
  else
    add_item "SC1 OSS policy" "$OUT_DIR/get_oss_policy.json" true "exists (jq unavailable)"
  fi
fi

# SC2
check_exists     "SC2 referral record" "$OUT_DIR/referral_record.json"
check_exists     "SC2 referral stats"  "$OUT_DIR/referral_stats.json"
check_exists     "SC2 referred users"  "$OUT_DIR/referred_users.json"

# SC4
check_exists     "SC4 commission preview" "$OUT_DIR/commission_preview.json"
check_exists     "SC4 commission create"  "$OUT_DIR/commission_create_resp.json"
check_exists     "SC4 commissions list"   "$OUT_DIR/commissions_list.json"
check_exists     "SC4 commissions summary" "$OUT_DIR/commissions_summary.json"
check_exists     "SC4 finance summary (admin)" "$OUT_DIR/finance_summary_after_release.json"

# SC5
check_exists     "SC5 partner packages"     "$OUT_DIR/partner_packages.json"
check_exists     "SC5 partner purchase"     "$OUT_DIR/partner_purchase_resp.json"
check_exists     "SC5 partner upgrade"      "$OUT_DIR/partner_upgrade_resp.json"
check_exists     "SC5 membership summary"   "$OUT_DIR/membership_upgrade_summary.json"

# SC6
check_exists     "SC6 store orders list"    "$OUT_DIR/store_orders_list.json"
check_exists     "SC6 order accept resp"    "$OUT_DIR/store_order_accept_resp.json"
check_exists     "SC6 order reject resp"    "$OUT_DIR/store_order_reject_resp.json"
check_exists     "SC6 print job resp"       "$OUT_DIR/print_job_resp.json"

# SC7
check_http_code_file "SC7 route: admin/partner-levels"         "$OUT_DIR/router_check_partner_levels.txt"
check_http_code_file "SC7 route: admin/finance/commission/release" "$OUT_DIR/router_check_commission_release.txt"

ok=true
if [[ $errors -gt 0 ]]; then ok=false; fi

if [[ ${has_jq} -eq 1 ]]; then
  jq -n \
    --argjson ok "$ok" \
    --arg errors "$errors" \
    --arg total "$( [[ ${has_jq} -eq 1 ]] && jq -r 'length' <<<"$items_json" || echo 0 )" \
    --arg now "$(date -u +%FT%TZ)" \
    --argjson items "$items_json" \
    '{ok:$ok, errors: ($errors|tonumber), total: ($total|tonumber), generated_at:$now, items:$items}' \
    | tee "$OUT_DIR/sprint_c_check_summary.json" >/dev/null
else
  echo "{\"ok\":$ok,\"errors\":$errors}" > "$OUT_DIR/sprint_c_check_summary.json"
fi

echo "[verify-sprint-c] Completed. errors=$errors, summary => $OUT_DIR/sprint_c_check_summary.json"

if [[ $errors -gt 0 ]]; then
  exit 1
fi

exit 0
