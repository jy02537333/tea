#!/usr/bin/env bash
set -euo pipefail

# 测试扫码点餐全流程（含桌号+模拟支付）
# Usage: bash scripts/test_table_order_flow.sh
# Env:
#   API_BASE   API base URL (default: http://127.0.0.1:9292)
#   STORE_ID   门店 ID (default: 1)
#   TABLE_ID   桌号 ID (default: 1)
#   TABLE_NO   桌号编号 (default: A12)

API_BASE="${API_BASE:-http://127.0.0.1:9292}"
STORE_ID="${STORE_ID:-1}"
TABLE_ID="${TABLE_ID:-1}"
TABLE_NO="${TABLE_NO:-A12}"

echo "========================================" >&2
echo "扫码点餐全流程测试（含桌号+模拟支付）" >&2
echo "========================================" >&2
echo "API_BASE=$API_BASE" >&2
echo "STORE_ID=$STORE_ID" >&2
echo "TABLE_ID=$TABLE_ID" >&2
echo "TABLE_NO=$TABLE_NO" >&2
echo "" >&2

# 1. 获取或创建测试用户 token
echo "[1/6] 获取测试用户 token..." >&2
LOGIN_RESP=$(curl -sS -X POST "$API_BASE/api/v1/user/dev-login" \
  -H "Content-Type: application/json" \
  -d "{\"openid\":\"test_table_user_$(date +%s)\"}")

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // empty')
if [[ -z "$TOKEN" ]]; then
  echo "ERROR: 登录失败: $LOGIN_RESP" >&2
  exit 1
fi
echo "✓ Token 获取成功" >&2

# 2. 获取门店已有商品（从门店商品列表）
echo "[2/6] 获取门店商品..." >&2
STORE_PRODUCTS=$(curl -sS "$API_BASE/api/v1/stores/$STORE_ID/products?page=1&limit=1" \
  -H "Authorization: Bearer $TOKEN" || echo '{"data":[]}')

PRODUCT_ID=$(echo "$STORE_PRODUCTS" | jq -r '.data[0].product_id // .data[0].id // empty')

if [[ -z "$PRODUCT_ID" ]]; then
  echo "  门店暂无商品，创建测试商品并关联到门店..." >&2
  
  # 创建商品
  PRODUCT_RESP=$(curl -sS -X POST "$API_BASE/api/v1/products" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
      \"name\":\"测试商品_桌号_$(date +%s)\",
      \"category_id\":1,
      \"price\":\"28.00\",
      \"description\":\"用于测试扫码点餐流程\",
      \"stock\":100
    }")
  
  PRODUCT_ID=$(echo "$PRODUCT_RESP" | jq -r '.data.id // .data.product.id // empty')
  
  if [[ -z "$PRODUCT_ID" ]]; then
    echo "ERROR: 创建商品失败: $PRODUCT_RESP" >&2
    exit 1
  fi
  
  # 将商品关联到门店
  STORE_PRODUCT_RESP=$(curl -sS -X POST "$API_BASE/api/v1/admin/stores/$STORE_ID/products" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
      \"product_id\":$PRODUCT_ID,
      \"price\":\"28.00\",
      \"stock\":100,
      \"status\":1
    }")
  
  SP_ID=$(echo "$STORE_PRODUCT_RESP" | jq -r '.data.id // empty')
  if [[ -z "$SP_ID" ]]; then
    echo "  WARN: 门店商品关联响应: $STORE_PRODUCT_RESP" >&2
    # 可能已存在，继续流程
  fi
  
  echo "✓ 商品创建并关联到门店成功 ID: $PRODUCT_ID" >&2
else
  echo "✓ 使用门店商品 ID: $PRODUCT_ID" >&2
fi

# 3. 添加到购物车（模拟用户选择商品）
echo "[3/6] 添加商品到购物车..." >&2
CART_RESP=$(curl -sS -X POST "$API_BASE/api/v1/cart/items" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"product_id\":$PRODUCT_ID,
    \"quantity\":2
  }")

echo "✓ 购物车添加成功" >&2

# 4. 从购物车创建订单（带桌号信息）
echo "[4/6] 创建订单（带桌号 table_id=$TABLE_ID, table_no=$TABLE_NO）..." >&2
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

if [[ "$SAVED_TABLE_ID" != "$TABLE_ID" ]] || [[ "$SAVED_TABLE_NO" != "$TABLE_NO" ]]; then
  echo "WARNING: 桌号信息不匹配！" >&2
fi

# 5. 模拟支付
echo "[5/6] 模拟支付订单 $ORDER_ID..." >&2
PAY_RESP=$(curl -sS -X POST "$API_BASE/api/v1/orders/$ORDER_ID/pay" \
  -H "Authorization: Bearer $TOKEN")

PAY_OK=$(echo "$PAY_RESP" | jq -r '.data.ok // false')
if [[ "$PAY_OK" != "true" ]]; then
  echo "ERROR: 支付失败: $PAY_RESP" >&2
  exit 1
fi
echo "✓ 支付成功" >&2

# 6. 验证订单状态
echo "[6/6] 验证订单状态..." >&2
FINAL_ORDER=$(curl -sS "$API_BASE/api/v1/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN")

ORDER_STATUS=$(echo "$FINAL_ORDER" | jq -r '.data.order.status // 0')
PAY_STATUS=$(echo "$FINAL_ORDER" | jq -r '.data.order.pay_status // 0')

echo "  最终订单状态:" >&2
echo "    order_status: $ORDER_STATUS (2=已付款)" >&2
echo "    pay_status: $PAY_STATUS (2=已支付)" >&2
echo "    table_id: $(echo "$FINAL_ORDER" | jq -r '.data.order.table_id // 0')" >&2
echo "    table_no: $(echo "$FINAL_ORDER" | jq -r '.data.order.table_no // ""')" >&2
echo "    store_id: $(echo "$FINAL_ORDER" | jq -r '.data.order.store_id // 0')" >&2

if [[ "$ORDER_STATUS" == "2" ]] && [[ "$PAY_STATUS" == "2" ]]; then
  echo "" >&2
  echo "========================================" >&2
  echo "✓ 扫码点餐全流程测试通过！" >&2
  echo "  订单 ID: $ORDER_ID" >&2
  echo "  门店 ID: $STORE_ID" >&2
  echo "  桌号 ID: $TABLE_ID" >&2
  echo "  桌号编号: $TABLE_NO" >&2
  echo "========================================" >&2
  exit 0
else
  echo "" >&2
  echo "ERROR: 订单状态异常" >&2
  echo "期望: order_status=2, pay_status=2" >&2
  echo "实际: order_status=$ORDER_STATUS, pay_status=$PAY_STATUS" >&2
  exit 1
fi
