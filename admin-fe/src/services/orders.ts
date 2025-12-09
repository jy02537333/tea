import { api, PaginatedResult, unwrap } from './api';

export interface AdminOrder {
  id: number;
  order_no: string;
  store_id: number;
  user_id: number;
  pay_amount: number;
  status: number;
  pay_status: number;
  created_at?: string;
  updated_at?: string;
}

export interface AdminOrderItem {
  id: number;
  product_id: number;
  product_name: string;
  sku_name?: string;
  quantity: number;
  price: number | string;
  amount: number | string;
}

export interface AdminOrderDetail {
  order: AdminOrder;
  items: AdminOrderItem[];
}

export interface AdminOrderListParams {
  page?: number;
  limit?: number;
  store_id?: number;
  status?: number;
}

export async function getAdminOrders(params: AdminOrderListParams): Promise<PaginatedResult<AdminOrder>> {
  const res = await api.get('/api/v1/admin/orders', { params });
  const payload = res.data ?? {};
  const list: AdminOrder[] = payload.data ?? [];
  const limit = payload.limit ?? payload.size ?? params.limit ?? (list.length || 20);
  return {
    list,
    total: payload.total ?? list.length,
    page: payload.page ?? params.page ?? 1,
    limit,
  };
}

export async function getAdminOrderDetail(id: number) {
  const res = await api.get(`/api/v1/admin/orders/${id}`);
  return unwrap<AdminOrderDetail>(res);
}

export async function exportAdminOrders(params: AdminOrderListParams = {}) {
  const res = await api.get('/api/v1/admin/orders/export', {
    params,
    responseType: 'blob',
  });
  return res.data as Blob;
}

export async function postOrderAction(id: number, action: string, body?: Record<string, any>) {
  const res = await api.post(`/api/v1/orders/${id}/${action}`, body ?? {});
  return unwrap(res);
}
