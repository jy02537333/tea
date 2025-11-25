# 前端任务卡（Admin-FE 与 WX-FE）

说明：本文件把每个页面/模块拆为独立可分配的任务卡，包含目标、输入/输出、关键接口、验收标准与估时（粗略）。

---

## 目录
- Admin-FE
  - A1 登录（Admin Auth）
  - A2 仪表盘（Dashboard）
  - A3 用户管理（Users）
  - A4 RBAC 管理（Roles & Permissions）
  - A5 商品分类（Categories）
  - A6 商品管理（Products）
  - A7 门店管理（Stores）
  - A8 管理端订单（Admin Orders）
  - A9 报表 / 计提（Accruals / Exports）
  - A10 操作日志 / 访问日志（Logs）
  - A11 财务/退款/提现（Payments/Refunds/Withdraws）
- WX-FE
  - W1 首页（Index）
  - W2 分类 / 商品列表（Category / Products）
  - W3 商品详情（Product Detail）
  - W4 购物车（Cart）
  - W5 结算 / 下单（Checkout / Orders）
  - W6 我的 / 订单（Profile / Orders）
  - W7 优惠券（Coupons）
  - W8 门店列表 / 选择（Stores）
- 共享 / 基础设施
  - S1 API services（`services/*.ts` 模板）
  - S2 Auth & Token 管理（login, refresh, interceptor）
  - S3 通用 UI 组件（DataTable, CrudModal, Uploader, ProductCard）
  - S4 Mock & E2E 支持（MSW / Cypress）
  - S5 CI + Lint + Tests（PR 检查）

---

## Admin-FE 任务卡

### A1 登录（Admin Auth）
- 目标：实现管理员登录页并完成 token 存储、拦截器注入与自动跳转。
- 路由：`#/login`
- 关键接口：`POST /api/v1/user/login`、`POST /api/v1/user/dev-login`（dev）
- 输入：用户名/密码 或 openid（dev）
- 输出：保存 `token`，跳转到 `#/dashboard`（或 redirect 参数目标）
- 验收标准：
  - 成功登录后 `localStorage.token` 存在且 axios 拦截器在后续请求中使用
  - 失败时显示错误 Toast；401 自动跳回登录
- 粗略估时：0.5d

### A2 仪表盘（Dashboard）
- 目标：实现 KPI 卡片、主要汇总图表与快速操作按钮
- 路由：`#/dashboard`
- 关键接口：`GET /api/v1/admin/accrual/summary`、（可选）其它统计接口
- 组件：`Card`, `Statistic`, `Chart`（ECharts）
- 验收标准：
  - KPI 卡片显示来自 `accrual/summary` 的数值
  - 点击“触发计提”弹出确认并调用 `POST /api/v1/admin/accrual/run`，返回成功消息
- 粗略估时：1d

### A3 用户管理（Users）
- 目标：实现用户列表、搜索、详情查看与导出基础功能
- 路由：`#/users`
- 关键接口：`GET /api/v1/admin/users?page=&limit=&user_id=`、`GET /api/v1/user/:id`
- 组件：`DataTable`, `SearchFilter`, `Drawer`（详情）
- 验收标准：
  - 列表可分页、按 `user_id` 精确查询
  - 点击用户打开详情 Drawer，能显示 `GET /api/v1/user/:id` 返回的数据
- 粗略估时：1.5d

### A4 RBAC 管理（Roles & Permissions）
- 目标：角色列表、权限树、角色权限分配、用户角色分配
- 路由：`#/rbac`
- 关键接口：
  - `GET /api/v1/admin/rbac/roles`
  - `GET /api/v1/admin/rbac/permissions`
  - `GET /api/v1/admin/rbac/role-permissions?role_id=`
  - `POST /api/v1/admin/rbac/role`、`POST /api/v1/admin/rbac/role/assign-permissions`
  - `POST /api/v1/admin/rbac/user/assign-role`
- 组件：`Tree`（权限）、`List`（角色）、`Modal`（新建）
- 验收标准：
  - 能列出角色/权限并对指定角色提交分配权限
  - 在用户详情页成功调用 assign-role 接口后用户列表反映变更（或提示刷新）
- 粗略估时：2d

### A5 商品分类（Categories）
- 目标：完成分类树/列表、新建、编辑、删除
- 路由：`#/categories`
- 关键接口：`GET /api/v1/categories`、`POST /api/v1/categories`、`PUT /api/v1/categories/:id`、`DELETE /api/v1/categories/:id`
- 验收标准：
  - 新建/编辑后列表刷新并显示最新数据
  - 删除需二次确认并正确调用 DELETE
- 粗略估时：1d

### A6 商品管理（Products）
- 目标：实现商品列表/搜索/筛选、新建、编辑、删除、库存更新
- 路由：`#/products`
- 关键接口：`GET /api/v1/products`、`POST /api/v1/products`、`PUT /api/v1/products/:id`、`PUT /api/v1/products/:id/stock`、`DELETE /api/v1/products/:id`
- 组件：`DataTable`, `Uploader`, `CrudModal`, `RichText`（详情）
- 验收标准：
  - 支持分页/关键字搜索/分类筛选
  - 新建商品成功后在列表出现并能查看详情
  - 库存更新接口按 action 参数正确调用并刷新行数据
- 粗略估时：3d

### A7 门店管理（Stores）
- 目标：门店列表、详情、创建、编辑、删除，支持经纬度校验
- 路由：`#/stores`
- 关键接口：`GET /api/v1/stores`、`GET /api/v1/stores/:id`、`POST /api/v1/stores`、`PUT /api/v1/stores/:id`、`DELETE /api/v1/stores/:id`
- 验收标准：
  - 列表与详情正确显示数据，经纬度在表单中验算范围
  - 创建/编辑表单通过校验并保存
- 粗略估时：1.5d

### A8 管理端订单（Admin Orders）
- 目标：实现管理端订单查询/筛选/导出/详情/管理操作
- 路由：`#/admin/orders` 或 `#/orders`（管理员视图）
- 关键接口：`GET /api/v1/admin/orders`、`GET /api/v1/admin/orders/export`、`GET /api/v1/admin/orders/:id`、管理操作 POST `/api/v1/orders/:id/<action>`
- 验收标准：
  - 支持导出（CSV）并能下载
  - 在详情页面可执行管理员操作并且列表刷新
- 粗略估时：3d

### A9 报表 / 计提（Accruals / Exports）
- 目标：报表筛选、展示、导出与触发计提动作
- 路由：`#/accrual`
- 关键接口：`GET /api/v1/admin/accrual/summary`、`GET /api/v1/admin/accrual/export`、`POST /api/v1/admin/accrual/run`
- 验收标准：
  - 支持多种导出格式（CSV/XLSX/ZIP）并正确下载
  - 触发计提后给出 `updated` 数值提示
- 粗略估时：1.5d

### A10 操作日志 / 访问日志（Logs）
- 目标：实现日志筛选、查看、导出
- 路由：`#/logs`
- 关键接口：`GET /api/v1/admin/logs/operations`、`GET /api/v1/admin/logs/access`、`/export`
- 验收标准：分页/导出/筛选功能正常
- 粗略估时：1.5d

### A11 财务/退款/提现（Payments/Refunds/Withdraws）
- 目标：财务记录与提现审批工作流（列表/导出/审核动作）
- 路由：`#/finance`, `#/refunds`, `#/withdraws`
- 关键接口：`GET /api/v1/admin/payments`、`GET /api/v1/admin/refunds`、`GET /api/v1/admin/withdraws`、审核 POST `/api/v1/admin/withdraws/:id/approve|complete|reject`
- 验收标准：审批操作后状态变更显示并能导出记录
- 粗略估时：2d

---

## WX-FE 任务卡

### W1 首页（Index）
- 目标：实现搜索、banner、分类横向、推荐商品、附近门店
- 路由：`/pages/index`
- 关键接口：`GET /api/v1/categories`、`GET /api/v1/products`、`GET /api/v1/stores?lat=&lng=`
- 验收标准：界面与设计稿一致，商品点击进入详情，门店切换刷新商品
- 粗略估时：1.5d

### W2 分类 / 商品列表（Category / Products）
- 目标：实现筛选、排序、分页加载
- 路由：`/pages/products?category_id=&keyword=`
- 关键接口：`GET /api/v1/products?page=&limit=&category_id=&keyword=&store_id=`
- 验收标准：筛选与分页工作正常，空状态友好提示
- 粗略估时：1.5d

### W3 商品详情（Product Detail）
- 目标：图片轮播、规格选择、加入购物车、立即购买
- 路由：`/pages/product?id=`
- 关键接口：`GET /api/v1/products/:id`、`POST /api/v1/cart/items`
- 验收标准：加入购物车后 badge 更新，立即购买跳转 checkout 并预填商品
- 粗略估时：1.5d

### W4 购物车（Cart）
- 目标：购物车列表、编辑数量、删除、批量结算
- 路由：`/pages/cart`
- 关键接口：`GET /api/v1/cart`、`PUT /api/v1/cart/items/:id`、`DELETE /api/v1/cart/items/:id`、`DELETE /api/v1/cart/clear`
- 验收标准：数量修改实时更新、删除后刷新、结算跳转 checkout
- 粗略估时：1.5d

### W5 结算 / 下单（Checkout / Orders）
- 目标：下单流程（地址/自取、优惠券、下单、调用支付 intent）
- 路由：`/pages/checkout`
- 关键接口：`POST /api/v1/orders/from-cart`、`POST /api/v1/payment/intents`
- 验收标准：下单成功后跳转 order-detail 并展示支付入口或已支付状态
- 粗略估时：2d

### W6 我的 / 订单（Profile / Orders）
- 目标：个人信息、订单列表、订单详情与操作
- 路由：`/pages/profile`, `/pages/orders`, `/pages/order-detail?id=`
- 关键接口：`GET /api/v1/orders`、`GET /api/v1/orders/:id`、订单操作接口
- 验收标准：订单状态、取消、确认收货等操作有效
- 粗略估时：1.5d

### W7 优惠券（Coupons）
- 目标：可领取优惠券展示、领取动作
- 路由：`/pages/coupons`
- 关键接口：`GET /api/v1/coupons`、`POST /api/v1/coupons/grant`、`GET /api/v1/user/coupons`
- 验收标准：领取后在用户券列表中可见
- 粗略估时：1d

### W8 门店列表 / 选择（Stores）
- 目标：门店列表与选择，用于门店覆盖价/自取
- 路由：`/pages/stores`
- 关键接口：`GET /api/v1/stores`、`GET /api/v1/stores/:id`
- 验收标准：能按经纬度/附近筛选并切换当前门店
- 粗略估时：1d

---

## 共享 / 基础设施任务

### S1 API services（`services/*.ts` 模板）
- 目标：生成 TypeScript axios 实例和按资源分组的 service 模板（auth, products, cart, orders, rbac, stores, accrual）
- 产出：`admin-fe/src/services/*.ts` 与 `wx-fe/src/services/*.ts` 模板
- 验收标准：每个 service 示例包含至少 2 个 CRUD 调用与类型定义
- 粗略估时：2d

### S2 Auth & Token 管理
- 目标：实现 token 保存/刷新（refresh endpoint）、路由守卫、401 处理逻辑
- 产出：`auth.ts`、axios 拦截器、登录重定向逻辑
- 粗略估时：1d

### S3 通用 UI 组件
- 目标：实现 `DataTable`, `CrudModal`, `Uploader`, `ProductCard`, `PermissionTree`
- 验收标准：组件有文档/示例 Storybook，支持必要的 props 与事件
- 粗略估时：3d

### S4 Mock & E2E 支持
- 目标：配置 MSW 或 Mock Server，搭建 Cypress / Playwright E2E 基本用例
- 验收标准：覆盖关键流程（登录->下单->支付->订单确认）自动化测试能本地通过
- 粗略估时：3d

### S5 CI + Lint + Tests
- 目标：PR 流程触发 lint、unit tests；主分支触发构建并上传预览站点（如有）
- 粗略估时：1.5d

---

## 分配与优先级建议（Sprint 排序）
- Sprint 0（准备，2 天）：S1、S2、S3 基础服务与组件骨架 + Mock 环境（S4 初步）
- Sprint 1（admin 基本，1.5 周）：A1、A2、A3、A6（简化新建）
- Sprint 2（admin 完整，1.5 周）：A4、A5、A7、A8
- Sprint 3（wx 基本，1.5 周）：W1、W2、W3、W4
- Sprint 4（wx 完整 + 联调，1.5 周）：W5、W6、W7、W8 + S4 完善 + E2E

---

## 下一步（可选）
- 我可以基于这些任务自动生成每个任务的 GitHub Issue 模板（需要你的仓库权限），或
- 生成 `services/*.ts` 模板并把示例页面 scaffold 到项目中。

请告诉我接下来要我执行的项（生成 services / scaffold 页面 / 生成 Issue）。
