#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOGDIR="$ROOT/build-ci-logs"
BASE_URL="${BASE_URL:-}"

mkdir -p "$LOGDIR"

trim() { sed -e 's/\r$//' -e 's/[[:space:]]\+$//' ; }

detect_base_url() {
  if [[ -n "${BASE_URL:-}" ]]; then
    return
  fi

  local candidates=("http://127.0.0.1:9292" "http://127.0.0.1:9393")
  local c http
  for c in "${candidates[@]}"; do
    http="$(curl -sS -o /dev/null -w '%{http_code}' "$c/api/v1/health" || true)"
    if [[ "$http" == "200" ]]; then
      BASE_URL="$c"
      return
    fi
  done

  # Default fallback (keeps previous behavior) even if health is not reachable.
  BASE_URL="http://127.0.0.1:9292"
}

bearer() {
  local token="$1"
  token="$(printf '%s' "$token" | trim)"
  if [[ -z "$token" ]]; then
    echo ""
    return
  fi
  if [[ "$token" == Bearer\ * ]]; then
    echo "$token"
  else
    echo "Bearer $token"
  fi
}

# Tokens
ADMIN_TOKEN="${ADMIN_TOKEN:-}"
USER_TOKEN="${USER_TOKEN:-}"

ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin@123}"

# Prefer explicit env, then known evidence files.
admin_login() {
  detect_base_url

  cat > "$LOGDIR/invite_admin_login_latest_payload.json" <<JSON
{"username":"${ADMIN_USERNAME}","password":"${ADMIN_PASSWORD}"}
JSON

  local tmp_headers tmp_body http
  tmp_headers="$(mktemp)"
  tmp_body="$(mktemp)"
  http="$(curl -sS \
    -X POST "$BASE_URL/api/v1/user/login" \
    -D "$tmp_headers" \
    -o "$tmp_body" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    --data "@${LOGDIR}/invite_admin_login_latest_payload.json" \
    -w '%{http_code}' \
    || true)"
  mv "$tmp_headers" "$LOGDIR/invite_admin_login_latest_headers.txt"
  mv "$tmp_body" "$LOGDIR/invite_admin_login_latest_body.json"
  if [[ "$http" != "200" ]]; then
    echo "ERROR: admin login failed: HTTP $http" >&2
    echo "  url: $BASE_URL/api/v1/user/login" >&2
    echo "  headers: $LOGDIR/invite_admin_login_latest_headers.txt" >&2
    echo "  body:    $LOGDIR/invite_admin_login_latest_body.json" >&2
    exit 1
  fi

  ADMIN_TOKEN="$(python3 - <<PY
import json
p=r'${LOGDIR}/invite_admin_login_latest_body.json'
try:
  d=json.load(open(p,'r',encoding='utf-8'))
  print((d.get('data') or {}).get('token',''))
except Exception:
  print('')
PY
  )"

  if [[ -z "$(printf '%s' "$ADMIN_TOKEN" | trim)" ]]; then
    echo "ERROR: admin login succeeded but token missing in response" >&2
    echo "  body: $LOGDIR/invite_admin_login_latest_body.json" >&2
    exit 1
  fi
}

if [[ -z "$ADMIN_TOKEN" ]]; then
  # Prefer existing evidence token first (fast path), but we will refresh on 401.
  if [[ -f "$LOGDIR/invite_admin_login_latest_body.json" ]]; then
    ADMIN_TOKEN="$(python3 - <<PY
import json
p=r'${LOGDIR}/invite_admin_login_latest_body.json'
try:
  d=json.load(open(p,'r',encoding='utf-8'))
  print((d.get('data') or {}).get('token',''))
except Exception:
  print('')
PY
    )"
  fi
fi

if [[ -z "$USER_TOKEN" && -f "$LOGDIR/cart_dev_login_latest_token.txt" ]]; then
  USER_TOKEN="$(cat "$LOGDIR/cart_dev_login_latest_token.txt" | trim)"
fi

if [[ -z "$ADMIN_TOKEN" ]]; then
  admin_login
fi
if [[ -z "$USER_TOKEN" ]]; then
  echo "ERROR: missing USER_TOKEN (set env USER_TOKEN=... or ensure build-ci-logs/cart_dev_login_latest_token.txt exists)" >&2
  exit 1
fi

# Poster image URLs (use your OSS/CDN URLs if you want the share page to render real images).
POSTER_IMAGE_URL_1="${POSTER_IMAGE_URL_1:-https://example.com/share-poster-1.png}"
POSTER_IMAGE_URL_2="${POSTER_IMAGE_URL_2:-https://example.com/share-poster-2.png}"

# Referrer id for wxacode scene
REFERRER_ID="${REFERRER_ID:-}"
if [[ -z "$REFERRER_ID" && -f "$LOGDIR/invite_referrer_create_latest_body.json" ]]; then
  REFERRER_ID="$(python3 - <<'PY'
import json
p='build-ci-logs/invite_referrer_create_latest_body.json'
try:
  d=json.load(open(p,'r',encoding='utf-8'))
  print(d.get('data',{}).get('id',''))
except Exception:
  print('')
PY
)"
fi
if [[ -z "$REFERRER_ID" ]]; then
  # Fallback: use bound referrer_id evidence payload if present
  if [[ -f "$LOGDIR/invite_referral_bind_latest_payload.json" ]]; then
    REFERRER_ID="$(python3 - <<'PY'
import json
p='build-ci-logs/invite_referral_bind_latest_payload.json'
try:
  d=json.load(open(p,'r',encoding='utf-8'))
  print(d.get('referrer_id',''))
except Exception:
  print('')
PY
)"
  fi
fi
if [[ -z "$REFERRER_ID" ]]; then
  echo "ERROR: missing REFERRER_ID (set env REFERRER_ID=... or ensure invite_referrer_create_latest_body.json exists)" >&2
  exit 1
fi

curl_save() {
  local method="$1"; shift
  local url_path="$1"; shift
  local payload_file="$1"; shift
  local headers_out="$1"; shift
  local body_out="$1"; shift
  local auth_value="$1"; shift

  local tmp_headers tmp_body http
  tmp_headers="$(mktemp)"
  tmp_body="$(mktemp)"

  local args=("-sS" "-X" "$method" "$BASE_URL$url_path" "-D" "$tmp_headers" "-o" "$tmp_body" "-H" "Accept: application/json")
  if [[ -n "$auth_value" ]]; then
    args+=("-H" "Authorization: $auth_value")
  fi
  if [[ -n "$payload_file" ]]; then
    args+=("-H" "Content-Type: application/json" "--data" "@$payload_file")
  fi

  # Capture HTTP status code without printing response.
  http="$(curl "${args[@]}" -w '%{http_code}' || true)"

  mv "$tmp_headers" "$headers_out"
  mv "$tmp_body" "$body_out"

  printf '%s' "$http"
}

ADMIN_AUTH="$(bearer "$ADMIN_TOKEN")"
USER_AUTH="$(bearer "$USER_TOKEN")"

detect_base_url

# 0) Health check (best-effort)
# We do not fail here; it is only for debugging.
(
  set +e
  curl -sS "$BASE_URL/api/v1/health" > "$LOGDIR/invite_health_latest_body.json" 2>/dev/null
) || true

# 1) Admin: get current templates (refresh token on 401)
http="$(curl_save \
  "GET" \
  "/api/v1/admin/system/share-posters" \
  "" \
  "$LOGDIR/invite_admin_share_posters_get_latest_headers.txt" \
  "$LOGDIR/invite_admin_share_posters_get_latest_body.json" \
  "$ADMIN_AUTH" \
)"
if [[ "$http" == "401" ]]; then
  admin_login
  ADMIN_AUTH="$(bearer "$ADMIN_TOKEN")"
  http="$(curl_save \
    "GET" \
    "/api/v1/admin/system/share-posters" \
    "" \
    "$LOGDIR/invite_admin_share_posters_get_latest_headers.txt" \
    "$LOGDIR/invite_admin_share_posters_get_latest_body.json" \
    "$ADMIN_AUTH" \
  )"
fi
if [[ "$http" != "200" ]]; then
  echo "ERROR: GET /api/v1/admin/system/share-posters failed: HTTP $http" >&2
  echo "  url: $BASE_URL/api/v1/admin/system/share-posters" >&2
  echo "  headers: $LOGDIR/invite_admin_share_posters_get_latest_headers.txt" >&2
  echo "  body:    $LOGDIR/invite_admin_share_posters_get_latest_body.json" >&2
  exit 1
fi

# 2) Admin: put templates (replace)
cat > "$LOGDIR/invite_admin_share_posters_put_latest_payload.json" <<JSON
{
  "list": [
    {
      "id": "tpl-1",
      "title": "默认模板1",
      "image_url": "${POSTER_IMAGE_URL_1}",
      "sort": 1,
      "status": 1
    },
    {
      "id": "tpl-2",
      "title": "默认模板2",
      "image_url": "${POSTER_IMAGE_URL_2}",
      "sort": 2,
      "status": 1
    }
  ]
}
JSON

http="$(curl_save \
  "PUT" \
  "/api/v1/admin/system/share-posters" \
  "$LOGDIR/invite_admin_share_posters_put_latest_payload.json" \
  "$LOGDIR/invite_admin_share_posters_put_latest_headers.txt" \
  "$LOGDIR/invite_admin_share_posters_put_latest_body.json" \
  "$ADMIN_AUTH" \
)"
if [[ "$http" != "200" ]]; then
  echo "ERROR: PUT /api/v1/admin/system/share-posters failed: HTTP $http" >&2
  echo "  url: $BASE_URL/api/v1/admin/system/share-posters" >&2
  echo "  headers: $LOGDIR/invite_admin_share_posters_put_latest_headers.txt" >&2
  echo "  body:    $LOGDIR/invite_admin_share_posters_put_latest_body.json" >&2
  exit 1
fi

# 3) Public: posters list (should now be non-empty)
http="$(curl_save \
  "GET" \
  "/api/v1/share/posters" \
  "" \
  "$LOGDIR/invite_share_posters_latest_headers.txt" \
  "$LOGDIR/invite_share_posters_latest_body.json" \
  "" \
)"
if [[ "$http" != "200" ]]; then
  echo "ERROR: GET /api/v1/share/posters failed: HTTP $http" >&2
  echo "  url: $BASE_URL/api/v1/share/posters" >&2
  echo "  headers: $LOGDIR/invite_share_posters_latest_headers.txt" >&2
  echo "  body:    $LOGDIR/invite_share_posters_latest_body.json" >&2
  exit 1
fi

# 4) User: wxacode
cat > "$LOGDIR/invite_wxacode_latest_payload.json" <<JSON
{
  "scene": "referrer_id=${REFERRER_ID}",
  "page": "pages/index/index",
  "width": 240,
  "is_hyaline": true
}
JSON

http="$(curl_save \
  "POST" \
  "/api/v1/wx/wxacode" \
  "$LOGDIR/invite_wxacode_latest_payload.json" \
  "$LOGDIR/invite_wxacode_latest_headers.txt" \
  "$LOGDIR/invite_wxacode_latest_body.json" \
  "$USER_AUTH" \
)"
if [[ "$http" != "200" ]]; then
  # Local/dev environments may not have WeChat MiniProgram credentials.
  if [[ "$http" == "400" ]] && grep -q '"code":4002' "$LOGDIR/invite_wxacode_latest_body.json"; then
    echo "WARN: wxacode not generated because WECHAT_MINI_APPID/WECHAT_MINI_SECRET are not configured (evidence still saved)" >&2
  else
    echo "ERROR: POST /api/v1/wx/wxacode failed: HTTP $http" >&2
    echo "  url: $BASE_URL/api/v1/wx/wxacode" >&2
    echo "  headers: $LOGDIR/invite_wxacode_latest_headers.txt" >&2
    echo "  body:    $LOGDIR/invite_wxacode_latest_body.json" >&2
    echo "  hint: if this is 401, refresh USER_TOKEN by re-running your user login flow (cart_dev_login_latest_*)" >&2
    exit 1
  fi
fi

echo "OK: invite share evidence updated in $LOGDIR"
