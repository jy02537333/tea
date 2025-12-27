## 部署与集成附录

### CI 工件上传（OSS）

- 使用已提供脚本 `scripts/upload_artifacts_to_oss.sh` 可将本地生成的 Playwright 截图与 trace 上传到 OSS。
- 需提供 `ADMIN_TOKEN`（或在 `build-ci-logs/admin_token.txt` 中放置 token）、`API_BASE`（默认 `http://127.0.0.1:9292`）。
- 示例：

```bash
export API_BASE="http://127.0.0.1:9292"
export ADMIN_TOKEN="$(cat build-ci-logs/admin_token.txt)"
bash scripts/upload_artifacts_to_oss.sh \
	build-ci-logs/playwright/partner-withdrawal-before-click.png \
	build-ci-logs/playwright/partner-withdrawal-modal.png \
	build-ci-logs/playwright/partner-withdrawal-trace.zip \
	build-ci-logs/playwright/partner-withdrawal-trace-report.html
```

脚本依赖后端 `POST /api/v1/admin/storage/oss/policy` 生成表单直传策略；建议将 CI 工件前缀设置为 `ci_artifact/` 并配置 30–90 天自动过期策略。

### 显式数据库迁移（替代 AutoMigrate）

- 在 `db/migrations/` 新增了 `20251227_add_wallet_and_withdrawals.sql`，包含 `wallets`、`wallet_transactions`、`withdrawal_requests` 的 DDL。
- 部署时建议关闭 AutoMigrate（`TEA_AUTO_MIGRATE=0`），并用显式 SQL 迁移工具（如 `mysql` CLI、`migrate`）执行上述文件。

### CI 草案（GitHub Actions）

- 在 `.github/workflows/e2e-artifacts.yml` 添加了一个 `workflow_dispatch` 工作流，用于占位运行 Playwright 并上传工件到 GitHub Artifacts；后续可在配置好后端与密钥后开启 `pull_request` 触发并接入 OSS 上传。

