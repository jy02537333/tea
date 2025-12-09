#!/usr/bin/env bash
set -euo pipefail

# 目标分支（默认当前分支）
BRANCH="${1:-$(git rev-parse --abbrev-ref HEAD)}"
ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

echo "注意：请先在 GitHub 上撤销/轮换被泄露的 token。确认已完成后输入 YES："
read -r CONFIRM
if [ "$CONFIRM" != "YES" ]; then
  echo "已取消。请先撤销/轮换 token 后重试。"
  exit 1
fi

# 敏感文件列表（根据 push 阻断日志）
SENSITIVE_FILES=(
  "pwd.bmiwj"
  "wx-fe/pwd2.bmiwj"
  "tea-api/configs/config.mysql.local.yaml"
)

# 需要整目录或通配删除的路径（glob）
SENSITIVE_GLOBS=(
  "tea-api/dist/**"
)

"git" checkout "$BRANCH"
for f in "${SENSITIVE_FILES[@]}"; do
  if [ -f "$f" ] || [ -f "$ROOT/$f" ]; then
    git rm -f --ignore-unmatch "$f" || true
    rm -f "$f" || true
  fi
done

# 更新 .gitignore
if ! grep -q "\*.bmiwj" .gitignore 2>/dev/null; then
  printf "\n# prevent committing bmiwj secrets\n*.bmiwj\n" >> .gitignore
  git add .gitignore
fi

git add -A
git commit -m "chore(secrets): remove sensitive files from working tree and ignore" || true

REMOTE_URL="$(git remote get-url origin 2>/dev/null || true)"

# 2) 重写历史以彻底删除这些文件
# 优先使用 git-filter-repo（更快/更可靠），否则 fallback 到 BFG
if command -v git-filter-repo >/dev/null 2>&1; then
  echo "使用 git-filter-repo 清理历史..."
  # 备份当前分支并创建同名新分支（避免覆盖）
  backup_branch="${BRANCH}-backup-$(date +%Y%m%d%H%M%S)"
  git branch -m "${BRANCH}" "${backup_branch}" || true
  git checkout -b "${BRANCH}"
  # 将敏感路径写入临时文件供过滤使用
  tmp_paths_file="$(mktemp)"
  for p in "${SENSITIVE_FILES[@]}"; do
    printf "%s\n" "$p" >> "$tmp_paths_file"
  done
  # 运行过滤（invert-paths 删除指定文件/目录）
  args=(--invert-paths --paths-from-file "$tmp_paths_file")
  for g in "${SENSITIVE_GLOBS[@]}"; do
    args+=(--path-glob "$g")
  done
  git filter-repo "${args[@]}" || {
    echo "git-filter-repo 失败，请检查版本。"
    rm -f "$tmp_paths_file"
    exit 1
  }
  rm -f "$tmp_paths_file"
else
  echo "未检测到 git-filter-repo，尝试使用 BFG（需预先安装 bfg）..."
  if ! command -v bfg >/dev/null 2>&1; then
    echo "未安装 BFG。请安装 git-filter-repo 或 BFG 后重试。退出。"
    exit 1
  fi
  # 使用 BFG 删除敏感文件
  echo "创建临时裸仓库副本供 BFG 使用..."
  git clone --mirror . repo-mirror.git
  pushd repo-mirror.git >/dev/null
  # 删除指定文件
  for f in "${SENSITIVE_FILES[@]}"; do
    bfg --delete-files "$f" || true
  done
  git reflog expire --expire=now --all
  git gc --prune=now --aggressive
  popd >/dev/null
  # 将清理后的镜像推回本地仓库目录（慎用）
  rm -rf .git
  mv repo-mirror.git .git
  git reset --hard
fi

# 3) 强制回收垃圾与检查
git reflog expire --expire=now --all || true
git gc --prune=now --aggressive || true

# 4) 强推到远端（使用 --force-with-lease 更安全）
if ! git remote get-url origin >/dev/null 2>&1; then
  if [ -n "${REMOTE_URL}" ]; then
    echo "origin 被移除，重新添加：${REMOTE_URL}"
    git remote add origin "${REMOTE_URL}"
  else
    echo "警告：未能恢复 origin 远端，请手动添加后推送。"
  fi
fi

echo "准备强制推送到 origin/${BRANCH} （使用 --force-with-lease）..."
git push --force-with-lease origin "${BRANCH}"

echo "已完成历史清理并推送（如远端 push protection 仍阻塞，请在 GitHub 页面按提示解除或联系仓库管理员）。"
echo "完成后请检查 GitHub 的 Secret Scanning 页面确认泄露已解除。"