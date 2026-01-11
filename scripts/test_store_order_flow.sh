#!/usr/bin/env bash
set -euo pipefail

# 新流程自动测试：
# 门店注册 -> 门店设置桌号(订单维度 set-table/或下单携带) -> 上架商品 -> 用户下单 -> 付款 -> 门店出货
#
# Usage:
#   bash scripts/test_store_order_flow.sh
#
# Env:
#   API_BASE    默认 http://127.0.0.1:9292
#   PRODUCT_ID  要上架并下单的商品ID（可选；默认从 build-ci-logs 推断或 163）
#   TABLE_ID    默认 1
#   TABLE_NO    默认 A12

API_BASE="${API_BASE:-http://127.0.0.1:9292}"
PRODUCT_ID="${PRODUCT_ID:-}"
TABLE_ID="${TABLE_ID:-1}"
TABLE_NO="${TABLE_NO:-A12}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: 缺少依赖命令: $1" >&2
    exit 1
  }
}

need_cmd curl
need_cmd jq

if [[ -z "$PRODUCT_ID" ]] && [[ -f build-ci-logs/upsert_store1_product_resp.json ]]; then
  PRODUCT_ID=$(jq -r '.data.product_id // empty' build-ci-logs/upsert_store1_product_resp.json 2>/dev/null || true)
fi
if [[ -z "$PRODUCT_ID" ]]; then
  PRODUCT_ID=163
fi

json_get() {
  local json="$1"
  local jq_expr="$2"
  echo "$json" | jq -r "$jq_expr"
}

assert_ok_code() {
  local resp="$1"
  local code
  code=$(json_get "$resp" '.code // -1')
  if [[ "$code" != "0" ]]; then
    echo "ERROR: 请求失败: $resp" >&2
    exit 1
  fi
}

echo "========================================" >&2
echo "新流程自动测试" >&2
echo "========================================" >&2
echo "API_BASE=$API_BASE" >&2
echo "PRODUCT_ID=$PRODUCT_ID" >&2
echo "TABLE_ID=$TABLE_ID" >&2
echo "TABLE_NO=$TABLE_NO" >&2
echo "" >&2

# 0) 获取 admin token（dev-login: admin_openid 会自动赋予 admin 角色）
echo "[0/6] 获取 admin token..." >&2
ADMIN_LOGIN=$(curl -sS -X POST "$API_BASE/api/v1/user/dev-login" \
  -H "Content-Type: application/json" \
  -d '{"openid":"admin_openid"}')
assert_ok_code "$ADMIN_LOGIN"
ADMIN_TOKEN=$(json_get "$ADMIN_LOGIN" '.data.token // empty')
if [[ -z "$ADMIN_TOKEN" ]]; then
  echo "ERROR: admin token 获取失败: $ADMIN_LOGIN" >&2
  exit 1
fi
echo "✓ admin token OK" >&2

# 1) 门店注册（用普通用户 token 创建门店）
echo "[1/6] 门店注册..." >&2
OWNER_LOGIN=$(curl -sS -X POST "$API_BASE/api/v1/user/dev-login" \
  -H "Content-Type: application/json" \
  -d "{\"openid\":\"store_owner_$(date +%s)\"}")
assert_ok_code "$OWNER_LOGIN"
OWNER_TOKEN=$(json_get "$OWNER_LOGIN" '.data.token // empty')

STORE_CREATE=$(curl -sS -X POST "$API_BASE/api/v1/stores" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OWNER_TOKEN" \
  -d "{
    \"name\":\"测试门店_$(date +%s)\",
    \"address\":\"测试地址\",
    \"phone\":\"13800000000\",
    \"status\":1
  }")
assert_ok_code "$STORE_CREATE"
STORE_ID=$(json_get "$STORE_CREATE" '.data.id // empty')
if [[ -z "$STORE_ID" ]]; then
  echo "ERROR: 创建门店返回缺少 id: $STORE_CREATE" >&2
  exit 1
fi
echo "✓ 门店创建成功 STORE_ID=$STORE_ID" >&2

# 2) 上架商品（门店商品库存绑定）
echo "[2/6] 门店上架商品（绑定库存）..." >&2
UPSERT=$(curl -sS -X POST "$API_BASE/api/v1/admin/stores/$STORE_ID/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"product_id\": $PRODUCT_ID,
    \"stock\": 50,
    \"price_override\": \"\"
  }")
assert_ok_code "$UPSERT"
echo "✓ 商品已上架到门店" >&2

# 3) 用户下单（携带桌号）
echo "[3/6] 用户下单（加购→下单，携带桌号）..." >&2
USER_LOGIN=$(curl -sS -X POST "$API_BASE/api/v1/user/dev-login" \
  -H "Content-Type: application/json" \
  -d "{\"openid\":\"order_user_$(date +%s)\"}")
assert_ok_code "$USER_LOGIN"
USER_TOKEN=$(json_get "$USER_LOGIN" '.data.token // empty')

_=$(curl -sS -X POST "$API_BASE/api/v1/cart/items" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "{\"product_id\":$PRODUCT_ID,\"quantity\":1}")

ORDER_CREATE=$(curl -sS -X POST "$API_BASE/api/v1/orders/from-cart" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "{
    \"delivery_type\":1,
    \"store_id\":$STORE_ID,
    \"order_type\":2,
    \"table_id\":$TABLE_ID,
    \"table_no\":\"$TABLE_NO\",
    \"remark\":\"新流程自动测试\"
  }")
assert_ok_code "$ORDER_CREATE"
ORDER_ID=$(json_get "$ORDER_CREATE" '.data.id // empty')
if [[ -z "$ORDER_ID" ]]; then
  echo "ERROR: 创建订单返回缺少 id: $ORDER_CREATE" >&2
  exit 1
fi
echo "✓ 订单创建成功 ORDER_ID=$ORDER_ID" >&2

# 3.1) 门店设置桌号（订单维度：允许门店/管理员在必要时修正桌号）
echo "[3.1/6] 门店设置桌号（订单维度 set-table）..." >&2
SET_TABLE=$(curl -sS -X POST "$API_BASE/api/v1/orders/$ORDER_ID/set-table" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{\"table_id\":$TABLE_ID,\"table_no\":\"$TABLE_NO\",\"reason\":\"e2e\"}")
assert_ok_code "$SET_TABLE"
echo "✓ set-table 成功" >&2

DETAIL1=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" -H "Authorization: Bearer $USER_TOKEN")
assert_ok_code "$DETAIL1"
SAVED_TID=$(json_get "$DETAIL1" '.data.order.table_id // 0')
SAVED_TNO=$(json_get "$DETAIL1" '.data.order.table_no // ""')
if [[ "$SAVED_TID" != "$TABLE_ID" ]] || [[ "$SAVED_TNO" != "$TABLE_NO" ]]; then
  echo "ERROR: 桌号未正确保存，期望 $TABLE_ID/$TABLE_NO，实际 $SAVED_TID/$SAVED_TNO" >&2
  exit 1
fi

# 4) 付款（用户支付）
echo "[4/6] 用户付款（模拟支付）..." >&2
PAY=$(curl -sS -X POST "$API_BASE/api/v1/orders/$ORDER_ID/pay" \
  -H "Authorization: Bearer $USER_TOKEN")
assert_ok_code "$PAY"

DETAIL2=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" -H "Authorization: Bearer $USER_TOKEN")
assert_ok_code "$DETAIL2"
STATUS2=$(json_get "$DETAIL2" '.data.order.status // 0')
PAY_STATUS2=$(json_get "$DETAIL2" '.data.order.pay_status // 0')
if [[ "$STATUS2" != "2" ]] || [[ "$PAY_STATUS2" != "2" ]]; then
  echo "ERROR: 付款后状态不正确，status=$STATUS2 pay_status=$PAY_STATUS2" >&2
  exit 1
fi
echo "✓ 支付成功，订单已付款" >&2

# 5) 门店出货（deliver）
echo "[5/7] 门店出货（deliver）..." >&2
DELIVER=$(curl -sS -X POST "$API_BASE/api/v1/orders/$ORDER_ID/deliver" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
assert_ok_code "$DELIVER"

DETAIL3=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" -H "Authorization: Bearer $USER_TOKEN")
assert_ok_code "$DETAIL3"
STATUS3=$(json_get "$DETAIL3" '.data.order.status // 0')
if [[ "$STATUS3" != "3" ]]; then
  echo "ERROR: 出货后状态不正确，status=$STATUS3 (期望 3=配送中/出货中)" >&2
  exit 1
fi

# 6) 完成订单（complete）
echo "[6/7] 完成订单（complete）..." >&2
COMPLETE=$(curl -sS -X POST "$API_BASE/api/v1/orders/$ORDER_ID/complete" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
assert_ok_code "$COMPLETE"

DETAIL4=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" -H "Authorization: Bearer $USER_TOKEN")
assert_ok_code "$DETAIL4"
STATUS4=$(json_get "$DETAIL4" '.data.order.status // 0')
if [[ "$STATUS4" != "4" ]]; then
  echo "ERROR: 完成后状态不正确，status=$STATUS4 (期望 4=已完成)" >&2
  exit 1
fi

echo "" >&2
echo "========================================" >&2
echo "✓ 新流程测试通过" >&2
echo "  STORE_ID=$STORE_ID" >&2
echo "  ORDER_ID=$ORDER_ID" >&2
echo "  TABLE_ID=$TABLE_ID TABLE_NO=$TABLE_NO" >&2
echo "  status: $STATUS4" >&2
echo "========================================" >&2
