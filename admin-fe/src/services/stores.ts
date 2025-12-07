import { api, PaginatedResult, unwrap, unwrapPagination } from './api';

export interface Store {
  id: number;
  name: string;
  address?: string;
  phone?: string;
  latitude?: number;
  longitude?: number;
  business_hours?: string;
  images?: string;
  status?: number;
  created_at?: string;
  updated_at?: string;
}

export interface StoreListParams {
  page?: number;
  limit?: number;
  status?: number;
}

export interface StorePayload {
  name: string;
  address?: string;
  phone?: string;
  latitude?: number;
  longitude?: number;
  business_hours?: string;
  images?: string;
  status?: number;
}

export interface StoreOrderStats {
  store_id: number;
  total_orders: number;
  completed_amount: string;
  status_counts: { status: number; count: number }[];
}

export async function getStores(params: StoreListParams): Promise<PaginatedResult<Store>> {
  const res = await api.get('/api/v1/stores', { params });
  return unwrapPagination<Store>(res);
}

export async function getStore(id: number) {
  const res = await api.get(`/api/v1/stores/${id}`);
  return unwrap<Store>(res);
}

export async function createStore(payload: StorePayload) {
  const res = await api.post('/api/v1/stores', payload);
  return unwrap<Store>(res);
}

export async function updateStore(id: number, payload: StorePayload) {
  const res = await api.put(`/api/v1/stores/${id}`, payload);
  return unwrap(res);
}

export async function deleteStore(id: number) {
  const res = await api.delete(`/api/v1/stores/${id}`);
  return unwrap(res);
}

export async function getStoreOrderStats(id: number) {
  const res = await api.get(`/api/v1/admin/stores/${id}/orders/stats`);
  return unwrap<StoreOrderStats>(res);
}
