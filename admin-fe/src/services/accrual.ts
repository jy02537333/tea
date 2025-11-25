import api, { unwrapResponse } from './api';
import { AccrualSummary } from './types';

export async function runAccrual(payload: { date?: string; rate?: number; user_id?: number } = {}): Promise<{ updated: number }> {
  const res = await api.post('/api/v1/admin/accrual/run', payload);
  return unwrapResponse<{ updated: number }>(res);
}

export async function accrualSummary(params: { start?: string; end?: string } = {}): Promise<AccrualSummary> {
  const res = await api.get('/api/v1/admin/accrual/summary', { params });
  return unwrapResponse<AccrualSummary>(res);
}

export async function accrualExport(params: { start?: string; end?: string; format?: string; lang?: string; fields?: string; zip?: number } = {}) {
  const res = await api.get('/api/v1/admin/accrual/export', { params, responseType: 'blob' });
  return res.data; // blob
}
