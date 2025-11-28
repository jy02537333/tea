# A8 管理端订单（Admin Orders）

**目标**：实现管理端订单查询/筛选/导出/详情/管理操作

**路由**：`#/admin/orders` 或 `#/orders`（管理员视图）

**关键接口**：`GET /api/v1/admin/orders`、`GET /api/v1/admin/orders/export`、`GET /api/v1/admin/orders/:id`、管理操作 POST `/api/v1/orders/:id/<action>`

**验收标准**：
- 支持导出（CSV）并能下载
- 在详情页面可执行管理员操作并且列表刷新

**估时**：3d
