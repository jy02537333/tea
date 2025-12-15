#!/usr/bin/env bash
# Sprint A/B 回归测试脚本
# 测试关键接口：购物车、下单、可用券列表

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/build-ci-logs/sprint_ab_regression"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.txt"
: > "$SUMMARY"

BASE="${BASE:-http://127.0.0.1:9292}"
TOKEN_FILE="${TOKEN_FILE:-$ROOT/build-ci-logs/admin_login_response.json}"
TOKEN=""

echo "==================================================================" | tee -a "$SUMMARY"
echo "Sprint A/B 回归测试 - 关键接口稳定性验证" | tee -a "$SUMMARY"
echo "测试范围：购物车、下单、可用券列表" | tee -a "$SUMMARY"
echo "==================================================================" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

# 获取 token
if [[ -f "$TOKEN_FILE" ]]; then
  TOKEN=$(grep -Po '"token"\s*:\s*"\K[^"]+' "$TOKEN_FILE" || true)
fi

if [[ -z "$TOKEN" ]]; then
  echo "警告: 未找到有效的 token，某些接口可能无法测试" | tee -a "$SUMMARY"
  echo "请确保 $TOKEN_FILE 文件存在且包含有效的 token" | tee -a "$SUMMARY"
  echo "" | tee -a "$SUMMARY"
fi

# 测试请求函数
run_test() {
  local category="$1"
  local method="$2"
  local path="$3"
  local needs_auth="$4"
  local description="$5"
  
  local url="$BASE$path"
  local slug=$(echo "${category}_${method}_${path}" | sed -e 's|https\?://||' -e 's|[^a-zA-Z0-9._-]|_|g')
  local outfile="$OUT_DIR/${slug}.json"

  echo "[$category] $description" | tee -a "$SUMMARY"
  echo "  $method $url" | tee -a "$SUMMARY"

  local headers=(-H "Accept: application/json")
  if [[ "$needs_auth" == "yes" ]]; then
    if [[ -n "$TOKEN" ]]; then
      headers+=(-H "Authorization: Bearer $TOKEN")
    else
      echo "  ⚠️  需要认证但 token 缺失" | tee -a "$SUMMARY"
      echo "" | tee -a "$SUMMARY"
      return
    fi
  fi

  local response
  if ! response=$(curl -sS -X "$method" "${headers[@]}" --max-time 20 -w "\nHTTP_STATUS:%{http_code}" "$url" 2>&1); then
    echo "  ❌ 请求失败: $response" | tee -a "$SUMMARY"
    echo "" | tee -a "$SUMMARY"
    return
  fi

  local http_code
  http_code=$(printf '%s' "$response" | awk -F 'HTTP_STATUS:' 'NF>1 {print $NF}' | tail -n1 | tr -d '\r\n')
  local body
  body=$(printf '%s' "$response" | sed 's/HTTP_STATUS:.*//')
  printf '%s' "$body" > "$outfile"

  if [[ "$http_code" == "200" ]]; then
    echo "  ✅ 状态码: $http_code (成功)" | tee -a "$SUMMARY"
  elif [[ "$http_code" == "401" ]]; then
    echo "  ⚠️  状态码: $http_code (认证失败)" | tee -a "$SUMMARY"
  elif [[ "$http_code" == "404" ]]; then
    echo "  ⚠️  状态码: $http_code (资源不存在)" | tee -a "$SUMMARY"
  else
    echo "  ⚠️  状态码: $http_code" | tee -a "$SUMMARY"
  fi

  # 解析响应体
  if [[ -s "$outfile" ]]; then
    if python3 -c "import json; json.load(open('$outfile'))" 2>/dev/null; then
      local code=$(python3 -c "import json; obj=json.load(open('$outfile')); print(obj.get('code', 'N/A'))" 2>/dev/null || echo "N/A")
      local msg=$(python3 -c "import json; obj=json.load(open('$outfile')); print(obj.get('message', '')[:50])" 2>/dev/null || echo "")
      echo "  响应码: $code" | tee -a "$SUMMARY"
      if [[ -n "$msg" ]]; then
        echo "  消息: $msg" | tee -a "$SUMMARY"
      fi
      echo "  详情: $outfile" | tee -a "$SUMMARY"
    else
      echo "  响应体: 非 JSON 格式" | tee -a "$SUMMARY"
    fi
  fi
  
  echo "" | tee -a "$SUMMARY"
}

# Sprint A - 购物车相关接口
echo "=== Sprint A: 购物车 (Cart) ===" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

run_test "购物车" "GET" "/api/v1/cart" "yes" "获取用户购物车"
run_test "购物车" "GET" "/api/v1/cart/items" "yes" "获取购物车商品列表"

# Sprint A - 下单相关接口
echo "=== Sprint A: 下单 (Order) ===" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

run_test "下单" "GET" "/api/v1/orders" "yes" "获取用户订单列表"
run_test "下单" "GET" "/api/v1/orders/1" "yes" "获取订单详情 (ID=1)"
run_test "下单" "GET" "/api/v1/products" "no" "获取商品列表（下单前置）"
run_test "下单" "GET" "/api/v1/products/1" "no" "获取商品详情（下单前置）"
run_test "下单" "GET" "/api/v1/stores" "no" "获取门店列表（门店下单前置）"

# Sprint A - 支付相关接口
echo "=== Sprint A: 支付 (Payment) ===" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

run_test "支付" "GET" "/api/v1/user/info" "yes" "获取用户信息（支付前置）"

# Sprint B - 优惠券相关接口
echo "=== Sprint B: 优惠券 (Coupon) ===" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

run_test "优惠券" "GET" "/api/v1/coupons" "no" "获取可用优惠券模板列表"
run_test "优惠券" "GET" "/api/v1/user/coupons" "yes" "获取用户优惠券列表"
run_test "优惠券" "GET" "/api/v1/coupons/templates" "no" "获取优惠券模板"

# Sprint B - 用户与会员接口
echo "=== Sprint B: 用户与会员 (User & Membership) ===" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

run_test "用户" "GET" "/api/v1/user/me" "yes" "获取当前用户信息"
run_test "用户" "GET" "/api/v1/users/me/summary" "yes" "获取用户中心聚合数据"
run_test "会员" "GET" "/api/v1/wallet" "yes" "获取用户钱包信息"
run_test "积分" "GET" "/api/v1/points" "yes" "获取用户积分"

# 分类接口（购物车和下单依赖）
echo "=== 基础数据: 分类与门店 (Categories & Stores) ===" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"

run_test "分类" "GET" "/api/v1/categories" "no" "获取商品分类列表"
run_test "门店" "GET" "/api/v1/stores/1" "no" "获取门店详情 (ID=1)"

# 生成测试摘要统计
echo "==================================================================" | tee -a "$SUMMARY"
echo "测试摘要" | tee -a "$SUMMARY"
echo "==================================================================" | tee -a "$SUMMARY"

# 统计成功/失败数量
SUCCESS_COUNT=$(grep -c "✅" "$SUMMARY" || echo "0")
WARNING_COUNT=$(grep -c "⚠️" "$SUMMARY" || echo "0")
FAIL_COUNT=$(grep -c "❌" "$SUMMARY" || echo "0")

echo "" | tee -a "$SUMMARY"
echo "✅ 成功: $SUCCESS_COUNT" | tee -a "$SUMMARY"
echo "⚠️  警告: $WARNING_COUNT" | tee -a "$SUMMARY"
echo "❌ 失败: $FAIL_COUNT" | tee -a "$SUMMARY"
echo "" | tee -a "$SUMMARY"
echo "详细日志目录: $OUT_DIR" | tee -a "$SUMMARY"
echo "完整摘要文件: $SUMMARY" | tee -a "$SUMMARY"
echo "==================================================================" | tee -a "$SUMMARY"

# 输出完整摘要
echo ""
echo "回归测试完成！"
echo "查看完整报告: cat $SUMMARY"

# 返回退出码（如果有失败则返回非零）
if [[ $FAIL_COUNT -gt 0 ]]; then
  exit 1
fi

exit 0
