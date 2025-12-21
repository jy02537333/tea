## PR Checklist

- 描述清晰：本次变更的目的、范围、影响已在正文说明
- 已读 PRD/相关文档：确保与 `doc/`、`docs/` 的规则一致
- 最小联调：如需验证，请参考下方“快速联调”说明

## 快速联调（Minimal Integration）
- 文档说明：docs/ci/minimal-integration.md
- 本地一键：`make run-min-integration`
- 产物位置：`build-ci-logs/**`（摘要：`build-ci-logs/local_api_summary.txt`）

### CI 提示
- 工作流：.github/workflows/minimal-integration.yml（push/PR 自动执行）
- 临时开启严格模式（可选）：参考 docs/ci/minimal-integration.md 中“临时开启严格模式（CI）”，可在 Actions 页面通过 “Run workflow” 输入 `STRICT_MIN=1`，或使用 `gh workflow run minimal-integration.yml -f STRICT_MIN=1`。

> 备注：严格模式仅用于更严格的断言观测，失败任务仍会保留并上传 `build-ci-logs/**` 以便排查。
