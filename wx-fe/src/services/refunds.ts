import api, { unwrapResponse } from './api';
import type { PaginationResponse, Refund } from './types';

export async function listMyRefunds(params: { order_id?: number; status?: number; start?: string; end?: string; page?: number; limit?: number } = {}): Promise<PaginationResponse<Refund>> {
  const res = await api.get('/api/v1/refunds', { params });
  return unwrapResponse<PaginationResponse<Refund>>(res);
}
