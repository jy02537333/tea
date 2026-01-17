import { api, unwrapPagination } from './api';

export interface StoreRefundOrder {
  id: number;
  order_no: string;
  store_id: number;
}

export interface StoreRefundRecord {
  id: number;
  order_id: number;
  payment_id: number;
  refund_no: string;
  refund_amount: string;
  refund_reason?: string;
  status: number;
  refunded_at?: string;
  created_at?: string;
  updated_at?: string;
  order?: StoreRefundOrder;
}

export interface StoreRefundQuery {
  page?: number;
  limit?: number;
  order_id?: number;
  refund_no?: string;
  status?: number;
  start?: string;
  end?: string;
}

export async function listStoreRefundsRecords(storeId: number, params: StoreRefundQuery) {
  const res = await api.get(`/api/v1/stores/${storeId}/refunds`, { params });
  return unwrapPagination<StoreRefundRecord>(res);
}
