# Master History Rewrite Maintenance Plan

本计划用于在维护窗口内对 `master` 分支进行一次性历史重写，以移除历史中的超大文件与敏感信息，解除 GitHub 限制并提升仓库卫生与安全。

## 目标与范围
- 目标：
  - 删除历史中超过 100MB 的大文件 blob（符合 GitHub 限制）。
  - 删除已知的敏感文件/目录历史记录（避免 Push Protection 阻断）。
- 初始范围（可按需增减）：
  - 超过 100MB 的所有 blob
  - `tea-api/dist/**`（构建产物）
  - `deleted_backups/*.tar.gz`（备份归档）
  - `tea-api/configs/config.mysql.local.yaml`（本地配置）
  - `*.bmiwj`、`pwd.bmiwj`、`wx-fe/pwd2.bmiwj`（示例敏感扩展/文件）

## 风险与影响
- 历史重写会改变 `master` 的提交图谱：
  - 已存在的 fork、克隆、未合并分支、打开的 PR 需重新对齐新历史；
  - 依赖旧对象（提交/树/blob）的引用可能失效；
  - 需要临时允许对 `master` 的强制推送（force-with-lease）。

## 维护窗口准备
- 与团队确认维护时间（建议 30–60 分钟），期间暂缓合并与 push。
- 记录 `origin` 远端地址；确认拥有 `master` 强推权限（可暂时放宽保护）。
- 操作机需具备：`git 2.30+`、`curl`、`bash`；安装或可下载 `git-filter-repo`。
- 通知模版见文末“沟通模版”。

## 执行步骤（操作者本地）
> 以下命令为 Linux/macOS bash；Windows 可用 Git Bash。请在仓库根目录执行。

```bash
set -euo pipefail

BRANCH=master
BACKUP="${BRANCH}-backup-$(date +%Y%m%d%H%M%S)"

# 1) 准备本地 master（与远端对齐）
git fetch origin
# 以远端 origin/master 为基准，覆盖本地 master
git checkout -B "$BRANCH" origin/master

# 2) 备份分支（本地保留旧历史）
git branch "$BACKUP"

echo "Backup branch created: $BACKUP"

# 3) 准备 git-filter-repo（若未安装则下载脚本到本地 bin）
mkdir -p "$HOME/.local/bin"
export PATH="$HOME/.local/bin:$PATH"
if ! command -v git-filter-repo >/dev/null 2>&1; then
  echo "Installing git-filter-repo into ~/.local/bin ..."
  curl -fsSL -o "$HOME/.local/bin/git-filter-repo" \
    https://raw.githubusercontent.com/newren/git-filter-repo/main/git-filter-repo
  chmod +x "$HOME/.local/bin/git-filter-repo"
fi

# 4) 记录远端（git-filter-repo 会移除 origin）
REMOTE_URL=$(git remote get-url origin)

# 5) 定义需要移除的路径清单
PATHS_FILE=$(mktemp)
cat > "$PATHS_FILE" <<'EOF'
tea-api/configs/config.mysql.local.yaml
pwd.bmiwj
wx-fe/pwd2.bmiwj
EOF

# 6) 历史重写（剥离 >100MB，移除指定路径/目录）
# 说明：--invert-paths + --paths-from-file 删除清单；--path-glob 删除整目录；
#       --strip-blobs-bigger-than 100M 删除所有超过 100MB 的 blob。

git filter-repo --force \
  --strip-blobs-bigger-than 100M \
  --invert-paths --paths-from-file "$PATHS_FILE" \
  --path-glob 'tea-api/dist/**' \
  --path-glob 'deleted_backups/*.tar.gz'

rm -f "$PATHS_FILE"

# 7) 清理与打包（加速）
git reflog expire --expire=now --all || true
git gc --prune=now --aggressive || true

# 8) 恢复 origin 并安全强推到远端
if ! git remote get-url origin >/dev/null 2>&1; then
  git remote add origin "$REMOTE_URL"
fi

echo "Pushing rewritten history to origin/$BRANCH with --force-with-lease ..."
# 注意：确保维护窗口内执行
git push --force-with-lease origin "$BRANCH"

echo "Done. Rewritten history pushed."
```

## 验证清单
- GitHub Push Protection 未再阻断（无超大文件、无敏感文件提示）。
- 基本 CI 构建通过（本计划不涉及业务代码变更）。
- 仓库大小显著下降（可选）。

## 协作者迁移指引
- 建议重新克隆：
```bash
git clone https://github.com/<owner>/<repo>.git
```
- 或在现有克隆中硬重置：
```bash
git fetch origin
git checkout master
git reset --hard origin/master
```
- 若本地有进行中的分支：
  - 先备份补丁：`git format-patch origin/master..my/feature -o /tmp/patches`；
  - 重置对齐后再 `git am /tmp/patches/*.patch` 重新应用。

## 回滚预案
- 本地立即回滚：
  - 将备份分支强推覆盖：
    ```bash
    git checkout "$BACKUP"
    git push --force-with-lease origin "$BRANCH"
    ```
- 远端已有新提交的复杂情形：
  - 与受影响分支/PR 负责人同步后评估回滚影响；必要时暂存新提交、先恢复旧历史，再按补丁方式回放。
- 维护窗口沟通同步后执行，尽量避免双重重写。

## 沟通模版

- 维护公告（提前 24 小时）：
  > 各位好，计划在 <日期 时间> 对 `master` 分支做一次历史清理（移除超过 100MB 的大文件与敏感信息），期间将临时允许强制推送。请在窗口期间避免向 `master` 推送/合并。如遇影响请联系 <负责人>。

- 维护开始：
  > 历史清理开始，预计 <X 分钟> 完成。期间请勿向 `master` 推送/合并。

- 维护完成：
  > 历史清理已完成。请执行以下操作对齐新历史：
  > ```bash
  > git fetch origin
  > git checkout master
  > git reset --hard origin/master
  > ```
  > 如需继续未完成的分支开发，可先导出补丁再应用（见计划文档“协作者迁移指引”）。

## 附录：辅助排查命令
- 查找特定 blob-id 对应路径：
```bash
git rev-list --all --objects | grep <blobid>
```
- 统计大 blob（示例 50MB+）：
```bash
git rev-list --all --objects | \
  git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | \
  awk '$1=="blob" && $3>50*1024*1024 {print $0}' | sort -k3 -n
```
- 在分支上只移除大 blob（不删路径）：
```bash
git filter-repo --force --strip-blobs-bigger-than 100M
```

---

如需我在维护窗口内代执行以上步骤，可进一步：
- 准备执行脚本与试运行报告；
- 出具团队通知邮件/群消息；
- 维护后验证与复盘报告。
