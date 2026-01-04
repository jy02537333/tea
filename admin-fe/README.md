# 茶心阁 管理端（admin-fe）

回到索引： [项目根 README](../README.md)

> 本 README 补充后台联调最小指引；更多上下文参见主仓文档。

提示：管理端演示发货/完成/取消可直接使用统一 CLI `scripts/dev-order-cli.sh`（子命令：deliver/receive/cancel），详见项目根 README 的“端到端订单联调与开发者 CLI（演示）”。

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

## 门店财务与流水

- 入口位置：侧边栏“门店财务”（路由 `/store-finance`）。
- 页面功能：
	- “门店财务提现”卡片：展示钱包概要（总收入/总退款/总提现/可用余额），支持发起提现申请与查看分页提现记录；可按状态筛选并导出当前页 CSV。
	- “资金流水（支付/退款/提现）”表格卡片：按时间倒序聚合展示门店收支变动（收款/退款/提现），含方向（收入/支出）、类型、金额、手续费、支付方式、关联号、备注；支持类型与起止日期筛选、分页与导出当前筛选为 CSV。
- 权限要求：`store:wallet:view`（用于查询流水与导出），提现操作需相应业务权限。
- 后端接口：
	- 列表：`GET /api/v1/stores/:id/finance/transactions?start=&end=&page=&limit=&type=`
	- 导出：`GET /api/v1/stores/:id/finance/transactions/export?start=&end=&type=`

> 本地快速预览（无需登录）：构建后可通过脚本 `node tools/automation/admin-static-server.js` 启动静态服务器，浏览器访问 `http://127.0.0.1:10113/store-finance?tk=临时token`（仅用于预览 UI，实际数据需后端与权限就绪）。

## 订单操作与权限

- 发货：需要权限 `order:deliver`，在订单状态“已付款”显示。
- 完成：需要权限 `order:complete`，在订单状态“配送中”显示。
- 管理员取消：需要权限 `order:cancel`，在状态未完成前显示；提交前需填写“取消原因”。
- 退款相关：需要权限 `order:refund`，包含“标记退款中/确认退款完成/立即退款”；提交前需填写“原因”。
- 说明：敏感操作统一使用弹窗表单收集原因，提交成功后列表/详情自动刷新。

### 开发者工具

- 侧边栏新增“开发者工具”链接，指向仓库根 README 的脚本说明，便于查看 `scripts/dev-order-cli.sh` 等联调脚本用法。

## 最小 UI 测试（Playwright）

项目已内置极简 E2E 用例（无需专门后端桩，仅要求可访问的 dev-login 接口）。

```bash
# 在 admin-fe 目录执行
cd admin-fe

# 安装依赖（如未安装）
npm ci || npm install

# 运行所有 E2E 测试
npm run e2e

# 仅运行订单原因弹窗最小用例
npx playwright test tests/orders-reason-modal.spec.ts

# 可选：覆盖目标站点与后端地址
# ADMIN_FE_URL 缺省为 http://127.0.0.1:5173
# API_BASE 缺省为 http://127.0.0.1:9292
ADMIN_FE_URL=http://127.0.0.1:5173 \
API_BASE=http://127.0.0.1:9292 \
npx playwright test tests/orders-reason-modal.spec.ts
```

- 测试行为：登录管理员、进入“订单”页；若存在“管理员取消/退款”操作，确认后应弹出填写原因的弹窗；若当前无可操作数据，用例退化为页面可达性校验。
- 调试产物：截图与 Trace 存放于 `build-ci-logs/playwright/`。