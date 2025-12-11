import { api, unwrapPagination } from './api';

export interface PaymentRecord {
  id: number;
  order_id: number;
  payment_no: string;
  amount: string;
  status: number;
  payment_method: number;
  created_at?: string;
  updated_at?: string;
  store_name?: string;
}

export interface RefundRecord {
  id: number;
  order_id: number;
  refund_no: string;
  refund_amount: string;
  status: number;
  created_at?: string;
  updated_at?: string;
  store_name?: string;
}

export interface FinanceRecordQuery {
  page?: number;
  limit?: number;
  order_id?: number;
  store_id?: number;
  payment_no?: string;
  refund_no?: string;
  status?: number | string;
  method?: number | string;
  start?: string;
  end?: string;
}

export async function listPaymentsRecords(params: FinanceRecordQuery) {
  const res = await api.get('/api/v1/admin/payments', { params });
  return unwrapPagination<PaymentRecord>(res);
}

export async function listRefundsRecords(params: FinanceRecordQuery) {
  const res = await api.get('/api/v1/admin/refunds', { params });
  return unwrapPagination<RefundRecord>(res);
}

export async function exportPaymentsRecords(params: FinanceRecordQuery & { format?: 'csv' | 'xlsx' }) {
  const res = await api.get('/api/v1/admin/payments/export', {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}

export async function exportRefundsRecords(params: FinanceRecordQuery & { format?: 'csv' | 'xlsx' }) {
  const res = await api.get('/api/v1/admin/refunds/export', {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}
