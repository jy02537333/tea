#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$REPO_ROOT"

LOG_DIR="$REPO_ROOT/build-ci-logs"
mkdir -p "$LOG_DIR"
API_LOG_DIR="$LOG_DIR/api_validation"
mkdir -p "$API_LOG_DIR"

# Ensure port 9292 is free; if occupied, kill the old process gracefully
ensure_port_free() {
  local port=9292
  if [ -n "${TEA_SERVER_PORT:-}" ]; then
    # allow override via env, e.g. TEA_SERVER_PORT=9393
    port="${TEA_SERVER_PORT#:}"
  fi
  echo "[run-tea-api] Checking if port ${port} is occupied..."

  # Prefer PID from pid file if available
  local pidfile="$LOG_DIR/tea-api.pid"
  if [ -f "$pidfile" ]; then
    local oldpid
    oldpid=$(cat "$pidfile" 2>/dev/null || true)
    if [ -n "${oldpid:-}" ] && kill -0 "$oldpid" 2>/dev/null; then
      echo "[run-tea-api] Found previous PID from pidfile: $oldpid -> terminating"
      kill "$oldpid" 2>/dev/null || true
      sleep 1
    fi
  fi

  # Fallback: detect any process listening on the port
  local pids
  pids=$(ss -lntp 2>/dev/null | awk -v port=":${port}" '$4 ~ port {print $NF}' | sed -E 's/users:\(\("[^"]+",pid=([0-9]+).*/\1/' | sort -u)
  if [ -n "$pids" ]; then
    echo "[run-tea-api] Port ${port} occupied by PIDs: $pids -> terminating"
    for p in $pids; do
      kill "$p" 2>/dev/null || true
    done
    sleep 1
  fi

  # fallback: use fuser if available (may require sudo depending on owner)
  if ss -lntp 2>/dev/null | grep -q ":${port} "; then
    if command -v fuser >/dev/null 2>&1; then
      echo "[run-tea-api] Trying fuser -k ${port}/tcp"
      fuser -k "${port}/tcp" 2>/dev/null || true
      sleep 1
    fi
  fi

  # Double-check
  if ss -lntp 2>/dev/null | grep -q ":${port} "; then
    echo "[run-tea-api] Port ${port} still occupied after kill attempts. Using kill -9 as last resort."
    pids=$(ss -lntp 2>/dev/null | awk -v port=":${port}" '$4 ~ port {print $NF}' | sed -E 's/users:\(\("[^"]+",pid=([0-9]+).*/\1/' | sort -u)
    for p in $pids; do
      kill -9 "$p" 2>/dev/null || true
    done
    sleep 1
  fi

  if ss -lntp 2>/dev/null | grep -q ":${port} "; then
    echo "[run-tea-api] ERROR: Port ${port} is still occupied. Aborting."
    exit 1
  fi
  echo "[run-tea-api] Port ${port} is free."
}

ensure_port_free

echo "[run-tea-api] Building tea-api main.go binary (unified entry)..."
cd tea-api
GOOS=$(go env GOOS 2>/dev/null || echo linux)
GOARCH=$(go env GOARCH 2>/dev/null || echo amd64)
go mod tidy >/dev/null 2>&1 || true
# Output binary to project root without clashing with tea-api directory name
go build -o ../tea-api.bin ./
cd "$REPO_ROOT"

# Environment (example values)
export TEA_DSN='root:gs963852@tcp(127.0.0.1:3308)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local'
export TEA_REDIS_HOST=127.0.0.1
export TEA_REDIS_PORT=6379
export TEA_REDIS_PASSWORD=123456
export TEA_RABBITMQ_HOST=127.0.0.1
export TEA_RABBITMQ_PORT=5672
export TEA_RABBITMQ_USERNAME=guest
export TEA_RABBITMQ_PASSWORD=guest
export TEA_RABBITMQ_VHOST=/

# Safety: ensure automatic migration is disabled by default when starting via this script.
# If you need to run automatic GORM AutoMigrate on startup, set TEA_AUTO_MIGRATE=1 explicitly.
export TEA_AUTO_MIGRATE=${TEA_AUTO_MIGRATE:-0}
echo "[run-tea-api] TEA_AUTO_MIGRATE=$TEA_AUTO_MIGRATE"

# Ensure unified JWT secret (env overrides file), can be provided by caller
export TEA_JWT_SECRET=${TEA_JWT_SECRET:-tea-shop-jwt-secret-key-2023}
echo "[run-tea-api] TEA_JWT_SECRET set (len=${#TEA_JWT_SECRET})"

# Optional: override server port at runtime without editing repo file
RUNTIME_CONFIG="$API_LOG_DIR/config.runtime.yaml"
BASE_CONFIG="tea-api/configs/config.yaml"
if [ -n "${TEA_SERVER_PORT:-}" ]; then
  cp "$BASE_CONFIG" "$RUNTIME_CONFIG"
  # normalize leading colon
  PORT_VAL=":${TEA_SERVER_PORT#:}"
  sed -i "s/port: \".*\"/port: \"$PORT_VAL\"/" "$RUNTIME_CONFIG"
  echo "[run-tea-api] Overriding server.port -> $PORT_VAL via $RUNTIME_CONFIG"
  CFG_PATH="$RUNTIME_CONFIG"
else
  CFG_PATH="$BASE_CONFIG"
fi

echo "[run-tea-api] Starting tea-api (main.go) in background..."
# Choose a writable log file path (avoid root-owned stale files)
LOG_FILE="$API_LOG_DIR/tea-api.out"
if [ -e "$LOG_FILE" ] && [ ! -w "$LOG_FILE" ]; then
  ts=$(date +%s)
  LOG_FILE="$API_LOG_DIR/tea-api.$ts.out"
  echo "[run-tea-api] Detected non-writable log file; using $LOG_FILE instead"
fi
nohup ./tea-api.bin -config "$CFG_PATH" > "$LOG_FILE" 2>&1 &
API_PID=$!
echo $API_PID > "$LOG_DIR/tea-api.pid"
echo "[run-tea-api] PID: $API_PID"

# Wait for health endpoint
PORT_USE=${TEA_SERVER_PORT:-9292}
HEALTH_URL="http://127.0.0.1:${PORT_USE}/api/v1/health"
MAX_WAIT=60
i=0
while [ $i -lt $MAX_WAIT ]; do
  status=$(curl -sS -o /dev/null -w "%{http_code}" "$HEALTH_URL" || true)
  echo "[run-tea-api] health check attempt $i -> $status"
  if [ "$status" = "200" ]; then
    break
  fi
  i=$((i+1))
  sleep 1
done

if [ "$status" != "200" ]; then
  echo "[run-tea-api] health check failed after $MAX_WAIT seconds. See $LOG_FILE"
  exit 2
fi

# Try developer login to obtain admin token
LOGIN_URL="http://127.0.0.1:${PORT_USE}/api/v1/auth/login"
RESP_FILE="$LOG_DIR/admin_login_response.json"

# First attempt: username/password
curl -sS -H "Content-Type: application/json" -d '{"username":"admin","password":"pass"}' "$LOGIN_URL" -o "$RESP_FILE" || true
TOKEN=$(python3 - "$RESP_FILE" <<'PY'
import sys, json
try:
    obj = json.load(open(sys.argv[1]))
except Exception:
    print("")
    sys.exit(0)
data = obj.get('data') or obj
print((data.get('token') if isinstance(data, dict) else "") or "")
PY
)

# If token empty, try openid login
if [ -z "$TOKEN" ]; then
  curl -sS -H "Content-Type: application/json" -d '{"openid":"admin_openid"}' "$LOGIN_URL" -o "$RESP_FILE" || true
  TOKEN=$(python3 - "$RESP_FILE" <<'PY'
import sys, json
try:
    obj = json.load(open(sys.argv[1]))
except Exception:
    print("")
    sys.exit(0)
data = obj.get('data') or obj
print((data.get('token') if isinstance(data, dict) else "") or "")
PY
  )
fi

# If still empty, fallback to dev-login in local/dev envs
if [ -z "$TOKEN" ]; then
  curl -sS -H "Content-Type: application/json" -d '{"openid":"admin_openid"}' "http://127.0.0.1:${PORT_USE}/api/v1/user/dev-login" -o "$RESP_FILE" || true
  TOKEN=$(python3 - "$RESP_FILE" <<'PY'
import sys, json
try:
    obj = json.load(open(sys.argv[1]))
except Exception:
    print("")
    sys.exit(0)
data = obj.get('data') or obj
print((data.get('token') if isinstance(data, dict) else "") or "")
PY
  )
fi

if [ -z "$TOKEN" ]; then
  echo "[run-tea-api] Failed to obtain token from $LOGIN_URL. See $RESP_FILE and server log."
  exit 3
fi

echo "[run-tea-api] Obtained token (truncated): ${TOKEN:0:16}..."

# Fetch admin users
USERS_URL="http://127.0.0.1:${PORT_USE}/api/v1/admin/users?page=1&limit=20"
USERS_OUT="$LOG_DIR/admin_users_response.json"
HTTP_STATUS=$(curl -sS -H "Authorization: Bearer $TOKEN" -o "$USERS_OUT" -w "%{http_code}" "$USERS_URL" || true)

echo "[run-tea-api] GET $USERS_URL -> HTTP $HTTP_STATUS"
if [ -s "$USERS_OUT" ]; then
  echo "[run-tea-api] Saved response to $USERS_OUT"
else
  echo "[run-tea-api] No response body saved. See logs."
fi

# Done
echo "[run-tea-api] Completed. Logs: $API_LOG_DIR/tea-api.out, login: $RESP_FILE, users: $USERS_OUT"
exit 0
