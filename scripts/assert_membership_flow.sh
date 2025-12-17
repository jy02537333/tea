#!/usr/bin/env bash
set -euo pipefail

LOG_DIR=build-ci-logs
CHECK_JSON="$LOG_DIR/membership_b_flow_checked.json"
REQUIRE_MEMBERSHIP_CHECK=${REQUIRE_MEMBERSHIP_CHECK:-0}

if [[ -f "$CHECK_JSON" ]]; then
  if jq -e '.ok==true' "$CHECK_JSON" >/dev/null 2>&1; then
    level=$(jq -r '.membership_level // ""' "$CHECK_JSON" 2>/dev/null || echo "")
    echo "Sprint B membership flow check passed (level=${level:-""})."
    exit 0
  else
    echo "ERROR: Sprint B membership flow check failed (.ok!=true) in $CHECK_JSON" >&2
    exit 3
  fi
else
  if [[ "$REQUIRE_MEMBERSHIP_CHECK" = "1" ]]; then
    echo "ERROR: Sprint B membership evidence not found while REQUIRE_MEMBERSHIP_CHECK=1" >&2
    echo "Expected: $CHECK_JSON" >&2
    exit 2
  else
    echo "WARNING: Sprint B membership evidence not found; skipping membership check (non-strict mode)." >&2
    exit 0
  fi
fi
