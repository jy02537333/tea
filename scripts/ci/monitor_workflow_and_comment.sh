#!/usr/bin/env bash
set -euo pipefail

# Optional env vars:
#   GH_TOKEN: GitHub token with repo/workflow scope (optional for read-only; required to auto-comment)
#   OWNER: repo owner (e.g., jy02537333)
#   REPO: repo name (e.g., tea)
#   PR_NUMBER: pull request number (e.g., 52)
#   BRANCH: head branch name for the PR
#   WORKFLOW_FILE: workflow file name (e.g., api-validation.yml)
# Optional:
#   POLL_INTERVAL: seconds between polls (default 30)
#   RUN_ID: if provided, monitor this specific workflow run id

: "${OWNER:?OWNER is required}"
: "${REPO:?REPO is required}"
: "${PR_NUMBER:?PR_NUMBER is required}"
: "${BRANCH:?BRANCH is required}"
: "${WORKFLOW_FILE:?WORKFLOW_FILE is required}"

POLL_INTERVAL=${POLL_INTERVAL:-30}
API="https://api.github.com"
AUTH=(-H "Accept: application/vnd.github+json")
# Fallback: read token from local file if not set (and avoid printing it)
if [[ -z "${GH_TOKEN:-}" && -f .github_token ]]; then
  GH_TOKEN=$(cat .github_token)
fi
if [[ -n "${GH_TOKEN:-}" ]]; then
  AUTH=(-H "Authorization: token ${GH_TOKEN}" -H "Accept: application/vnd.github+json")
fi

err() { echo "[monitor] $*" 1>&2; }
log() { echo "[monitor] $*"; }

json_field() {
  # Usage: echo "$json" | json_field 'key.subkey'
  python3 -c '
import sys, json
path = sys.argv[1]
obj = json.load(sys.stdin)
cur = obj
for p in path.split("."):
  cur = (cur or {}).get(p)
if isinstance(cur, (dict, list)):
  print(json.dumps(cur))
elif cur is None:
  print("")
else:
  print(cur)
' "$1"
}

get_run_id() {
  local runs_json
  runs_json=$(curl -sS "${API}/repos/${OWNER}/${REPO}/actions/workflows/${WORKFLOW_FILE}/runs?branch=${BRANCH}&event=pull_request&per_page=1" "${AUTH[@]}")
  if [[ -z "$runs_json" ]]; then
    err "Empty runs response"
    return 1
  fi
  echo "$runs_json" | python3 -c 'import sys,json; j=json.load(sys.stdin); r=j.get("workflow_runs") or []; print(r[0]["id"] if r else "")'
}

get_run_json() {
  local run_id=$1
  curl -sS "${API}/repos/${OWNER}/${REPO}/actions/runs/${run_id}" "${AUTH[@]}"
}

post_pr_comment() {
  local body=$1
  if [[ -z "${GH_TOKEN:-}" ]]; then
    mkdir -p build-ci-logs || true
    printf "%s\n" "$body" > build-ci-logs/monitor_pr${PR_NUMBER}_comment.txt
    log "GH_TOKEN not set; wrote comment draft to build-ci-logs/monitor_pr${PR_NUMBER}_comment.txt"
    return 0
  fi
  curl -sS -X POST "${API}/repos/${OWNER}/${REPO}/issues/${PR_NUMBER}/comments" \
    "${AUTH[@]}" \
    -H 'Content-Type: application/json' \
    -d "$(printf '{"body":%s}' "$(python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))' <<<"$body")")" >/dev/null
}

main() {
  local run_id
  if [[ -n "${RUN_ID:-}" ]]; then
    run_id="$RUN_ID"
    log "Monitoring specified RUN_ID=${run_id} (branch=${BRANCH})"
  else
    log "Locating latest workflow run for ${BRANCH} (${WORKFLOW_FILE})..."
    run_id=$(get_run_id)
    if [[ -z "$run_id" ]]; then
      err "No workflow run found for branch=${BRANCH} event=pull_request"
      exit 2
    fi
    log "Monitoring run_id=${run_id} (branch=${BRANCH})"
  fi

  local status conclusion html_url
  while true; do
    local run_json
    run_json=$(get_run_json "$run_id")
    status=$(echo "$run_json" | json_field 'status')
    conclusion=$(echo "$run_json" | json_field 'conclusion')
    html_url=$(echo "$run_json" | json_field 'html_url')

    log "status=${status} conclusion=${conclusion} url=${html_url}"

    if [[ "$status" == "completed" ]]; then
      break
    fi
    sleep "$POLL_INTERVAL"
  done

  # Build PR comment in Chinese with result and link
  local result_emoji result_text
  if [[ "$conclusion" == "success" ]]; then
    result_emoji="✅"
    result_text="本次 API Validation 已完成且通过 (success)。"
  else
    result_emoji="❌"
    result_text="本次 API Validation 已完成但未通过 (conclusion=${conclusion:-unknown})。"
  fi

  local comment
  comment=$(cat <<EOF
${result_emoji} CI 运行完成：${result_text}

- 运行链接：${html_url}
- 分支：${BRANCH}
- 工作流：${WORKFLOW_FILE}

如需查看证据文件与日志，请在运行页面的 Artifacts 中下载（已归档 Sprint A/B 相关 JSON/CSV）。
EOF
)

  log "Posting PR #${PR_NUMBER} comment..."
  post_pr_comment "$comment"
  log "Done."
  # Persist a brief JSON result for other tools to consume
  mkdir -p build-ci-logs || true
  printf '{"run_id":%s,"status":"%s","conclusion":"%s","html_url":%s}\n' \
    "$run_id" "$status" "$conclusion" "$(python3 -c 'import json,sys; print(json.dumps(sys.argv[1]))' "$html_url")" \
    > build-ci-logs/monitor_pr${PR_NUMBER}_result.json
}

main "$@"
