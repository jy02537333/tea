#!/usr/bin/env bash
set -euo pipefail

# Backend connectivity summary generator
# Requires env: API_BASE, ADMIN_TOKEN
# Outputs: build-ci-logs/backend_connectivity.md

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LOG_DIR="$REPO_ROOT/build-ci-logs"
OUT_FILE="$LOG_DIR/backend_connectivity.md"
mkdir -p "$LOG_DIR"

API_BASE="${API_BASE:-}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

if [ -z "$API_BASE" ]; then
  echo "API_BASE is empty; skip connectivity summary" >&2
  exit 0
fi

authHeader=("-H" "Authorization: Bearer $ADMIN_TOKEN")
if [ -z "$ADMIN_TOKEN" ]; then
  # Allow unauthenticated health check only
  authHeader=()
fi

function http_json() {
  local url="$1";
  curl -sS -H "Content-Type: application/json" "${authHeader[@]}" "$url" || true
}

function get_code() { echo "$1" | jq -r '.code // empty' 2>/dev/null || true; }
function get_total() { echo "$1" | jq -r '.total // empty' 2>/dev/null || true; }
function get_status_200() { curl -sS -o /dev/null -w "%{http_code}" "$1" || true; }

health_url="$API_BASE/api/v1/health"
stores_url="$API_BASE/api/v1/stores?page=1&limit=2"

health_status=$(get_status_200 "$health_url")
stores_json=$(http_json "$stores_url")
store_id=$(echo "$stores_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

finance_json=""; finance_code=""; finance_total=""; export_status=""
if [ -n "${store_id:-}" ]; then
  finance_url="$API_BASE/api/v1/stores/$store_id/finance/transactions?page=1&limit=5"
  export_url="$API_BASE/api/v1/stores/$store_id/finance/transactions/export?type=payment"
  finance_json=$(http_json "$finance_url")
  finance_code=$(get_code "$finance_json")
  finance_total=$(get_total "$finance_json")
  export_status=$(get_status_200 "$export_url")
fi

cat > "$OUT_FILE" <<MD
### 后端连通性摘要

- API_BASE: $API_BASE
- 健康检查: GET /api/v1/health → HTTP $health_status
- 门店列表: GET /api/v1/stores?page=1&limit=2 → code=$(get_code "$stores_json")
- 财务流水: GET /api/v1/stores/$store_id/finance/transactions?page=1&limit=5 → code=${finance_code:-} total=${finance_total:-}
- 财务导出: GET /api/v1/stores/$store_id/finance/transactions/export?type=payment → HTTP ${export_status:-}

> 注：如缺少 ADMIN_TOKEN，仅进行健康检查；如 stores 为空，跳过后续财务接口检查。
MD

echo "Wrote connectivity summary to $OUT_FILE"
