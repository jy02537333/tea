#!/usr/bin/env bash
set -euo pipefail

# 为扫码点餐测试准备数据（创建商品+关联门店）
# 需要 admin 权限
# Usage: bash scripts/prepare_table_order_data.sh

API_BASE="${API_BASE:-http://127.0.0.1:9292}"
STORE_ID="${STORE_ID:-1}"

# 读取 admin token
ADMIN_TOKEN=$(cat build-ci-logs/admin_token.txt 2>/dev/null || echo "")
if [[ -z "$ADMIN_TOKEN" ]]; then
  echo "ERROR: 需要 admin token" >&2
  echo "请先登录管理后台生成 token 并保存到 build-ci-logs/admin_token.txt" >&2
  exit 1
fi

echo "准备扫码点餐测试数据..." >&2

# 1. 使用已有商品（假设 ID=1 存在）
echo "[1/3] 使用已有商品 ID=1..." >&2
PRODUCT_ID=1

# 2. 关联商品到门店
echo "[2/3] 关联商品到门店 $STORE_ID..." >&2
LINK_RESP=$(curl -sS -X POST "$API_BASE/api/v1/admin/stores/$STORE_ID/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"product_id\":$PRODUCT_ID,
    \"status\":1,
    \"stock\":100
  }")

echo "✓ 商品关联成功" >&2

# 3. 验证门店商品列表
echo "[3/3] 验证门店商品列表..." >&2
STORE_PRODUCTS=$(curl -sS "$API_BASE/api/v1/stores/$STORE_ID/products?status=1" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

AVAILABLE_COUNT=$(echo "$STORE_PRODUCTS" | jq '.data | length')
echo "✓ 门店当前有 $AVAILABLE_COUNT 个已上架商品" >&2

echo "" >&2
echo "数据准备完成！" >&2
echo "现在可以运行: bash scripts/test_table_order_simple.sh" >&2
