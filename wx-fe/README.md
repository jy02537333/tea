# 茶心阁 小程序前端（wx-fe）

回到索引： [项目根 README](../README.md)

> 本 README 补充用户端联调最小指引；更多上下文参见主仓文档。

## 联调快速指南（后端最小联调一键跑通）

在仓库根目录执行（确保本机或容器内 `tea-api` 可访问，默认 `http://127.0.0.1:9292`）：

```bash
# 生成 ADMIN_TOKEN / USER_TOKEN，并写入 build-ci-logs/tokens.env
make prepare-tokens

# 运行 Sprint C 最小联调（生成推荐/佣金/会员升级等证据）
make run-min-integration

# 可选：加载导出的令牌到当前 Shell 会话
source build-ci-logs/tokens.env
```

- 产物位置：`build-ci-logs/**`（如 `referral_*`、`commissions_*`、`membership_upgrade_summary.json` 等）。
- 服务地址：可通过 `API_BASE` 覆盖默认地址，例如：`export API_BASE="http://127.0.0.1:9292"`。
- 详细说明：见文档 [doc/prd_sprints.md](../doc/prd_sprints.md) 中“Sprint C — 最小联调与证据”。

安全提示：`build-ci-logs/tokens.env` 含登录令牌，仅用于本地联调，勿提交到仓库或外泄。

## 启动本地开发

```bash
npm install
npm run dev
```

- 开发阶段调用后端接口时，确保后端可达（或在 `.env`/运行时配置中设置 API 基地址）。

### API_BASE 配置（.env）

- 本项目在 `config/index.js` 中通过 `defineConstants` 注入 `WX_API_BASE_URL`：

	1) 复制示例文件：

	```bash
	cp .env.example .env
	```

	2) 根据实际后端地址修改 `.env`：

	```bash
	# .env
	WX_API_BASE_URL=http://127.0.0.1:9292
	```

- 运行时即可在代码中使用 `WX_API_BASE_URL` 常量（经编译期注入）；命令行临时覆盖也可：

```bash
WX_API_BASE_URL="http://127.0.0.1:9393" npm run dev
```

## 贡献指南（小程序端前端）

为保持与管理端一致的契约与工具复用，提交前请自检：

- 统一聚合来源：登录后统一从 `GET /api/v1/users/me/summary` 获取用户聚合信息；避免直接依赖历史接口（`/api/v1/user/info`）。
- 提现 remark 解析：统一使用工具文件（建议路径 `src/utils/withdraw.ts`）的方法：
	- 展示：`getRemarkField(remark, key, fallback)`；
	- CSV/导出：`getRemarkFieldsForCsv(remark)`；
	- 约束：避免在页面中散落 `JSON.parse(record.remark)`；保持与管理端工具的签名一致，便于迁移与复用。
- 基本检查：执行类型检查/构建任务（例如 `npm run build` 或项目约定的 `typecheck`）确保通过。

如涉及接口或字段调整，请同步更新 `doc/prd.md` 与 `doc/prd_sprints.md` 的约定说明。