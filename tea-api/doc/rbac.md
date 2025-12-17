# RBAC 权限与缓存说明

本服务的权限控制采用“角色-权限”模型：
- 用户通过 UserRole 关联到角色
- 角色通过 RolePermission 关联到权限（Permission.name 如 accrual:run）
- 中间件按请求要求的权限校验；admin 角色跳过

为减少 DB 压力，用户的权限集合支持缓存：
- 优先使用 Redis（键：perm:user:<user_id>，TTL=30m）
- 无 Redis 时降级为进程内缓存（仅单进程有效）
- 缓存一致性策略：当成功命中缓存时，不再回落 DB 直查；因此“新授权”不立即生效，直到缓存失效
- 失效方式：
  - 手动：POST /api/v1/admin/rbac/cache/invalidate {user_id}
  - 自动：当执行 RBAC 变更（授予/撤销权限、授予/撤销角色）后，服务自动失效相关用户缓存

## 只读接口（需 rbac:view 或 admin）
- GET /api/v1/admin/rbac/roles
- GET /api/v1/admin/rbac/permissions
- GET /api/v1/admin/rbac/role-permissions?role_id=1
- GET /api/v1/admin/rbac/user-permissions?user_id=1

## 变更接口（需 rbac:manage 或 admin）
- POST /api/v1/admin/rbac/role {name, display_name}
- DELETE /api/v1/admin/rbac/role/:id
- POST /api/v1/admin/rbac/permission {name, display_name, module, action, resource}
- POST /api/v1/admin/rbac/role/assign-permission {role_id, permission_id}
- POST /api/v1/admin/rbac/role/revoke-permission {role_id, permission_id}
- POST /api/v1/admin/rbac/user/assign-role {user_id, role_id}
- POST /api/v1/admin/rbac/user/revoke-role {user_id, role_id}
- POST /api/v1/admin/rbac/cache/invalidate {user_id}

以上变更接口会触发缓存自动失效：
- 赋予/撤销角色权限：失效拥有该角色的所有用户缓存
- 赋予/撤销用户角色：失效该用户缓存

## 示例流程：授予后自动生效
1. 用户登录并尝试访问需要 accrual:run 的接口，得到 403
2. 管理员查询该用户权限（构建缓存）
3. 管理员授予“审计员”角色 accrual:run 权限
4. 用户再次访问，立即 200（因为步骤3自动失效了缓存）

可直接运行 `scripts/e2e_rbac_auto_invalidate.ps1` 演示完整流程。

## 一键初始化：创建审计员并赋权

脚本：`scripts/init_rbac_auditor.ps1`

用途：
- dev-login 创建指定 openid 的用户
- 运行 `cmd/seed_rbac` 写入基础角色/权限
- 为该用户赋予 auditor 角色
- 用管理员校验用户权限（应至少含 rbac:view 与 accrual:summary）

示例（已启动服务）：

```powershell
./scripts/init_rbac_auditor.ps1 -BaseUrl http://localhost:9292 -OpenId my_auditor_openid
```

示例（脚本内启动服务并初始化）：

```powershell
./scripts/init_rbac_auditor.ps1 -StartServer -OpenId my_auditor_openid
```
