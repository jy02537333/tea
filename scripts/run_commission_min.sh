#!/usr/bin/env bash
set -euo pipefail

# 最小集成脚本：分销结算（下单→完成→解冻）
# - 依赖: curl, jq
# - 需要: 管理员 Token（env: ADMIN_TOKEN），若未提供将尝试从日志发现或本地登录(admin/pass)
# - 需要: 订单ID（env: ORDER_ID），若未提供将尝试调用 scripts/local_api_check.sh 创建并发现
# - 输出: 构建产物写入 build-ci-logs/

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT_DIR/build-ci-logs"
mkdir -p "$LOG_DIR"
SUMMARY_FILE="$LOG_DIR/local_api_summary.txt"
touch "$SUMMARY_FILE"

API_BASE=${API_BASE:-"http://127.0.0.1:9292"}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}
ORDER_ID=${ORDER_ID:-""}

have_jq=1
command -v jq >/dev/null 2>&1 || have_jq=0

pp_json() {
  if [[ $have_jq -eq 1 ]]; then jq .; else cat; fi
}

discover_admin_token() {
  if [[ -n "$ADMIN_TOKEN" ]]; then return 0; fi
  # 优先从 admin_login_response.json
  local f1="$LOG_DIR/admin_login_response.json"
  if [[ -z "$ADMIN_TOKEN" && -f "$f1" && $have_jq -eq 1 ]]; then
    ADMIN_TOKEN=$(jq -r '.data.token // .token // empty' "$f1" 2>/dev/null || echo "") || true
  fi
  # 次选从 stateful logs
  local f2="$ROOT_DIR/build-ci-logs/api_validation_stateful/POST__api_v1_auth_login.json"
  if [[ -z "$ADMIN_TOKEN" && -f "$f2" && $have_jq -eq 1 ]]; then
    ADMIN_TOKEN=$(jq -r '.data.token // .token // empty' "$f2" 2>/dev/null || echo "") || true
  fi
  # 回退：本地登录
  if [[ -z "$ADMIN_TOKEN" ]]; then
    local resp
    resp=$(curl -sS -X POST -H 'Content-Type: application/json' \
      -d '{"username":"admin","password":"pass"}' \
      "$API_BASE/api/v1/auth/login" || true)
    if [[ $have_jq -eq 1 ]]; then
      ADMIN_TOKEN=$(echo "$resp" | jq -r '.data.token // .token // empty' 2>/dev/null || echo "") || true
    fi
    if [[ -n "$ADMIN_TOKEN" ]]; then
      echo "$resp" > "$LOG_DIR/admin_login_response.json" || true
    fi
  fi
}

auth_header() {
  [[ -n "$ADMIN_TOKEN" ]] && echo "Authorization: Bearer $ADMIN_TOKEN" || true
}

fail() { echo "ERROR: $*" >&2; exit 1; }

echo "[commission] Discovering admin token..."
discover_admin_token
[[ -n "$ADMIN_TOKEN" ]] || fail "未获取到管理员 Token（请导出 ADMIN_TOKEN 或确保本地登录可用）"

discover_order_id() {
  # 1) env
  if [[ -n "$ORDER_ID" ]]; then return 0; fi
  # 2) 最近的 order_detail_*.json
  local latest
  latest=$(ls -t "$LOG_DIR"/order_detail_*.json 2>/dev/null | head -n1 || true)
  if [[ -z "$ORDER_ID" && -n "$latest" && -f "$latest" ]]; then
    if [[ $have_jq -eq 1 ]]; then
      ORDER_ID=$(jq -r '.data.id // .id // empty' "$latest" 2>/dev/null || echo "") || true
    fi
    if [[ -z "$ORDER_ID" ]]; then
      # 从文件名兜底解析
      ORDER_ID=$(basename "$latest" | sed -E 's/order_detail_([0-9]+)\.json/\1/' || true)
    fi
  fi
  # 3) from-cart 响应
  local from_cart="$ROOT_DIR/build-ci-logs/api_validation_stateful/POST__api_v1_orders_from-cart.json"
  if [[ -z "$ORDER_ID" && -f "$from_cart" && $have_jq -eq 1 ]]; then
    ORDER_ID=$(jq -r '.data.id // .id // empty' "$from_cart" 2>/dev/null || echo "") || true
  fi
}

# 若订单ID未知，尝试调用 stateful 脚本创建
discover_order_id
if [[ -z "$ORDER_ID" ]]; then
  echo "[commission] ORDER_ID 未提供，尝试运行 scripts/local_api_check.sh 创建最小订单..."
  bash "$ROOT_DIR/scripts/local_api_check.sh" || true
  discover_order_id
fi

[[ -n "$ORDER_ID" ]] || fail "未找到可用 ORDER_ID，请确认商品/购物车最小链路可走通"
echo "[commission] 使用 ORDER_ID=$ORDER_ID"

echo "[1] 完成订单（履约 complete）"
complete_resp=$(curl -sS -X POST -H "$(auth_header)" "$API_BASE/api/v1/orders/$ORDER_ID/complete" || true)
echo "$complete_resp" | pp_json > "$LOG_DIR/commission_order_${ORDER_ID}_complete_resp.json" || true

echo "[2] 触发佣金解冻（财务）"
release_resp=$(curl -sS -X POST -H "$(auth_header)" "$API_BASE/api/v1/admin/finance/commission/release" || true)
echo "$release_resp" | pp_json > "$LOG_DIR/commission_release_resp.json" || true

status_release=$(curl -s -o /dev/null -w '%{http_code}' -H "$(auth_header)" -X POST "$API_BASE/api/v1/admin/finance/commission/release" || true)
if [[ "$status_release" != "200" ]]; then
  fail "佣金解冻接口返回非200: $status_release"
fi

# 可选：财务摘要
summary_resp=$(curl -sS -H "$(auth_header)" "$API_BASE/api/v1/admin/finance/summary" || true)
echo "$summary_resp" | pp_json > "$LOG_DIR/finance_summary_after_release.json" || true

echo "commission_min: order_id=$ORDER_ID, release_status=$status_release" >> "$SUMMARY_FILE" || true
echo "Done. 输出已写入 $LOG_DIR/"
