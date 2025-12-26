import api, { unwrapPagination, unwrapResponse } from './api';

export interface AdminWithdrawalRow {
  id: number;
  withdraw_no: string;
  user_id: number;
  amount: string;
  fee: string;
  actual_amount: string;
  status: string;
  requested_at: string;
  processed_at?: string | null;
}

export async function listAdminWithdrawals(params: { page?: number; limit?: number; user_id?: string; withdraw_no?: string; status?: string } = {}) {
  const res = await api.get('/api/v1/admin/withdraws', { params });
  return unwrapPagination<AdminWithdrawalRow>(res);
}

export async function approveWithdrawal(id: number, payload: { remark?: string } = {}) {
  const res = await api.post(`/api/v1/admin/withdraws/${id}/approve`, payload);
  return unwrapResponse<any>(res);
}

export async function rejectWithdrawal(id: number, payload: { remark?: string } = {}) {
  const res = await api.post(`/api/v1/admin/withdraws/${id}/reject`, payload);
  return unwrapResponse<any>(res);
}
