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

## 端到端订单演示与统一 CLI 快捷操作

- 简要步骤：
	- 后端：`bash ./run-tea-api.sh && curl -sS -i http://127.0.0.1:9292/api/v1/health`
	- 前端：`pnpm -C wx-fe run dev:h5`（或 `dev:weapp`），在小程序端走“加购→结算→下单（自动统一下单+支付模拟）→订单详情”。
	- 订单状态快捷操作（管理员令牌自动获取/复用）：

```bash
# 发货（状态→ 配送中 3）
bash scripts/dev-order-cli.sh deliver <ORDER_ID>

# 确认收货（状态→ 已完成 4）
bash scripts/dev-order-cli.sh receive <ORDER_ID>

# 取消订单（状态→ 已取消 5；支持原因）
bash scripts/dev-order-cli.sh cancel <ORDER_ID> -r "超时未支付"
```

- 说明：订单详情页已内置自动轮询/手动刷新，便于观察状态变化；脚本响应保存在 `build-ci-logs/`。
- 更多细节与说明：见项目根 [README](../README.md) 与 [START_LOCAL.md](../START_LOCAL.md) 的对应章节。

## 近期特性补充（门店详情）

- 门店详情页：证照、导航/拨号、门店商品映射已实现（最小版）。入口位于首页门店列表的“查看详情”，或门店列表页；详情页支持一键导航到地图、拨打门店电话、跳转查看本店商品（带 `store_id`）。

## 近期特性补充（门店财务与流水）

- 入口与导航：
	- 门店详情页与门店列表页均提供“查看财务流水”入口；无权限会弹出提示并不跳转。
	- 财务页顶部提供“返回门店详情”“回到门店列表”，底部也提供“回到门店列表”，适配长列表滚动。
- 权限与范围：
	- 使用门店级接口 `/api/v1/stores/:id/finance/transactions`，天然限定到指定门店。
	- 无 `store:finance` 权限时，仅提示不跳转；已具备权限时正常进入并加载数据。
- 筛选与导出：
	- 支持类型（收款/退款/提现）与日期范围筛选，分页浏览；“导出”仅发起后端请求，不在小程序端处理文件下载。
- 兜底逻辑：
	- 路由缺少 `store_id` 参数时，自动读取 `Taro.getStorageSync('current_store_id')` 作为当前门店使用；无法读取则提示“缺少门店信息，请从门店列表进入”。
- 空态与降级处理：
	- 若后端尚未提供财务路由（返回 404，例如 `/api/v1/stores/:id/finance/transactions` 或其 `export` 接口），前端在财务页展示“后端路由未提供”空态，并禁用导出与分页控件，避免误导用户；待后端升级部署后即可恢复正常功能。
- 代码位置：
	- 财务页页面：`src/pages/store-finance/index.tsx`
	- 门店详情入口：`src/pages/store-detail/index.tsx`
	- 门店列表入口与“设为当前门店”提示：`src/pages/stores/index.tsx`

> 使用建议：先在“门店列表”设定当前门店（无权限账号也可设置为前端上下文）；有财务权限账号可直接在“门店详情”或“门店列表”进入“财务流水”。缺少权限时将提示不可用。