# A6 商品管理（Products）

**目标**：实现商品列表/搜索/筛选、新建、编辑、删除、库存更新

**路由**：`#/products`

**关键接口**：`GET /api/v1/products`、`POST /api/v1/products`、`PUT /api/v1/products/:id`、`PUT /api/v1/products/:id/stock`、`DELETE /api/v1/products/:id`

**组件**：`DataTable`, `Uploader`, `CrudModal`, `RichText`（详情）

**验收标准**：
- 支持分页/关键字搜索/分类筛选
- 新建商品成功后在列表出现并能查看详情
- 库存更新接口按 action 参数正确调用并刷新行数据

**估时**：3d
