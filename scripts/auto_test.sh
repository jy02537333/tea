#!/usr/bin/env bash
set -euo pipefail

# 一键自动测试（聚焦本次新流程能力）
# 覆盖：
# - tea-api handler 单测（mock pay / set-table）
# - 扫码点餐简化流程（下单带桌号 -> 模拟支付）
# - 新流程（门店注册 -> 上架商品 -> 下单 -> 支付 -> 出货 -> 完成）
#
# Usage:
#   bash scripts/auto_test.sh
#
# Env:
#   API_BASE    默认 http://127.0.0.1:9292
#   PRODUCT_ID  可选：指定可下单商品 ID
#   TABLE_ID    默认 1
#   TABLE_NO    默认 A12

API_BASE="${API_BASE:-http://127.0.0.1:9292}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: 缺少依赖命令: $1" >&2
    exit 1
  }
}

need_cmd curl
need_cmd jq
need_cmd go

health_check() {
  curl -fsS "$API_BASE/api/v1/health" >/dev/null 2>&1
}

echo "========================================" >&2
echo "一键自动测试" >&2
echo "========================================" >&2
echo "API_BASE=$API_BASE" >&2

if ! health_check; then
  echo "" >&2
  echo "ERROR: tea-api 未就绪：$API_BASE/api/v1/health 不可达" >&2
  echo "你可以先运行: make up" >&2
  exit 1
fi

echo "✓ tea-api health OK" >&2

echo "" >&2
echo "[1/3] 运行 Go handler 单测..." >&2
(cd tea-api && go test ./internal/handler -run 'Test(AdminSetTable_UpdatesOrder|OrderPay_)')

echo "" >&2
echo "[2/3] 运行桌号下单+模拟支付脚本..." >&2
API_BASE="$API_BASE" bash scripts/test_table_order_simple.sh

echo "" >&2
echo "[3/3] 运行门店新流程脚本..." >&2
API_BASE="$API_BASE" bash scripts/test_store_order_flow.sh

echo "" >&2
echo "========================================" >&2
echo "✓ 自动测试全部通过" >&2
echo "========================================" >&2
