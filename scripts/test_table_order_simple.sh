#!/usr/bin/env bash
set -euo pipefail

# 扫码点餐简化测试（使用门店现有商品 + 模拟支付）
# 前提：门店必须至少有一个已上架的商品
# Usage: bash scripts/test_table_order_simple.sh
# Env:
#   API_BASE   API base URL (default: http://127.0.0.1:9292)
#   STORE_ID   门店 ID (default: 1)
#   TABLE_ID   桌号 ID (default: 1)
#   TABLE_NO   桌号编号 (default: A12)

API_BASE="${API_BASE:-http://127.0.0.1:9292}"
STORE_ID="${STORE_ID:-1}"
TABLE_ID="${TABLE_ID:-1}"
TABLE_NO="${TABLE_NO:-A12}"
PRODUCT_ID="${PRODUCT_ID:-}"

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
echo "扫码点餐简化测试（模拟支付）" >&2
echo "========================================" >&2
echo "API_BASE=$API_BASE" >&2
echo "STORE_ID=$STORE_ID" >&2
echo "TABLE_ID=$TABLE_ID" >&2
echo "TABLE_NO=$TABLE_NO" >&2
echo "" >&2

# 1. 获取测试用户 token
echo "[1/5] 获取测试用户 token..." >&2
LOGIN_RESP=$(curl -sS -X POST "$API_BASE/api/v1/user/dev-login" \
  -H "Content-Type: application/json" \
  -d "{\"openid\":\"test_table_user_$(date +%s)\"}")

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // empty')
if [[ -z "$TOKEN" ]]; then
  echo "ERROR: 登录失败: $LOGIN_RESP" >&2
  exit 1
fi
echo "✓ Token 获取成功" >&2

# 1.1 获取 admin token（用于门店上架商品）
echo "[1.1/5] 获取 admin token..." >&2
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

# 1.2 确保门店存在（若默认 STORE_ID 不存在则创建一个）
echo "[1.2/5] 确保门店存在..." >&2
STORE_ID_NUM=$(echo "$STORE_ID" | tr -cd '0-9')
if [[ -n "$STORE_ID_NUM" && "$STORE_ID_NUM" != "0" ]]; then
  STORE_DETAIL=$(curl -sS "$API_BASE/api/v1/stores/$STORE_ID_NUM" || true)
  STORE_DETAIL_CODE=$(echo "$STORE_DETAIL" | jq -r '.code // -1' 2>/dev/null || echo -1)
  STORE_DETAIL_ID=$(echo "$STORE_DETAIL" | jq -r '.data.id // empty' 2>/dev/null || echo "")
  STORE_DETAIL_STATUS=$(echo "$STORE_DETAIL" | jq -r '.data.status // 1' 2>/dev/null || echo 1)
  if [[ "$STORE_DETAIL_CODE" == "0" && -n "$STORE_DETAIL_ID" && "$STORE_DETAIL_STATUS" != "0" ]]; then
    STORE_ID="$STORE_ID_NUM"
  else
    STORE_ID=""
  fi
else
  STORE_ID=""
fi

if [[ -z "$STORE_ID" ]]; then
  STORE_PAYLOAD=$(jq -n --arg name "扫码点餐测试门店_$(date +%s)" '{name:$name,address:"测试地址",phone:"13800000000",status:1}')
  STORE_CREATE=$(curl -sS -X POST "$API_BASE/api/v1/stores" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$STORE_PAYLOAD")
  assert_ok_code "$STORE_CREATE"
  STORE_ID=$(json_get "$STORE_CREATE" '.data.id // empty')
  if [[ -z "$STORE_ID" ]]; then
    echo "ERROR: 创建门店失败: $STORE_CREATE" >&2
    exit 1
  fi
fi
echo "✓ 使用门店 STORE_ID=$STORE_ID" >&2

# 2. 选择一个可下单商品（优先使用 env，其次从 CI 日志推断，再退化到 163）
echo "[2/5] 选择可下单商品..." >&2

if [[ -z "$PRODUCT_ID" ]] && [[ -f build-ci-logs/upsert_store1_product_resp.json ]]; then
  PRODUCT_ID=$(jq -r '.data.product_id // empty' build-ci-logs/upsert_store1_product_resp.json 2>/dev/null || true)
fi

if [[ -z "$PRODUCT_ID" ]]; then
  PRODUCT_ID=163
fi

echo "✓ 使用商品 ID: $PRODUCT_ID" >&2

# 2.1 确保商品已上架到门店（绑定库存）
echo "[2.1/5] 确保商品已上架到门店..." >&2
UPSERT_PAYLOAD=$(jq -n --argjson product_id "$PRODUCT_ID" --arg stock "50" '{product_id:$product_id,stock:($stock|tonumber),price_override:""}')
UPSERT=$(curl -sS -X POST "$API_BASE/api/v1/admin/stores/$STORE_ID/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "$UPSERT_PAYLOAD")
assert_ok_code "$UPSERT"
echo "✓ 商品已上架到门店" >&2

# 3. 添加到购物车
echo "[3/5] 添加商品到购物车..." >&2
CART_RESP=$(curl -sS -X POST "$API_BASE/api/v1/cart/items" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"product_id\":$PRODUCT_ID,
    \"store_id\":$STORE_ID,
    \"quantity\":2
  }")

echo "✓ 购物车添加成功" >&2

# 4. 创建订单（带桌号信息）
echo "[4/5] 创建订单（带桌号 table_id=$TABLE_ID, table_no=$TABLE_NO）..." >&2
ORDER_RESP=$(curl -sS -X POST "$API_BASE/api/v1/orders/from-cart" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"delivery_type\":1,
    \"store_id\":$STORE_ID,
    \"order_type\":2,
    \"table_id\":$TABLE_ID,
    \"table_no\":\"$TABLE_NO\",
    \"remark\":\"扫码点餐测试订单\"
  }")

ORDER_ID=$(echo "$ORDER_RESP" | jq -r '.data.id // .data.order.id // empty')
if [[ -z "$ORDER_ID" ]]; then
  echo "ERROR: 订单创建失败: $ORDER_RESP" >&2
  exit 1
fi
echo "✓ 订单创建成功 ID: $ORDER_ID" >&2

# 验证订单中的桌号信息
ORDER_DETAIL=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN")
SAVED_TABLE_ID=$(echo "$ORDER_DETAIL" | jq -r '.data.order.table_id // 0')
SAVED_TABLE_NO=$(echo "$ORDER_DETAIL" | jq -r '.data.order.table_no // ""')

echo "  订单中保存的桌号信息:" >&2
echo "    table_id: $SAVED_TABLE_ID" >&2
echo "    table_no: $SAVED_TABLE_NO" >&2

# 5. 模拟支付
echo "[5/5] 模拟支付订单 $ORDER_ID..." >&2
PAY_RESP=$(curl -sS -X POST "$API_BASE/api/v1/orders/$ORDER_ID/pay" \
  -H "Authorization: Bearer $TOKEN")

PAY_OK=$(echo "$PAY_RESP" | jq -r '.data.ok // false')
if [[ "$PAY_OK" != "true" ]]; then
  echo "ERROR: 支付失败: $PAY_RESP" >&2
  exit 1
fi
echo "✓ 支付成功" >&2

# 最终验证
FINAL_ORDER=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN")

ORDER_STATUS=$(echo "$FINAL_ORDER" | jq -r '.data.order.status // 0')
PAY_STATUS=$(echo "$FINAL_ORDER" | jq -r '.data.order.pay_status // 0')

echo "" >&2
echo "========================================" >&2
echo "最终订单状态:" >&2
echo "  订单 ID: $ORDER_ID" >&2
echo "  订单状态: $ORDER_STATUS (2=已付款)" >&2
echo "  支付状态: $PAY_STATUS (2=已支付)" >&2
echo "  门店 ID: $(echo "$FINAL_ORDER" | jq -r '.data.order.store_id // 0')" >&2
echo "  桌号 ID: $(echo "$FINAL_ORDER" | jq -r '.data.order.table_id // 0')" >&2
echo "  桌号编号: $(echo "$FINAL_ORDER" | jq -r '.data.order.table_no // ""')" >&2
echo "========================================" >&2

if [[ "$ORDER_STATUS" == "2" ]] && [[ "$PAY_STATUS" == "2" ]]; then
  echo "" >&2
  echo "✓ 扫码点餐全流程测试通过！" >&2
  exit 0
else
  echo "" >&2
  echo "ERROR: 订单状态异常" >&2
  echo "期望: order_status=2, pay_status=2" >&2
  echo "实际: order_status=$ORDER_STATUS, pay_status=$PAY_STATUS" >&2
  exit 1
fi
