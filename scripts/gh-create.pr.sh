#!/usr/bin/env bash
set -euo pipefail

# 使用示例：
# GIT_USER_NAME="Your Name" GIT_USER_EMAIL="you@example.com" PR_TITLE="ci: add diagnostics" PR_BODY="..." BASE_BRANCH="feat/frontend-scaffold" BRANCH="feat/ci-diagnostics" ./scripts/gh-create-pr.sh

# 可配置项（环境变量优先），也可在脚本中直接修改默认值
BASE_BRANCH="${BASE_BRANCH:-feat/frontend-scaffold}"
BRANCH="${BRANCH:-$(git rev-parse --abbrev-ref HEAD)}"
PR_TITLE="${PR_TITLE:-"ci: add diagnostics workflow for ui-thumbnail/wx-fe"}"
PR_BODY="${PR_BODY:-"Add CI diagnostics script/workflow and minimal package/config tweaks to reproduce and collect build logs for wx-fe Taro build."}"
GIT_USER_NAME="${GIT_USER_NAME:-}"
GIT_USER_EMAIL="${GIT_USER_EMAIL:-}"
ASSIGNEE="${ASSIGNEE:-}"       # 可选
REVIEWER="${REVIEWER:-}"       # 可选, 多个可逗号分隔
LABELS="${LABELS:-ci,diagnostics}" # 可选
DRAFT="${DRAFT:-false}"        # 设置为 "true" 则创建草稿 PR
OPEN="${OPEN:-false}"          # 创建后是否在浏览器中打开 PR

# 确保在仓库根运行
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo ".")"
cd "$REPO_ROOT"

# 设置本仓库 git 身份（仅当提供环境变量时）
if [ -n "$GIT_USER_EMAIL" ]; then
  git config user.email "$GIT_USER_EMAIL"
fi
if [ -n "$GIT_USER_NAME" ]; then
  git config user.name "$GIT_USER_NAME"
fi

# 确保分支存在
if ! git rev-parse --verify "$BRANCH" >/dev/null 2>&1; then
  echo "ERROR: 分支 '$BRANCH' 不存在。本地请先创建并切换到该分支。"
  exit 1
fi

# 确保有提交（如有暂存改动则提交）
if ! git diff --staged --quiet || ! git diff --quiet; then
  echo "⚠️  Uncommitted changes detected / 检测到未提交的改动"
  echo "✅ Proceeding: Auto-committing all changes / 正在自动提交所有改动..."
  git add -A
  git commit -m "${PR_TITLE}" || true
  echo "✅ Changes committed successfully / 改动已提交"
fi

# 推送分支到远端（设置上游）
echo "推送分支 ${BRANCH} 到远端..."
git push --set-upstream origin "$BRANCH"

# 检查 gh CLI 和登陆
if ! command -v gh >/dev/null 2>&1; then
  echo "ERROR: 未检测到 gh CLI，请安装并登录（gh auth login），然后重试。"
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "ERROR: gh 未登录，请运行: gh auth login"
  exit 1
fi

# 检查是否已有针对相同 base/head 的 PR
EXISTING_PR_URL="$(gh pr list --head "$BRANCH" --base "$BASE_BRANCH" --json url --jq '.[0].url' 2>/dev/null || true)"

if [ -n "$EXISTING_PR_URL" ] && [ "$EXISTING_PR_URL" != "null" ]; then
  echo "已存在 PR： $EXISTING_PR_URL"
  if [ "$OPEN" = "true" ]; then
    gh web -- "$EXISTING_PR_URL" >/dev/null 2>&1 || true
  fi
  exit 0
fi

# 构建 gh pr create 命令
GH_CMD=(gh pr create --base "$BASE_BRANCH" --head "$BRANCH" --title "$PR_TITLE" --body "$PR_BODY")
[ -n "$ASSIGNEE" ] && GH_CMD+=(--assignee "$ASSIGNEE")
# 支持多个 reviewer 用逗号分隔
if [ -n "$REVIEWER" ]; then
  IFS=',' read -ra RVS <<< "$REVIEWER"
  for r in "${RVS[@]}"; do
    GH_CMD+=(--reviewer "$r")
  done
fi
[ -n "$LABELS" ] && GH_CMD+=(--label "$LABELS")
if [ "$DRAFT" = "true" ]; then
  GH_CMD+=(--draft)
fi

# 执行创建并捕获 URL 或错误
echo "创建 PR..."
PR_URL="$("${GH_CMD[@]}" --json url --jq '.url' 2>/dev/null || true)"

if [ -z "$PR_URL" ]; then
  # 兼容老版本 gh，尝试不使用 --json
  PR_URL="$("${GH_CMD[@]}" 2>/dev/null | awk '/https?:\/\/github.com/ {print $0; exit}' || true)"
fi

if [ -z "$PR_URL" ]; then
  echo "创建 PR 可能失败，请手动运行以下命令查看原因："
  echo "${GH_CMD[*]}"
  exit 1
fi

echo "PR 已创建： $PR_URL"

if [ "$OPEN" = "true" ]; then
  echo "在浏览器中打开 PR..."
  gh web -- "$PR_URL" >/dev/null 2>&1 || true
fi

exit 0