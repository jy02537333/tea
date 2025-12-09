#!/usr/bin/env bash
set -euo pipefail
#!/usr/bin/env bash
set -euo pipefail

# remove-large-files.sh
# 简介：重写 Git 历史以移除超过指定大小的 blob（默认 100M）。
# 警告：这是不可逆的历史重写操作，推送后所有协作者需要重新克隆或重置。

THRESHOLD="${1:-100M}"

if ! command -v git-filter-repo >/dev/null 2>&1; then
  echo "错误：未找到 git-filter-repo。请先安装，如："
  echo "  pipx install git-filter-repo  或  pip3 install --user git-filter-repo"
  echo "文档：https://github.com/newren/git-filter-repo"
  exit 1
fi

if [ ! -d .git ]; then
  echo "错误：请在 Git 仓库根目录运行此脚本（包含 .git 目录）。"
  exit 1
fi

echo "即将重写历史并剥离所有大于 ${THRESHOLD} 的 blob。"
echo "注意：操作会改写提交历史。建议先创建备份分支并与团队确认。"

# 运行历史重写
git filter-repo --force --strip-blobs-bigger-than "${THRESHOLD}"

# 回收不可达对象，缩小仓库体积
git reflog expire --expire=now --all
git gc --prune=now --aggressive

current_branch="$(git rev-parse --abbrev-ref HEAD)"
echo "\n完成。本地历史已重写。"
echo "下一步：使用安全方式推送当前分支："
echo "  git push --force-with-lease origin ${current_branch}"
echo "提示：推送后，其他协作者需要重新克隆或执行硬重置对齐历史。"

# remove-large-files.sh
# 清理超过 100MB 的大对象，并移除指定路径（如备份包），随后进行垃圾回收并安全强制推送当前分支。
# 使用前请确保：
# - 已在 GitHub 完成凭据登录（gh auth login 或配置 PAT）
# - 已知需要移除的具体路径（例如 deleted_backups/*.tar.gz）
# - 已安装 git-filter-repo（推荐：sudo apt install git-filter-repo 或 pipx install git-filter-repo）
#
# 用法：
#   scripts/remove-large-files.sh [path_to_remove]
# 示例：
#   scripts/remove-large-files.sh deleted_backups/deleted_backup_20251207_160850.tar.gz

PATH_TO_REMOVE="${1:-}"
ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
echo "[info] repo: $ROOT, branch: $CURRENT_BRANCH"

# 1) 删除工作区指定大文件并忽略同类文件（如 .tar.gz 备份）
if [[ -n "$PATH_TO_REMOVE" ]]; then
  if [ -f "$PATH_TO_REMOVE" ]; then
    echo "[info] removing working-tree file: $PATH_TO_REMOVE"
    git rm -f --ignore-unmatch "$PATH_TO_REMOVE" || true
    rm -f "$PATH_TO_REMOVE" || true
  fi
fi

# 添加忽略规则（避免再次提交备份包）
if ! grep -q "deleted_backups/*.tar.gz" .gitignore 2>/dev/null; then
  printf "\n# ignore backup tarballs\ndeleted_backups/*.tar.gz\n" >> .gitignore
  git add .gitignore || true
fi

git add -A
(git commit -m "chore: remove large backup tarball and ignore future ones" || true)

# 2) 使用 git-filter-repo 清理历史中大对象（>100M），并移除指定路径（若提供）
if command -v git-filter-repo >/dev/null 2>&1; then
  echo "[info] running git-filter-repo..."
  if [[ -n "$PATH_TO_REMOVE" ]]; then
    git filter-repo --force --strip-blobs-bigger-than 100M \
      --invert-paths --paths "$PATH_TO_REMOVE" || {
        echo "[error] git-filter-repo failed"; exit 1;
      }
  else
    git filter-repo --force --strip-blobs-bigger-than 100M || {
      echo "[error] git-filter-repo failed"; exit 1;
    }
  fi
else
  echo "[error] git-filter-repo not found. Install via: sudo apt install git-filter-repo (or pipx install git-filter-repo)"
  exit 1
fi

# 3) 垃圾回收
(git reflog expire --expire=now --all || true)
(git gc --prune=now --aggressive || true)

# 4) 推送当前分支（使用 --force-with-lease 更安全）
echo "[info] pushing branch $CURRENT_BRANCH with --force-with-lease..."
 git push --force-with-lease origin "$CURRENT_BRANCH"

echo "[done] large-file cleanup complete and pushed. If remote still blocks, check for remaining >100MB blobs via git-filter-repo or verify remote protections."
