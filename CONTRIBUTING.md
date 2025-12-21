# CONTRIBUTING

## 目标
- 提供最小可执行联调路径，统一证据归档到 `build-ci-logs/`。
- 在 CI 中自动运行最小脚本，并支持可选严格模式（默认关闭）。

## 快速联调（Minimal Integration）
- 本地一键：
```bash
make run-min-integration
```
- 产物位置：`build-ci-logs/**`；摘要：`build-ci-logs/local_api_summary.txt`。
- 相关脚本：`scripts/run_admin_product_min.sh`、`scripts/run_commission_min.sh`、`scripts/local_api_check.sh`（可选）。

## CI 工作流
- 工作流文件：`.github/workflows/minimal-integration.yml`
- 触发：push/PR 时运行；内置 MySQL/Redis；上传 `build-ci-logs/**` 为 Actions Artifacts。

## 严格模式（Strict，可选）
- 行为：额外断言关键环节（创建数据、上传 URL、图片回填、佣金释放等），失败保留证据与摘要。
- 本地开启：
```bash
make run-min-integration-strict   # 等价 STRICT_MIN=1
```

## 临时开启严格模式（CI）
- UI：在 Actions 选择 `minimal-integration` 工作流 → `Run workflow` → 输入 `STRICT_MIN=1`。
- CLI（示例）：
```bash
# 使用 GitHub CLI 手动触发一次严格模式运行（需已安装 gh 并登录）
# 注意：工作流名称以仓库实际为准
gh workflow run minimal-integration.yml -f STRICT_MIN=1
```
- 说明：严格模式默认关闭，仅在需要时临时开启；失败会上传完整证据便于排查。

## 参考链接
- 最小集成与 CI 指南：`docs/ci/minimal-integration.md`（合并后在 master 可用）
- README 的“快速联调”小节：`README.md`

