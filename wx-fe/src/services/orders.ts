import api, { unwrapResponse } from './api';
import { Order, PaginationResponse } from './types';

export async function createOrderFromCart(payload: { delivery_type: number; address_info?: string; remark?: string; user_coupon_id?: number; store_id?: number; order_type?: number }): Promise<Order> {
  const res = await api.post('/api/v1/orders/from-cart', payload);
  return unwrapResponse<Order>(res);
}

export async function listOrders(params: { page?: number; limit?: number; status?: number; store_id?: number } = {}): Promise<PaginationResponse<Order>> {
  const res = await api.get('/api/v1/orders', { params });
  return unwrapResponse<PaginationResponse<Order>>(res);
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
  const res = await api.post(`/api/v1/orders/${id}/pay`, {});
  return unwrapResponse<void>(res);
}

export async function receiveOrder(id: number): Promise<void> {
  const res = await api.post(`/api/v1/orders/${id}/receive`, {});
  return unwrapResponse<void>(res);
}
