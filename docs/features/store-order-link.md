## 门店与订单联动（方向2）

### 目标

实现 Admin 后台中“门店面板 ↔ 订单列表 ↔ 订单操作区”的联动：在门店面板内可以直接查看该门店的订单、把某条订单带入订单操作区并展示详情；在财务/支付/日志视图中点击“查看门店/查看用户”能一键跳转并聚焦相关资源。

本设计作为 M4 的核心交付之一，目的是把已有的门店管理能力与订单管理能力做深度耦合，提升运维/运营排障效率。

### 范围（最小可交付）
- 门店面板内展示该门店的订单列表（分页、按时间区间/状态筛选）
- 列表行支持“带入订单操作区”按钮，点击后把订单ID填入操作区并拉取订单详情
- 从财务页/支付流水/操作日志的“查看门店”跳转到门店面板并自动选中该门店

### API 设计（示例）

1) 获取门店订单列表
- Method: GET
- Path: /api/v1/admin/stores/{store_id}/orders
- Query:
  - page (int)
  - page_size (int)
  - status (optional, string 或 int)
  - start_time, end_time (optional, RFC3339)
  - order_id (optional)
- Response:
  - { total, page, page_size, data: [ { order_id, status, total_amount, paid_at, buyer_id, items: [...] , ... } ] }

2) 获取单条订单详情（已有接口，若无请新增）
- Method: GET
- Path: /api/v1/admin/orders/{order_id}
- Response: { order object with items, payments, logs }

3) 跳转辅助（前端行为）
- 在财务页中点击“查看门店”时，前端构造 URL: `/admin.html#stores?focus={store_id}` 或通过已有的标签页切换 helper `goToStore(store_id)`，保证前端切换并调用后端 `/stores/{id}/orders` 接口。

### 数据模型变更（若需）
- 无需新增表，复用 `orders` 表并在查询层加 `WHERE store_id = ?`。
- 若当前订单模型没有 `store_id` 字段，需要新增该字段并补充历史数据迁移脚本（见迁移计划）。

### 后端实现要点
- 在 `tea-api` 中新增 handler `GetStoreOrders(ctx)`，实现分页、筛选与权限校验（仅超级管理员或该门店管理员可查看对应门店数据）。
- 确保返回的订单摘要字段足够列表展示（order_id, status, total_amount, paid_at, buyer_id, item_count）。
- 实现合理的索引（`store_id`, `paid_at`）以保证查询性能。

### 前端实现要点（Admin-FE）
- 在门店面板添加“门店订单”子视图（表格 + 筛选控件）。
- 列表每行提供“带入操作区”按钮，点击后触发 `focusOrder(orderId)`：
  - 切换或显示“订单操作区”面板
  - 把 orderId 填入输入框并调用 `/api/v1/admin/orders/{order_id}` 拉取详情
- 在从财务/流水/日志页跳转场景中，提供 `goToStore(storeId)` helper 做标签切换并自动触发门店面板加载

### 权限与安全
- 所有接口均需校验 JWT 与用户角色/门店数据权限
- 前端仅根据返回的菜单/permissions 控制可见性，所有关键接口仍由后端二次鉴权

### 验收标准
- 门店面板的订单列表可在 2s 内返回首屏（1000 条数据内返回分页首页）
- 点击“带入操作区”后 1s 内能展示订单基本信息（网络+渲染）
- 在样例数据中（随机抽样 50 条）门店订单列表与 API-Server 返回一致率 100%

### 迁移/回滚计划（如需新增 store_id 字段）
- 新增字段脚本示例（MySQL）:
```sql
ALTER TABLE orders ADD COLUMN store_id BIGINT NULL DEFAULT NULL;
CREATE INDEX idx_orders_store_paid ON orders(store_id, paid_at);
```
- 回滚：删除字段并回滚索引（请先备份）

### 测试计划
- 单元测试：后端 handler 的参数解析、权限校验与分页逻辑
- 集成测试：在测试 DB 中创建门店/订单数据，调用 `/stores/{id}/orders` 验证返回结果
- E2E：用 Playwright 脚本模拟在 Admin-FE 中跳转到门店并带入订单，断言操作区展示正确订单ID与部分字段

### E2E 联调用例（示例）

> 说明：以下用例既可手工执行，也可作为 Playwright/类似工具脚本的步骤参考。

1. 门店订单列表 → 订单操作区
   - 前置条件：
     - 已存在门店 A（有有效的 `store_id`）。
     - 门店 A 下至少有一笔订单 O，状态为「已付款」或更后状态。
   - 步骤：
     1. 管理员登录 Admin 后台，进入「门店管理 → 门店列表」。
     2. 在门店列表中找到门店 A，点击操作列中的「订单」进入「门店订单列表」。
     3. 在门店订单列表中，找到订单 O 所在行，点击「在订单操作区打开」。
   - 期望结果：
     - 浏览器跳转到 `/orders?orderId={O.id}&storeId={A.id}`。
     - 全局「订单管理」页自动按门店 A 过滤（筛选表单中的门店 ID 已回填为 A.id）。
     - 页面右侧订单详情抽屉自动打开，展示订单 O 的关键信息（订单号、门店ID、用户ID、金额、状态等）。

2. 刷新保持联动状态
   - 前置条件：沿用用例 1 的最后状态（已在 `/orders?orderId={O.id}&storeId={A.id}` 页面，右侧抽屉已打开）。
   - 步骤：
     1. 直接刷新当前浏览器标签页。
   - 期望结果：
     - 刷新后仍停留在 `/orders?orderId={O.id}&storeId={A.id}`。
     - 订单列表继续按门店 A 过滤。
     - 右侧订单详情抽屉自动打开，展示订单 O 信息（无需再次手工点击）。

3. 仅按门店过滤（不指定订单）
   - 前置条件：存在门店 B，且其下有多笔订单。
   - 步骤：
     1. 直接访问 `/orders?storeId={B.id}`。
   - 期望结果：
     - 订单列表按门店 B 过滤；
     - 筛选表单中的门店 ID 回填为 B.id；
     - 不自动打开任何订单详情抽屉（保持待选状态）。

### 开发步骤（建议优先级）
1. 编写并提交本设计文档（已完成）
2. 后端：实现 `GetStoreOrders` handler + 路由 + 单元测试
3. DB：如需，添加 `store_id` 字段并补充索引与迁移脚本
4. 前端：在门店面板添加“门店订单”子视图与“带入”按钮
5. 集成测试与 E2E（Playwright）验证
6. 文档与运维说明补充（SOP）

---

文档作者: 开发小组
状态: 部分完成 / 持续迭代中
