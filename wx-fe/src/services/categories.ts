import api, { unwrapResponse } from './api';
import { Category, PaginationResponse } from './types';

export async function listCategories(params: { status?: number; parent_id?: number } = {}): Promise<PaginationResponse<Category>> {
  const res = await api.get('/api/v1/categories', { params });
  return unwrapResponse<PaginationResponse<Category>>(res);
}
