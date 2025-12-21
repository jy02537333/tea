# Contributing Guide

感谢您的贡献！为便于协作与验证，这里提供最小联调与 CI 指南。

## 最小联调（Minimal Integration）
- 文档：docs/ci/minimal-integration.md
- 本地一键：

```bash
make run-min-integration
```

- 产物位置：`build-ci-logs/**`（摘要：`build-ci-logs/local_api_summary.txt`）

## CI 工作流
- 工作流文件：.github/workflows/minimal-integration.yml
- 触发方式：push 到功能分支或提交对 master 的 PR 时自动执行，产物上传为 Actions Artifacts。

### 临时开启严格模式（CI）
- UI 方式：Actions 页面选择 “Minimal Integration CI” → “Run workflow”，将输入框 `STRICT_MIN` 设为 `1`，选择需要的分支后运行。
- gh CLI 方式：

```bash
# 针对指定分支触发严格模式
gh workflow run minimal-integration.yml -r feat/withdraw-remark-json-ui-docs -f STRICT_MIN=1

# 针对某个 PR 的 head 分支触发（以 PR #50 为例）
BRANCH=$(gh pr view 50 --json headRefName --jq .headRefName)
gh workflow run minimal-integration.yml -r "$BRANCH" -f STRICT_MIN=1
```

- 行为说明：严格模式执行额外断言；失败时任务标红，但仍会上传 `build-ci-logs/**` 用于排查。

## 开发约定
- 与 `doc/`、`docs/` 的约束保持一致，更新相关文档与清单。
- 如涉及上传与商品/分销链路，请优先使用最小联调命令验证，并在 PR 描述中附上关键产物说明。
