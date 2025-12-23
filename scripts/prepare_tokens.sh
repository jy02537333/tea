#!/usr/bin/env bash
set -euo pipefail

# Prepare ADMIN_TOKEN and USER_TOKEN for local/CI usage.
# Priority per token: existing file -> env token -> credentials login -> devlogin_resp.json -> empty (warn)

API_BASE="${API_BASE:-http://127.0.0.1:9292}"
OUT_DIR="build-ci-logs"
mkdir -p "$OUT_DIR"

need_bin() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[prepare_tokens] Missing dependency: $1" >&2
    exit 1
  fi
}

need_bin curl
need_bin jq

extract_token() {
  # Read JSON from stdin and extract token-like fields
  jq -r '(.token // .data.token // .access_token // .data.access_token // empty)'
}

save_json_if_not_empty() {
  local json="$1" path="$2"
  if [[ -n "$json" ]]; then
    printf '%s' "$json" > "$path"
  fi
}

login_with_phone_code() {
  local role="$1" phone="$2" code="$3"
  local url="$API_BASE/api/v1/auth/login"
  local payload
  payload=$(jq -n --arg phone "$phone" --arg code "$code" '{phone:$phone, code:$code}')
  local resp
  resp=$(curl -sS -X POST "$url" -H 'Content-Type: application/json' -d "$payload") || return 1
  save_json_if_not_empty "$resp" "$OUT_DIR/${role}_login_response.json"
  echo "$resp" | extract_token
}

login_with_username_password() {
  local role="$1" username="$2" password="$3"
  local url="$API_BASE/api/v1/auth/login"
  local payload
  payload=$(jq -n --arg username "$username" --arg password "$password" '{username:$username, password:$password}')
  local resp
  resp=$(curl -sS -X POST "$url" -H 'Content-Type: application/json' -d "$payload") || return 1
  save_json_if_not_empty "$resp" "$OUT_DIR/${role}_login_response.json"
  echo "$resp" | extract_token
}

ensure_token() {
  # args: role (admin|user)
  local role="$1"; shift || true
  local token_var role_ucase json_file env_token phone code username password
  role_ucase=$(echo "$role" | tr '[:lower:]' '[:upper:]')
  json_file="$OUT_DIR/${role}_login_response.json"
  env_token_var="${role_ucase}_TOKEN"
  phone_var="${role_ucase}_PHONE"
  code_var="${role_ucase}_CODE"
  username_var="${role_ucase}_USERNAME"
  password_var="${role_ucase}_PASSWORD"

  # 1) existing file
  if [[ -s "$json_file" ]]; then
    local t
    t=$(jq -r '(.token // .data.token // empty)' "$json_file" 2>/dev/null || true)
    if [[ -n "$t" && "$t" != "null" ]]; then
      echo "$t"
      return 0
    fi
  fi

  # 2) env token
  env_token="${!env_token_var-}"
  if [[ -n "${env_token}" ]]; then
    printf '{"token":"%s"}' "$env_token" > "$json_file"
    echo "$env_token"
    return 0
  fi

  # 3) credentials -> phone+code
  phone="${!phone_var-}"
  code="${!code_var-}"
  if [[ -n "$phone" && -n "$code" ]]; then
    local t
    t=$(login_with_phone_code "$role" "$phone" "$code" || true)
    if [[ -n "$t" ]]; then
      echo "$t"
      return 0
    fi
  fi

  # 4) credentials -> username+password
  username="${!username_var-}"
  password="${!password_var-}"
  if [[ -n "$username" && -n "$password" ]]; then
    local t
    t=$(login_with_username_password "$role" "$username" "$password" || true)
    if [[ -n "$t" ]]; then
      echo "$t"
      return 0
    fi
  fi

  # 5) fallback: devlogin_resp.json (if exists)
  if [[ -s "devlogin_resp.json" ]]; then
    local t
    t=$(jq -r '(.token // .data.token // empty)' devlogin_resp.json 2>/dev/null || true)
    if [[ -n "$t" && "$t" != "null" ]]; then
      printf '{"token":"%s"}' "$t" > "$json_file"
      echo "$t"
      return 0
    fi
  fi

  # 6) give up
  echo ""  # empty token
  return 0
}

# Generate tokens
ADMIN_TOKEN="$(ensure_token admin)"
USER_TOKEN="$(ensure_token user)"

# Save convenience files
printf '%s' "${ADMIN_TOKEN:-}" > "$OUT_DIR/admin_token.txt"
printf '%s' "${USER_TOKEN:-}" > "$OUT_DIR/user_token.txt"

# Emit tokens.env for sourcing
{
  echo "export API_BASE='$API_BASE'"
  echo "export ADMIN_TOKEN='${ADMIN_TOKEN:-}'"
  echo "export USER_TOKEN='${USER_TOKEN:-}'"
} > "$OUT_DIR/tokens.env"

# Summaries
mask() { local s="$1"; if [[ -z "$s" ]]; then echo "<empty>"; else echo "${s:0:12}...(${#s} chars)"; fi }

echo "[prepare_tokens] API_BASE: $API_BASE"
echo "[prepare_tokens] ADMIN_TOKEN: $(mask "$ADMIN_TOKEN")"
echo "[prepare_tokens] USER_TOKEN : $(mask "$USER_TOKEN")"

if [[ -z "$ADMIN_TOKEN" ]]; then
  echo "[prepare_tokens][WARN] ADMIN_TOKEN is empty. Admin-only APIs will 401. Provide ADMIN_TOKEN or ADMIN_PHONE/ADMIN_CODE." >&2
fi
if [[ -z "$USER_TOKEN" ]]; then
  echo "[prepare_tokens][WARN] USER_TOKEN is empty. User APIs will 401. Provide USER_TOKEN or USER_PHONE/USER_CODE." >&2
fi

exit 0
