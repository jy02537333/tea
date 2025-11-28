# A7 门店管理（Stores）

**目标**：门店列表、详情、创建、编辑、删除，支持经纬度校验

**路由**：`#/stores`

**关键接口**：`GET /api/v1/stores`、`GET /api/v1/stores/:id`、`POST /api/v1/stores`、`PUT /api/v1/stores/:id`、`DELETE /api/v1/stores/:id`

**验收标准**：
- 列表与详情正确显示数据，经纬度在表单中验算范围
- 创建/编辑表单通过校验并保存

**估时**：1.5d
