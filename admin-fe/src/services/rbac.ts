import api, { unwrapResponse } from './api';
import { Role, Permission } from './types';

export async function listRoles(): Promise<Role[]> {
  const res = await api.get('/api/v1/admin/rbac/roles');
  return unwrapResponse<Role[]>(res);
}

export async function listPermissions(): Promise<Permission[]> {
  const res = await api.get('/api/v1/admin/rbac/permissions');
  return unwrapResponse<Permission[]>(res);
}

export async function getRolePermissions(role_id: number): Promise<Permission[]> {
  const res = await api.get('/api/v1/admin/rbac/role-permissions', { params: { role_id } });
  return unwrapResponse<Permission[]>(res);
}

export async function createRole(payload: { name: string; display_name?: string }): Promise<Role> {
  const res = await api.post('/api/v1/admin/rbac/role', payload);
  return unwrapResponse<Role>(res);
}

export async function assignPermissionsToRole(role_id: number, permission_ids: number[]): Promise<void> {
  const res = await api.post('/api/v1/admin/rbac/role/assign-permissions', { role_id, permission_ids });
  return unwrapResponse<void>(res);
}

export async function assignRoleToUser(user_id: number, role_id: number): Promise<void> {
  const res = await api.post('/api/v1/admin/rbac/user/assign-role', { user_id, role_id });
  return unwrapResponse<void>(res);
}
