import { api, unwrap } from './api';

export interface AccrualSummaryResponse {
  record_count: number;
  user_count: number;
  total_interest: string;
  today_orders?: number;
}

export interface RunAccrualResponse {
  updated: number;
}

export async function getAccrualSummary(params: { start: string; end: string }) {
  const res = await api.get('/api/v1/admin/accrual/summary', { params });
  return unwrap<AccrualSummaryResponse>(res);
}

export async function runAccrual(payload: { date?: string; rate?: number }) {
  const res = await api.post('/api/v1/admin/accrual/run', payload);
  return unwrap<RunAccrualResponse>(res);
}

export async function exportAccrual(params: { start: string; end: string; format?: string }) {
  const res = await api.get('/api/v1/admin/accrual/export', {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}
