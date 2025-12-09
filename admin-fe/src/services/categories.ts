import { api, unwrap } from './api';

export interface Category {
  id: number;
  name: string;
  description?: string;
  sort?: number;
  parent_id?: number;
  image?: string;
  status?: number;
  created_at?: string;
  updated_at?: string;
}

export interface CategoryQueryParams {
  parent_id?: number;
  status?: number;
}

export interface CategoryPayload {
  name: string;
  description?: string;
  sort?: number;
  parent_id?: number;
  image?: string;
  status?: number;
}

export async function getCategories(params: CategoryQueryParams = {}) {
  const res = await api.get('/api/v1/categories', { params });
  return unwrap<Category[]>(res);
}

export async function createCategory(payload: CategoryPayload) {
  const res = await api.post('/api/v1/categories', payload);
  return unwrap(res);
}

export async function updateCategory(id: number, payload: CategoryPayload) {
  const res = await api.put(`/api/v1/categories/${id}`, payload);
  return unwrap(res);
}

export async function deleteCategory(id: number) {
  const res = await api.delete(`/api/v1/categories/${id}`);
  return unwrap(res);
}
