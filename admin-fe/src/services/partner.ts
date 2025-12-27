import api, { unwrapPagination, unwrapResponse } from './api';

export interface PartnerRow {
  id: number;
  nickname: string;
  phone: string;
  partner_level_id?: number | null;
  created_at: string;
}

export async function listPartners(params: { page?: number; limit?: number; q?: string; level?: number } = {}) {
  const res = await api.get('/api/v1/admin/partners', { params });
  return unwrapPagination<PartnerRow>(res);
}

export interface CommissionRow {
  id: number;
  commission_type: string;
  gross_amount: string;
  net_amount: string;
  status: string;
  created_at: string;
}

export async function listPartnerCommissions(id: number, params: { page?: number; limit?: number } = {}) {
  const res = await api.get(`/api/v1/admin/partners/${id}/commissions`, { params });
  return unwrapPagination<CommissionRow>(res);
}
