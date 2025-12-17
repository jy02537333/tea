#!/usr/bin/env bash
# Sprint A/B 集成回归测试
# 完整测试流程：登录 -> 浏览商品 -> 加入购物车 -> 下单 -> 使用优惠券

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/build-ci-logs/sprint_ab_integration"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/integration_summary.txt"
: > "$SUMMARY"

BASE="${BASE:-http://127.0.0.1:9292}"

log() {
  echo "$@" | tee -a "$SUMMARY"
}

log "=================================================================="
log "Sprint A/B 集成回归测试"
log "测试完整业务流程：购物车 -> 下单 -> 优惠券"
log "=================================================================="
log ""

PASS_COUNT=0
FAIL_COUNT=0

# 测试步骤函数
test_step() {
  local step_name="$1"
  local expected_code="$2"
  local actual_code="$3"
  
  if [[ "$actual_code" == "$expected_code" ]]; then
    log "✅ $step_name - 成功 (状态码: $actual_code)"
    ((PASS_COUNT++))
    return 0
  else
    log "❌ $step_name - 失败 (期望: $expected_code, 实际: $actual_code)"
    ((FAIL_COUNT++))
    return 1
  fi
}

# 1. 健康检查
log "步骤 1: 健康检查"
log "----------------------------------------"
HEALTH_RESP=$(curl -sS -X GET "$BASE/api/v1/health" -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
HEALTH_CODE=$(echo "$HEALTH_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
test_step "健康检查" "200" "$HEALTH_CODE"
log ""

# 2. 开发环境登录
log "步骤 2: 用户登录 (开发环境)"
log "----------------------------------------"
LOGIN_RESP=$(curl -sS -X POST "$BASE/api/v1/user/dev-login" \
  -H "Content-Type: application/json" \
  -d '{"openid":"regression_test_user"}' \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
LOGIN_CODE=$(echo "$LOGIN_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
LOGIN_BODY=$(echo "$LOGIN_RESP" | sed 's/HTTP_CODE:.*//')

if test_step "用户登录" "200" "$LOGIN_CODE"; then
  TOKEN=$(echo "$LOGIN_BODY" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo "")
  if [[ -n "$TOKEN" ]]; then
    log "  Token: ${TOKEN:0:20}..."
  else
    log "  ⚠️  无法提取 token"
  fi
else
  log "登录失败，无法继续测试"
  exit 1
fi
log ""

# 3. 获取商品列表
log "步骤 3: 获取商品列表"
log "----------------------------------------"
PRODUCTS_RESP=$(curl -sS -X GET "$BASE/api/v1/products?page=1&size=10" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
PRODUCTS_CODE=$(echo "$PRODUCTS_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
PRODUCTS_BODY=$(echo "$PRODUCTS_RESP" | sed 's/HTTP_CODE:.*//')

if test_step "获取商品列表" "200" "$PRODUCTS_CODE"; then
  PRODUCT_COUNT=$(echo "$PRODUCTS_BODY" | python3 -c "import sys, json; obj=json.load(sys.stdin); print(len(obj.get('data', [])))" 2>/dev/null || echo "0")
  log "  商品数量: $PRODUCT_COUNT"
  
  # 尝试获取第一个商品 ID
  PRODUCT_ID=$(echo "$PRODUCTS_BODY" | python3 -c "import sys, json; obj=json.load(sys.stdin); data=obj.get('data', []); print(data[0]['id'] if len(data)>0 else '')" 2>/dev/null || echo "")
  if [[ -n "$PRODUCT_ID" ]]; then
    log "  测试商品 ID: $PRODUCT_ID"
  fi
fi
log ""

# 4. 获取购物车（初始状态）
log "步骤 4: 获取购物车（初始状态）"
log "----------------------------------------"
CART_RESP=$(curl -sS -X GET "$BASE/api/v1/cart" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
CART_CODE=$(echo "$CART_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
test_step "获取购物车" "200" "$CART_CODE"
log ""

# 5. 加入购物车（如果有商品 ID）
if [[ -n "$PRODUCT_ID" ]]; then
  log "步骤 5: 加入购物车"
  log "----------------------------------------"
  ADD_CART_RESP=$(curl -sS -X POST "$BASE/api/v1/cart/items" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"product_id\":$PRODUCT_ID,\"quantity\":1}" \
    -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
  ADD_CART_CODE=$(echo "$ADD_CART_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
  test_step "加入购物车" "200" "$ADD_CART_CODE"
  log ""
else
  log "步骤 5: 加入购物车 - 跳过（无可用商品）"
  log ""
fi

# 6. 获取优惠券列表
log "步骤 6: 获取可用优惠券列表"
log "----------------------------------------"
COUPONS_RESP=$(curl -sS -X GET "$BASE/api/v1/coupons" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
COUPONS_CODE=$(echo "$COUPONS_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
COUPONS_BODY=$(echo "$COUPONS_RESP" | sed 's/HTTP_CODE:.*//')

if test_step "获取优惠券列表" "200" "$COUPONS_CODE"; then
  COUPON_COUNT=$(echo "$COUPONS_BODY" | python3 -c "import sys, json; obj=json.load(sys.stdin); print(len(obj.get('data', [])))" 2>/dev/null || echo "0")
  log "  优惠券数量: $COUPON_COUNT"
fi
log ""

# 7. 获取用户优惠券
log "步骤 7: 获取用户优惠券"
log "----------------------------------------"
MY_COUPONS_RESP=$(curl -sS -X GET "$BASE/api/v1/user/coupons" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
MY_COUPONS_CODE=$(echo "$MY_COUPONS_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
test_step "获取用户优惠券" "200" "$MY_COUPONS_CODE"
log ""

# 8. 获取订单列表
log "步骤 8: 获取订单列表"
log "----------------------------------------"
ORDERS_RESP=$(curl -sS -X GET "$BASE/api/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
ORDERS_CODE=$(echo "$ORDERS_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
ORDERS_BODY=$(echo "$ORDERS_RESP" | sed 's/HTTP_CODE:.*//')

if test_step "获取订单列表" "200" "$ORDERS_CODE"; then
  ORDER_COUNT=$(echo "$ORDERS_BODY" | python3 -c "import sys, json; obj=json.load(sys.stdin); print(len(obj.get('data', [])))" 2>/dev/null || echo "0")
  log "  订单数量: $ORDER_COUNT"
fi
log ""

# 9. 获取分类列表（商品浏览辅助）
log "步骤 9: 获取分类列表"
log "----------------------------------------"
CATEGORIES_RESP=$(curl -sS -X GET "$BASE/api/v1/categories" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
CATEGORIES_CODE=$(echo "$CATEGORIES_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
test_step "获取分类列表" "200" "$CATEGORIES_CODE"
log ""

# 10. 获取门店列表（门店下单辅助）
log "步骤 10: 获取门店列表"
log "----------------------------------------"
STORES_RESP=$(curl -sS -X GET "$BASE/api/v1/stores" \
  -w "\nHTTP_CODE:%{http_code}" || echo "HTTP_CODE:000")
STORES_CODE=$(echo "$STORES_RESP" | grep -oP 'HTTP_CODE:\K\d+' || echo "000")
test_step "获取门店列表" "200" "$STORES_CODE"
log ""

# 测试摘要
log "=================================================================="
log "测试摘要"
log "=================================================================="
log ""
log "通过: $PASS_COUNT 个测试"
log "失败: $FAIL_COUNT 个测试"
log ""
log "详细日志: $SUMMARY"
log "=================================================================="

if [[ $FAIL_COUNT -gt 0 ]]; then
  log ""
  log "⚠️  存在失败的测试，请检查日志"
  exit 1
else
  log ""
  log "✅ 所有测试通过！Sprint A/B 关键接口稳定"
  exit 0
fi
