#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$ROOT/build-ci-logs/api_validation"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.txt"
: > "$SUMMARY"

BASE="${BASE:-http://127.0.0.1:9292}"
TOKEN_FILE="${TOKEN_FILE:-$ROOT/build-ci-logs/admin_login_response.json}"
TOKEN=""
if [[ -f "$TOKEN_FILE" ]]; then
  TOKEN=$(grep -Po '"token"\s*:\s*"\K[^"]+' "$TOKEN_FILE" || true)
fi

REQUESTS=(
  "GET|/api/v1/health|no"
  "GET|/api/v1/products|no"
  "GET|/api/v1/products/1|no"
  "GET|/api/v1/categories|no"
  "GET|/api/v1/stores|no"
  "GET|/api/v1/coupons|no"
  "GET|/api/v1/user/info|yes"
  "GET|/api/v1/user/coupons|yes"
  "GET|/api/v1/cart|yes"
  "GET|/api/v1/orders|yes"
  "GET|/api/v1/orders/1|yes"
  "GET|/api/v1/admin/orders|yes"
  "GET|/api/v1/admin/users|yes"
  "GET|/api/v1/admin/stores/1/orders/stats|yes"
  "GET|/api/v1/admin/stores/1/products|yes"
  "GET|/api/v1/admin/rbac/permissions|yes"
  "GET|/api/v1/admin/rbac/roles|yes"
  "GET|/api/v1/admin/logs/operations|yes"
)

slugify() {
  echo "$1" | sed -e 's|https\?://||' -e 's|[^a-zA-Z0-9._-]|_|g'
}

run_req() {
  local method="$1"
  local path="$2"
  local needs_auth="$3"
  local url="$BASE$path"
  local outfile="$OUT_DIR/$(slugify "${method}_${path}").json"

  {
    echo "=== $method $url (auth=$needs_auth) ==="
  } >> "$SUMMARY"

  local headers=(-H "Accept: application/json")
  if [[ "$needs_auth" == "yes" ]]; then
    if [[ -n "$TOKEN" ]]; then
      headers+=(-H "Authorization: Bearer $TOKEN")
    else
      echo "WARN: auth requested but token missing; sent without Authorization header." >> "$SUMMARY"
    fi
  fi

  local response
  if ! response=$(curl -sS -X "$method" "${headers[@]}" --max-time 20 -w "\nHTTP_STATUS:%{http_code}" "$url" 2>&1); then
    echo "Curl execution failed: $response" >> "$SUMMARY"
    echo >> "$SUMMARY"
    return
  fi

  local http_code
  http_code=$(printf '%s' "$response" | awk -F 'HTTP_STATUS:' 'NF>1 {print $NF}' | tail -n1 | tr -d '\r\n')
  local body
  body=$(printf '%s' "$response" | sed 's/HTTP_STATUS:.*//')
  printf '%s' "$body" > "$outfile"

  {
    echo "Status: $http_code"
    echo "Body saved to: $outfile"
  } >> "$SUMMARY"

  python3 - "$outfile" >> "$SUMMARY" 2>&1 <<'PY'
import sys
import json
import textwrap
from pathlib import Path

body_file = Path(sys.argv[1])
if not body_file.exists():
    print(f"Body file missing: {body_file}")
    print()
    sys.exit(0)

text = body_file.read_text(encoding='utf-8', errors='replace')
if not text.strip():
    print('Body empty')
    print()
    sys.exit(0)

try:
    obj = json.loads(text)
except Exception as exc:
    snippet = textwrap.shorten(text.replace('\n', ' '), width=400, placeholder='...')
    print(f'Body not JSON or parse error: {exc}')
    if snippet:
        print('Raw body snippet:', snippet)
    print()
    sys.exit(0)

if isinstance(obj, dict):
    keys = sorted(obj.keys())
    print('Top-level keys:', keys)
elif isinstance(obj, list):
    print('Top-level is array, length=', len(obj))
else:
    print('Top-level type:', type(obj).__name__)

pretty = json.dumps(obj, ensure_ascii=False, indent=2)
lines = pretty.splitlines()
max_lines = 20
print(f'Body preview (max {max_lines} lines):')
for line in lines[:max_lines]:
    print(line)
if len(lines) > max_lines:
    print('... (truncated) ...')
print()
PY

  if [[ "$http_code" == "404" ]]; then
    echo "Note: 404 indicates resource not found. 请确认 ID 是否存在或测试数据已初始化。" >> "$SUMMARY"
  fi

  if [[ "$http_code" == "401" ]]; then
    echo "Note: 401 indicates auth failure. 请确认 TOKEN_FILE 内的 token 是否有效。" >> "$SUMMARY"
  fi

  echo >> "$SUMMARY"
}

for req in "${REQUESTS[@]}"; do
  IFS='|' read -r method path auth <<< "$req"
  run_req "$method" "$path" "$auth"
done

cat "$SUMMARY"
echo
echo "Detailed bodies saved under $OUT_DIR"
