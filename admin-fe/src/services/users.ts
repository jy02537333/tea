import { api, unwrap } from './api';

export interface AdminUser {
  id: number;
  nickname?: string;
  phone?: string;
  role?: string;
  status?: number;
  created_at?: string;
}

export interface AdminUserQuery {
  user_id?: number;
  limit?: number;
  page?: number;
}

export interface UpdateAdminUserPayload {
  nickname?: string;
  phone?: string;
  role?: string;
  status?: number;
}

export interface CreateAdminUserPayload {
  username: string;
  password: string;
  phone: string;
  nickname?: string;
  role?: string;
  status?: number;
}

export async function getAdminUsers(params: AdminUserQuery = {}) {
  const finalParams = { ...params };
  if (!finalParams.limit) {
    finalParams.limit = 200;
  }
  const res = await api.get('/api/v1/admin/users', { params: finalParams });
  return unwrap<AdminUser[]>(res);
}

export async function getUserPermissions(userId: number) {
  const res = await api.get('/api/v1/admin/rbac/user-permissions', { params: { user_id: userId } });
  return unwrap<string[]>(res);
}

export async function createAdminUser(payload: CreateAdminUserPayload) {
  const res = await api.post('/api/v1/admin/users', payload);
  return unwrap<AdminUser>(res);
}

export async function updateAdminUser(userId: number, payload: UpdateAdminUserPayload) {
  const res = await api.put(`/api/v1/admin/users/${userId}`, payload);
  return unwrap<AdminUser>(res);
}

export async function resetAdminUserPassword(userId: number, newPassword: string) {
  const res = await api.post(`/api/v1/admin/users/${userId}/reset-password`, {
    new_password: newPassword,
  });
  return unwrap<{ message: string }>(res);
}
