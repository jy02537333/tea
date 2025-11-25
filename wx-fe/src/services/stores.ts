import api, { unwrapResponse } from './api';
import { Store, PaginationResponse } from './types';

export async function listStores(params: any = {}): Promise<PaginationResponse<Store>> {
  const res = await api.get('/api/v1/stores', { params });
  return unwrapResponse<PaginationResponse<Store>>(res);
}

export async function getStore(id: number): Promise<Store> {
  const res = await api.get(`/api/v1/stores/${id}`);
  return unwrapResponse<Store>(res);
}
