#!/usr/bin/env bash
set -euo pipefail

OWNER=jy02537333
REPO=tea
BRANCH=chore/ci-disable-api-validation-on-master
WF=api-validation.yml
API="https://api.github.com"

if [[ ! -f .github_token ]]; then
  echo "[rerun] .github_token not found; cannot authenticate" >&2
  exit 2
fi
GH_TOKEN=$(cat .github_token)
AUTH=(-H "Authorization: token ${GH_TOKEN}" -H "Accept: application/vnd.github+json")

mkdir -p build-ci-logs

echo "[rerun] Fetching latest run id for branch=${BRANCH} wf=${WF}..."
RUNS_API="${API}/repos/${OWNER}/${REPO}/actions/workflows/${WF}/runs?branch=${BRANCH}&event=pull_request&per_page=1"
RUN_ID=$(curl -sS "${RUNS_API}" "${AUTH[@]}" | python3 -c 'import sys,json;j=json.load(sys.stdin);rs=j.get("workflow_runs") or [];print(rs[0]["id"] if rs else "")')
if [[ -z "$RUN_ID" ]]; then
  echo "[rerun] No workflow run found for branch=${BRANCH}" >&2
  exit 2
fi
echo "$RUN_ID" > build-ci-logs/last_pr52_run_id.txt
echo "[rerun] Target run_id=${RUN_ID}"

echo "[rerun] Triggering rerun..."
curl -sS -X POST "${API}/repos/${OWNER}/${REPO}/actions/runs/${RUN_ID}/rerun" "${AUTH[@]}" >/dev/null
echo "[rerun] Rerun requested. Starting monitor in background..."

nohup env RUN_ID="$RUN_ID" make monitor-pr52-ci > build-ci-logs/monitor_pr52_bg.log 2>&1 & echo $! > build-ci-logs/monitor_pr52_bg.pid
echo "[rerun] Monitor started. pid=$(cat build-ci-logs/monitor_pr52_bg.pid)"
echo "[rerun] Tail log: build-ci-logs/monitor_pr52_bg.log"
