#!/usr/bin/env bash
set -euo pipefail

# 文件用途：
# - 在“维护窗口”内对指定分支（默认 master）进行一次性历史重写，移除超阈值大文件与敏感路径。
# - 自动处理 git-filter-repo 安装、备份分支创建、origin 恢复与安全强推；输出对齐与回滚提示。
# - 供仓库管理员或维护者执行，需事先与团队沟通并临时放宽主分支保护策略。
# 风险说明：历史重写为破坏性操作；推送后所有协作者需重新克隆或硬重置对齐新历史。

TARGET_BRANCH="${TARGET_BRANCH:-master}"
STRIP_THRESHOLD="${STRIP_THRESHOLD:-100M}"

# 明确要清除的具体文件路径（逐行）
read -r -d '' SENSITIVE_PATHS <<'EOF'
tea-api/configs/config.mysql.local.yaml
pwd.bmiwj
wx-fe/pwd2.bmiwj
EOF

# 需要整目录或 glob 模式移除的路径；可多次传入 --path-glob
GLOB_PATHS=(
  "tea-api/dist/**"
  "deleted_backups/*.tar.gz"
)

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$ROOT" ] || [ ! -d "$ROOT/.git" ]; then
  echo "错误：请在 Git 仓库根目录执行（包含 .git 目录）。"
  exit 1
fi
cd "$ROOT"

echo "目标分支：$TARGET_BRANCH"
echo "大文件阈值：$STRIP_THRESHOLD"
echo "将移除的路径（文件清单）："
echo "$SENSITIVE_PATHS"
echo "将移除的路径（glob）：${GLOB_PATHS[*]}"
echo
echo "强烈建议：已与团队确认维护窗口，且仓库管理策略允许对 $TARGET_BRANCH 强制推送。"
echo -n "输入 YES 确认继续："
read -r CONFIRM
if [ "$CONFIRM" != "YES" ]; then
  echo "已取消。"
  exit 1
fi

echo "[1/8] 同步远端并切换目标分支..."
git fetch origin
git checkout "$TARGET_BRANCH"

BACKUP="${TARGET_BRANCH}-backup-$(date +%Y%m%d%H%M%S)"
echo "[2/8] 创建本地备份分支：$BACKUP"
git branch "$BACKUP" || true

echo "[3/8] 准备 git-filter-repo..."
mkdir -p "$HOME/.local/bin"
export PATH="$HOME/.local/bin:$PATH"
if ! command -v git-filter-repo >/dev/null 2>&1; then
  echo "未检测到 git-filter-repo，尝试下载到 ~/.local/bin ..."
  curl -fsSL -o "$HOME/.local/bin/git-filter-repo" \
    https://raw.githubusercontent.com/newren/git-filter-repo/main/git-filter-repo
  chmod +x "$HOME/.local/bin/git-filter-repo"
fi

echo "[4/8] 记录 origin 远端 URL ..."
REMOTE_URL="$(git remote get-url origin 2>/dev/null || true)"

echo "[5/8] 写入待移除路径清单 ..."
PATHS_FILE="$(mktemp)"
printf "%s\n" "$SENSITIVE_PATHS" > "$PATHS_FILE"

echo "[6/8] 重写历史（剥离大文件与敏感路径）..."
args=(
  --force
  --strip-blobs-bigger-than "$STRIP_THRESHOLD"
  --invert-paths --paths-from-file "$PATHS_FILE"
)
for g in "${GLOB_PATHS[@]}"; do
  args+=(--path-glob "$g")
done

git filter-repo "${args[@]}"
rm -f "$PATHS_FILE"

echo "[7/8] 清理不可达对象 ..."
git reflog expire --expire=now --all || true
git gc --prune=now --aggressive || true

echo "[8/8] 恢复 origin 并安全强推 ..."
if ! git remote get-url origin >/dev/null 2>&1; then
  if [ -n "$REMOTE_URL" ]; then
    git remote add origin "$REMOTE_URL"
  else
    echo "警告：未能恢复 origin 远端，请手动添加后再推送。"
  fi
fi

echo "准备强制推送：origin/$TARGET_BRANCH （--force-with-lease）"
git push --force-with-lease origin "$TARGET_BRANCH"

cat <<"EONXT"

完成。
协作者对齐建议：
  git fetch origin
  git checkout master
  git reset --hard origin/master

如需回滚：切换到备份分支并覆盖推送：
  git checkout <BACKUP_BRANCH>
  git push --force-with-lease origin master

EONXT
