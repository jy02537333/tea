#!/usr/bin/env bash
set -euo pipefail

# Auto fallback reupload for ZIP when link FAIL (404) and local file exists.
# - Fetch PR body, extract ZIP URL
# - HEAD check; if not 200 and local file exists, reupload via upload_artifacts_to_oss.sh
# - Update PR body mapping via rewrite_evidence_block.py
# - Optionally update README link if changed and push to PR branch

API=${API:-"https://api.github.com"}
OWNER=${OWNER:?}
REPO_NAME=${REPO_NAME:?}
PR_NUMBER=${PR_NUMBER:?}
GH_TOKEN=${GH_TOKEN:?}
BRANCH=${BRANCH:-"${GITHUB_HEAD_REF:-}"}

WORKDIR=$(pwd)
ZIP_LOCAL="build-ci-logs/playwright/partner-withdrawal-trace.zip"

fetch_body() {
  curl -sS -H "Authorization: token $GH_TOKEN" -H "Accept: application/vnd.github+json" \
    "$API/repos/$OWNER/$REPO_NAME/issues/$PR_NUMBER" | jq -r '.body // ""'
}

extract_zip_url() {
  python3 - <<'PY'
import re, sys
body=sys.stdin.read()
m=re.search(r"https?://\S+/ci_artifact/\d{4}/\d{2}/\d{2}/[^\s]*?trace\.zip", body)
print(m.group(0) if m else "")
PY
}

head_status() {
  local url="$1"
  curl -s -o /dev/null -w "%{http_code}" -I "$url" 2>/dev/null || true
}

update_pr_body() {
  python3 scripts/rewrite_evidence_block.py --owner "$OWNER" --repo "$REPO_NAME" --pr "$PR_NUMBER" --checkbox checked --urls-file build-ci-logs/e2e_urls.json || true
}

update_readme_link() {
  local url="$1"
  if [[ -n "$BRANCH" ]]; then
    git config user.name github-actions
    git config user.email github-actions@github.com
    # Replace in README if present
    sed -i -E "s#(- \[x\] Playwright Trace 压缩包：)\S+#\1${url}#" README.md || true
    if ! git diff --quiet -- README.md; then
      git add README.md
      git commit -m "docs(evidence): update trace ZIP link after fallback reupload"
      git push origin "$BRANCH" || true
    fi
  fi
}

main() {
  local body zip_url status
  body=$(fetch_body)
  zip_url=$(printf '%s' "$body" | extract_zip_url)
  if [[ -z "$zip_url" ]]; then
    echo "[auto_fallback_zip] No ZIP url found; skip." >&2
    return 0
  fi
  status=$(head_status "$zip_url")
  if [[ "$status" = "200" ]]; then
    echo "[auto_fallback_zip] ZIP url healthy (200); skip." >&2
    return 0
  fi
  if [[ ! -s "$ZIP_LOCAL" ]]; then
    echo "[auto_fallback_zip] Local ZIP not found ($ZIP_LOCAL); cannot reupload; skip." >&2
    return 0
  fi
  echo "[auto_fallback_zip] ZIP head=$status, reupload via octet fallback..." >&2
  chmod +x scripts/upload_artifacts_to_oss.sh || true
  scripts/upload_artifacts_to_oss.sh "$ZIP_LOCAL" || true
  # After reupload, update PR body mapping (usually unchanged)
  update_pr_body || true
  # Update README line if needed
  update_readme_link "$zip_url" || true
}

main "$@"
