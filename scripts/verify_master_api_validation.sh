#!/usr/bin/env bash
set -euo pipefail

# Verify latest master 'API Validation' workflow run, download artifacts,
# and check order evidence files for check=true.

REPO="jy02537333/tea"
WF_FILE="api-validation.yml"
BRANCH="master"
MAX_WAIT_SECS=${MAX_WAIT_SECS:-600}   # 10 minutes default
SLEEP_SECS=${SLEEP_SECS:-10}

command -v gh >/dev/null 2>&1 || { echo "gh CLI not found. Install GitHub CLI first: https://cli.github.com/"; exit 127; }
command -v jq >/dev/null 2>&1 || { echo "jq not found. Install jq first."; exit 127; }

echo "Fetching latest run for $WF_FILE on branch $BRANCH ..."
API="repos/$REPO/actions/workflows/$WF_FILE/runs"
RESP=$(gh api -H "Accept: application/vnd.github+json" "$API" -f branch="$BRANCH" -f per_page=1)
if [[ -z "$RESP" ]]; then
  echo "Failed to query workflow runs." && exit 1
fi

RUN_ID=$(echo "$RESP" | jq -r '.workflow_runs[0].id // empty')
STATUS=$(echo "$RESP" | jq -r '.workflow_runs[0].status // empty')
CONCL=$(echo "$RESP" | jq -r '.workflow_runs[0].conclusion // empty')
URL=$(echo "$RESP" | jq -r '.workflow_runs[0].html_url // empty')

if [[ -z "$RUN_ID" ]]; then
  echo "No recent runs found for $WF_FILE on $BRANCH." && exit 1
fi

echo "Run: $RUN_ID" 
echo "URL: $URL"
echo "Initial status: $STATUS, conclusion: ${CONCL:-n/a}"

# Poll until completed or timeout
WAITED=0
if [[ "$STATUS" != "completed" ]]; then
  echo "Waiting for run to complete ... (max ${MAX_WAIT_SECS}s)"
  while [[ "$STATUS" != "completed" && $WAITED -lt $MAX_WAIT_SECS ]]; do
    sleep "$SLEEP_SECS"
    WAITED=$((WAITED + SLEEP_SECS))
    RESP=$(gh api -H "Accept: application/vnd.github+json" "$API" -f branch="$BRANCH" -f per_page=1)
    STATUS=$(echo "$RESP" | jq -r '.workflow_runs[0].status // empty')
    CONCL=$(echo "$RESP" | jq -r '.workflow_runs[0].conclusion // empty')
    echo "  -> status=$STATUS (waited ${WAITED}s)"
  done
fi

if [[ "$STATUS" != "completed" ]]; then
  echo "Run did not complete within ${MAX_WAIT_SECS}s. Current status=$STATUS"
  exit 2
fi

echo "Final status: $STATUS, conclusion: ${CONCL:-n/a}"

# Download artifacts
DEST_DIR="build-ci-logs/ci_artifacts/run_${RUN_ID}"
mkdir -p "$DEST_DIR"
echo "Downloading artifacts to $DEST_DIR ..."
gh run download --run-id "$RUN_ID" -D "$DEST_DIR"

# Locate evidence
EVID_FILE=$(find "$DEST_DIR" -type f -name 'order_detail_*_checked.json' | head -n1 || true)
if [[ -z "${EVID_FILE:-}" ]]; then
  EVID_FILE=$(find "$DEST_DIR" -type f -name 'order_amounts_summary.json' | head -n1 || true)
fi

CHECK_VAL="missing"
if [[ -n "${EVID_FILE:-}" ]]; then
  CHECK_VAL=$(jq -r '.check // .ok // .Check // "missing"' "$EVID_FILE" 2>/dev/null || echo "missing")
fi

echo "Evidence file: ${EVID_FILE:-not found}"
echo "Evidence check: $CHECK_VAL"

echo "JSONs discovered (depth<=2):"
find "$DEST_DIR" -maxdepth 2 -type f -name '*.json' -printf '%p\n' 2>/dev/null | sed 's/^/ - /'

# Exit non-zero if evidence is present and check is not true
if [[ -n "${EVID_FILE:-}" ]]; then
  if [[ "$CHECK_VAL" != "true" ]]; then
    echo "Evidence found but check!=true (value=$CHECK_VAL)." && exit 3
  fi
fi

echo "Done."
