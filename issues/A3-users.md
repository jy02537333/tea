# A3 用户管理（Users）

**目标**：实现用户列表、搜索、详情查看与导出基础功能

**路由**：`#/users`

**关键接口**：`GET /api/v1/admin/users?page=&limit=&user_id=`、`GET /api/v1/user/:id`

**组件**：`DataTable`, `SearchFilter`, `Drawer`（详情）

**验收标准**：
- 列表可分页、按 `user_id` 精确查询
- 点击用户打开详情 Drawer，能显示 `GET /api/v1/user/:id` 返回的数据

**估时**：1.5d
