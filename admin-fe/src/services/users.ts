export async function createUser(data: Partial<UserDetail>) {
  const res = await axios.post<ApiResponse<UserDetail>>('/api/v1/admin/users', data);
  return res.data.data;
}
import { ApiResponse } from './types';
import axios from './api';

export interface UserQuery {
  page?: number;
  limit?: number;
  user_id?: number;
}

export interface UserDetail {
  id: number;
  nickname?: string;
  avatar?: string;
  phone?: string;
  roles?: string[];
  created_at?: string;
}

export async function listUsers(params: UserQuery) {
  const res = await axios.get<ApiResponse<{ data: UserDetail[]; total: number; page: number; limit: number }>>('/api/v1/admin/users', { params });
  return res.data.data;
}

export async function getUserDetail(id: number) {
  const res = await axios.get<ApiResponse<UserDetail>>(`/api/v1/user/${id}`);
  return res.data.data;
}

export async function updateUser(id: number, data: Partial<UserDetail>) {
  const res = await axios.put<ApiResponse<UserDetail>>(`/api/v1/user/${id}`, data);
  return res.data.data;
}

export async function disableUser(id: number) {
  const res = await axios.post<ApiResponse>(`/api/v1/user/${id}/disable`);
  return res.data;
}
