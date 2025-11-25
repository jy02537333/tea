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
