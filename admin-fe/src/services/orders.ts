
import api, { unwrapResponse } from './api';
import { OrderSummary, Order, PaginatedResponse } from './types';

export async function createOrderFromCart(payload: { delivery_type: number; address_info?: string; remark?: string; user_coupon_id?: number; store_id?: number; order_type?: number }): Promise<OrderSummary> {
  const res = await api.post('/api/v1/orders/from-cart', payload);
  return unwrapResponse<OrderSummary>(res);
}

export async function listOrders(params: { page?: number; limit?: number; status?: number; store_id?: number }): Promise<PaginatedResponse<OrderSummary>> {
  const res = await api.get('/api/v1/orders', { params });
  return unwrapResponse<PaginatedResponse<OrderSummary>>(res);
}

export async function getOrder(id: number): Promise<Order> {
  const res = await api.get(`/api/v1/orders/${id}`);
  return unwrapResponse<Order>(res);
}

export async function cancelOrder(id: number, reason?: string): Promise<void> {
  const res = await api.post(`/api/v1/orders/${id}/cancel`, { reason });
  return unwrapResponse<void>(res);
}

export async function payOrder(id: number): Promise<void> {
  const res = await api.post(`/api/v1/orders/${id}/pay`);
  return unwrapResponse<void>(res);
}

export async function receiveOrder(id: number): Promise<void> {
  const res = await api.post(`/api/v1/orders/${id}/receive`);
  return unwrapResponse<void>(res);
}

// Admin specific
export async function adminListOrders(params: { page?: number; limit?: number; status?: number; store_id?: number }): Promise<PaginatedResponse<OrderSummary>> {
  const res = await api.get('/api/v1/admin/orders', { params });
  return unwrapResponse<PaginatedResponse<OrderSummary>>(res);
}

export async function adminExportOrders(params: { status?: number; store_id?: number }) {
  const res = await api.get('/api/v1/admin/orders/export', { params, responseType: 'blob' });
  return res.data; // blob
}

export async function adminGetOrder(id: number): Promise<Order> {
  const res = await api.get(`/api/v1/admin/orders/${id}`);
  return unwrapResponse<Order>(res);
}
