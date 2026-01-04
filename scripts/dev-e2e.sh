#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT_DIR/build-ci-logs/playwright"
SPA_PORT="${SPA_PORT:-5173}"
API_PORT="${API_PORT:-9393}"
TRACE="${TRACE:-1}"
WORKERS="${WORKERS:-1}"
E2E_TESTS_INPUT="${E2E_TESTS:-all}"

mkdir -p "$LOG_DIR"

cleanup() {
  set +e
  if [[ -f "$LOG_DIR/mock-api.pid" ]]; then
    kill "$(cat "$LOG_DIR/mock-api.pid")" 2>/dev/null || true
    rm -f "$LOG_DIR/mock-api.pid"
  fi
  if [[ -f "$LOG_DIR/serve-spa.pid" ]]; then
    kill "$(cat "$LOG_DIR/serve-spa.pid")" 2>/dev/null || true
    rm -f "$LOG_DIR/serve-spa.pid"
  fi
}
trap cleanup EXIT

echo "[dev-e2e] Ensuring pnpm and playwright are available..."
if ! command -v pnpm >/dev/null 2>&1; then
  corepack enable || true
  corepack prepare pnpm@9.0.0 --activate || true
fi

echo "[dev-e2e] Install admin-fe deps + playwright browsers..."
pnpm -C "$ROOT_DIR/admin-fe" install
pnpm -C "$ROOT_DIR/admin-fe" exec playwright install --with-deps

echo "[dev-e2e] Build admin-fe..."
pnpm -C "$ROOT_DIR/admin-fe" run build

echo "[dev-e2e] Start Mock API on :$API_PORT..."
MOCK_API_PORT="$API_PORT" \
nohup node "$ROOT_DIR/scripts/mock-admin-api.js" \
  >"$LOG_DIR/mock-api.out" 2>&1 &
echo $! > "$LOG_DIR/mock-api.pid"

echo "[dev-e2e] Wait for Mock API ready..."
for i in {1..40}; do
  if curl -sSf "http://127.0.0.1:$API_PORT/api/v1/admin/orders" >/dev/null 2>&1; then
    echo "[dev-e2e] Mock API is ready"
    break
  fi
  sleep 0.25
  if [[ $i -eq 40 ]]; then
    echo "[dev-e2e] Mock API failed to start" >&2
    exit 1
  fi
done

echo "[dev-e2e] Start SPA server on :$SPA_PORT..."
nohup node "$ROOT_DIR/admin-fe/scripts/serve-spa.js" "$SPA_PORT" \
  >"$LOG_DIR/serve-spa.out" 2>&1 &
echo $! > "$LOG_DIR/serve-spa.pid"

echo "[dev-e2e] Wait for SPA ready..."
for i in {1..40}; do
  if curl -sSf "http://127.0.0.1:$SPA_PORT" >/dev/null 2>&1; then
    echo "[dev-e2e] SPA server is ready"
    break
  fi
  sleep 0.25
  if [[ $i -eq 40 ]]; then
    echo "[dev-e2e] SPA failed to start" >&2
    exit 1
  fi
done

echo "[dev-e2e] Determine tests to run..."
TEST_ARGS=( )
case "$E2E_TESTS_INPUT" in
  all|ALL|*)
    if [[ "$E2E_TESTS_INPUT" == "all" || "$E2E_TESTS_INPUT" == "ALL" ]]; then
      TEST_ARGS+=(
        "tests/orders-reason-modal.spec.ts"
        "tests/store-orders-refund.spec.ts"
        "tests/store-orders-cancel.spec.ts"
        "tests/order-adjust.spec.ts"
        "tests/partner-withdrawal.spec.ts"
      )
    else
      # custom input: allow passing file paths or globs, split by spaces
      # shellcheck disable=SC2206
      TEST_ARGS=($E2E_TESTS_INPUT)
    fi
    ;;
  orders)
    TEST_ARGS+=("tests/orders-reason-modal.spec.ts")
    ;;
  withdrawal|withdrawals)
    TEST_ARGS+=("tests/partner-withdrawal.spec.ts")
    ;;
esac

echo "[dev-e2e] Run Playwright tests with TRACE=$TRACE, WORKERS=$WORKERS ..."
TRACE="$TRACE" \
PLAYWRIGHT_HTML_REPORT="$ROOT_DIR/admin-fe/playwright-report" \
PLAYWRIGHT_HTML_OPEN="never" \
ADMIN_FE_URL="http://127.0.0.1:$SPA_PORT" \
API_BASE="http://127.0.0.1:$API_PORT" \
pnpm -C "$ROOT_DIR/admin-fe" exec playwright test \
  "${TEST_ARGS[@]}" \
  --workers="$WORKERS" \
  --reporter=dot,html

echo "[dev-e2e] Done. Artifacts in $LOG_DIR"
