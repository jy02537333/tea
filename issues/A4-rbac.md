# A4 RBAC 管理（Roles & Permissions）

**目标**：角色列表、权限树、角色权限分配、用户角色分配

**路由**：`#/rbac`

**关键接口**：
- `GET /api/v1/admin/rbac/roles`
- `GET /api/v1/admin/rbac/permissions`
- `GET /api/v1/admin/rbac/role-permissions?role_id=`
- `POST /api/v1/admin/rbac/role`、`POST /api/v1/admin/rbac/role/assign-permissions`
- `POST /api/v1/admin/rbac/user/assign-role`

**组件**：`Tree`（权限）、`List`（角色）、`Modal`（新建）

**验收标准**：
- 能列出角色/权限并对指定角色提交分配权限
- 在用户详情页成功调用 assign-role 接口后用户列表反映变更（或提示刷新）

**估时**：2d
