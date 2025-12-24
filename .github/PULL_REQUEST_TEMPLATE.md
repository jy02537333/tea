# Pull Request

请填写本次变更的背景与目的，并勾选检查项以便评审。

## 变更说明
- 背景/问题：
- 解决方案：
- 影响范围：

## 检查清单（必读）
- [ ] 统一登录与用户聚合接口：使用 `POST /api/v1/auth/login` 与 `GET /api/v1/users/me/summary`，避免混用历史接口（`/api/v1/user/login`、`/api/v1/user/info`）。
- [ ] 前端提现 remark 解析统一使用工具：已复用 [admin-fe/src/utils/withdraw.ts](../admin-fe/src/utils/withdraw.ts) 提供的方法（`getRemarkField` / `getRemarkFieldsForCsv` / `buildWithdrawRemarkColumns`），不得在页面中直接 `JSON.parse(record.remark)` 散落实现。
- [ ] 管理端前端已通过最小验证：在 `admin-fe` 执行 `pnpm run typecheck` 与 `pnpm run build` 通过。
- [ ] 如涉及后端接口/字段变更，已同步更新 `doc/prd.md` / `doc/prd_sprints.md` / 相关 API 文档（`tea-api/docs/*.md`）。

## 证据与链接（可选）
- CI 产物链接：
- 关键页面截图：
- 相关任务/需求链接：
