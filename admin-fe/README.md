# 茶心阁 管理端（admin-fe）

回到索引： [项目根 README](../README.md)

> 本 README 补充后台联调最小指引；更多上下文参见主仓文档。

## 联调快速指南（后端最小联调一键跑通）

在仓库根目录执行（确保本机或容器内 `tea-api` 可访问，默认 `http://127.0.0.1:9292`）：

```bash
# 生成 ADMIN_TOKEN / USER_TOKEN，并写入 build-ci-logs/tokens.env
make prepare-tokens

# 运行 Sprint C 最小联调（含 OSS policy、佣金、门店打印等样例）
make run-min-integration

# 可选：加载导出的令牌到当前 Shell 会话
source build-ci-logs/tokens.env
```

- 产物位置：`build-ci-logs/**`（如 `get_oss_policy.json`、`commissions_*`、`store_order_*` 等）。
- 服务地址：可通过 `API_BASE` 覆盖默认地址，例如：`export API_BASE="http://127.0.0.1:9292"`。
- 详细说明：见文档 [doc/prd_sprints.md](../doc/prd_sprints.md) 中“Sprint C — 最小联调与证据”。

安全提示：`build-ci-logs/tokens.env` 含登录令牌，仅用于本地联调，勿提交到仓库或外泄。

## 启动本地开发

```bash
npm install
npm run dev
```

- 开发阶段调用后端接口时，确保后端可达（或在 `.env`/运行时配置中设置 API 基地址）。
- 与 OSS 直传相关的管理端功能依赖后端接口 `POST /api/v1/admin/storage/oss/policy`，可先通过“联调快速指南”生成最小证据进行自检。

### API_BASE 配置（.env）

- 推荐通过 Vite 环境变量指定后端地址：

	1) 复制示例文件：

	```bash
	cp .env.example .env
	```

	2) 根据实际后端地址修改 `.env`：

	```bash
	# .env
	VITE_API_BASE=http://127.0.0.1:9292
	```

- 代码侧可通过 `import.meta.env.VITE_API_BASE` 读取；命令行临时覆盖也可：

```bash
VITE_API_BASE="http://127.0.0.1:9393" npm run dev
```

## 贡献指南（前端）

为提升一致性与可维护性，请在提交 PR 前自检以下要点：

- 使用统一的用户聚合接口：登录后从 `GET /api/v1/users/me/summary` 获取聚合信息，避免直接依赖历史接口（`/api/v1/user/info`）。
- 提现 remark 解析与导出统一使用工具：复用 [src/utils/withdraw.ts](./src/utils/withdraw.ts) 的方法：
	- 页面展示：`getRemarkField(remark, key)`；表格列：`...buildWithdrawRemarkColumns()`。
	- 导出 CSV：`getRemarkFieldsForCsv(remark)`。
	- 不要在页面中直接 `JSON.parse(record.remark)` 分散实现，避免重复与异常处理遗漏。
- 基本检查：在 `admin-fe` 目录下执行 `pnpm run typecheck` 与 `pnpm run build` 保证通过。

如涉及接口或字段调整，请同步更新 `doc/prd.md` 与 `doc/prd_sprints.md` 的约定说明。