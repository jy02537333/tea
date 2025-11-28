import api, { unwrapResponse } from './api';
import { Category } from './types';

export async function getCategories(params?: { parent_id?: number; status?: number }) {
  const res = await api.get('/api/v1/categories', { params });
  return unwrapResponse<Category[]>(res);
}
