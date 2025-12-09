import { api, unwrap } from './api';

export interface Role {
  id: number;
  name: string;
  display_name?: string;
}

export interface Permission {
  id: number;
  name: string;
  display_name?: string;
  module?: string;
  action?: string;
  resource?: string;
}

export interface RolePermission {
  id: number;
  role_id: number;
  permission_id: number;
}

export async function getRoles() {
  const res = await api.get('/api/v1/admin/rbac/roles');
  return unwrap<Role[]>(res);
}

export async function getPermissions() {
  const res = await api.get('/api/v1/admin/rbac/permissions');
  return unwrap<Permission[]>(res);
}

export async function getRolePermissions(roleId: number) {
  const res = await api.get('/api/v1/admin/rbac/role-permissions', { params: { role_id: roleId } });
  return unwrap<RolePermission[]>(res);
}

export async function assignPermissions(roleId: number, permissionIds: number[]) {
  const res = await api.post('/api/v1/admin/rbac/role/assign-permissions', { role_id: roleId, permission_ids: permissionIds });
  return unwrap(res);
}

export async function revokePermission(roleId: number, permissionId: number) {
  const res = await api.post('/api/v1/admin/rbac/role/revoke-permission', { role_id: roleId, permission_id: permissionId });
  return unwrap(res);
}

export async function assignRoleToUser(userId: number, roleId: number) {
  const res = await api.post('/api/v1/admin/rbac/user/assign-role', { user_id: userId, role_id: roleId });
  return unwrap(res);
}

export async function revokeRoleFromUser(userId: number, roleId: number) {
  const res = await api.post('/api/v1/admin/rbac/user/revoke-role', { user_id: userId, role_id: roleId });
  return unwrap(res);
}
