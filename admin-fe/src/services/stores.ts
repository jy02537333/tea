import api, { unwrapResponse } from './api';
import { Store, PaginatedResponse, StoreOrderStats } from './types';

export async function listStores(params: { page?: number; limit?: number; status?: number; lat?: number; lng?: number } = {}): Promise<PaginatedResponse<Store>> {
  const res = await api.get('/api/v1/stores', { params });
  return unwrapResponse<PaginatedResponse<Store>>(res);
}

export async function getStore(id: number): Promise<Store> {
  const res = await api.get(`/api/v1/stores/${id}`);
  return unwrapResponse<Store>(res);
}

export async function createStore(payload: Partial<Store>): Promise<Store> {
  const res = await api.post('/api/v1/stores', payload);
  return unwrapResponse<Store>(res);
}

export async function updateStore(id: number, payload: Partial<Store>): Promise<void> {
  const res = await api.put(`/api/v1/stores/${id}`, payload);
  return unwrapResponse<void>(res);
}

export async function deleteStore(id: number): Promise<void> {
  const res = await api.delete(`/api/v1/stores/${id}`);
  return unwrapResponse<void>(res);
}

export async function getStoreOrderStats(id: number): Promise<StoreOrderStats> {
  const res = await api.get(`/api/v1/admin/stores/${id}/orders/stats`);
  return unwrapResponse<StoreOrderStats>(res);
}
